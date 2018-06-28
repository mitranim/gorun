package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	l "log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/rjeczalik/notify"
)

const (
	DIR_MODE = 0700

	DESCRIPTION = `"gorun" runs a Go directory. Can watch and rerun.`

	EXAMPLES = `
  gorun .
  gorun ./dir
  gorun ./dir arg0 arg1 arg2 ...

  gorun -w               ./dir
  gorun -p=./dir,./dir0  ./dir
  gorun -n=name          ./dir
  gorun -v -w -n=name    ./dir
`
)

var (
	WATCH   = flag.Bool("w", false, "Watch and rerun")
	PATTERN = flag.String("p", "", `Comma-separated watch patterns. Implies "-w". Uses "..." as a wildcard.`)
	NAME    = flag.String("n", "", `Binary and process name. Works only when using "go build".`)
	VERBOSE = flag.Bool("v", false, "Verbose logging")

	log  = l.New(flag.CommandLine.Output(), "", 0)
	verb = l.New(ioutil.Discard, "[gorun] ", 0)

	TEMPDIR string
)

func main() {
	printUsage := func() {
		log.Print(EXAMPLES, "\nOptions:\n\n")
		flag.PrintDefaults()
	}

	// For implicit "-h"
	flag.Usage = func() {
		log.Println(DESCRIPTION, "Usage:")
		printUsage()
	}

	// In addition to parsing flags, if called with "-h", this will print help
	// and exit the process.
	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Printf(`Please specify a file or directory to run. Usage:`)
		printUsage()
		os.Exit(1)
	}

	if *VERBOSE {
		verb.SetOutput(flag.CommandLine.Output())
	}

	patterns := stringSplit(*PATTERN, ",")
	if len(patterns) > 0 {
		*WATCH = true
	}

	initTempAndCleanup()

	target, args := flag.Args()[0], flag.Args()[1:]

	var err error
	if !*WATCH {
		err = runOnce(target, args)
	} else {
		err = watchAndRerun(target, args, patterns)
	}

	if err != nil {
		logErr(err)
		os.Exit(1)
	}
}

func runOnce(target string, args []string) error {
	return runTarget(context.Background(), target, args)
}

func watchAndRerun(target string, args, patterns []string) error {
	if len(patterns) == 0 {
		patterns = []string{filepath.Join(target, "...")}
	}

	events := make(chan notify.EventInfo, 1)
	for _, pattern := range patterns {
		verb.Printf("Watching pattern %v", pattern)
		err := notify.Watch(pattern, events, notify.All)
		if err != nil {
			return err
		}
	}

rerunLoop:
	for {
		ctx, cancel := context.WithCancel(context.Background())
		done := gogo(func() error {
			return runTarget(ctx, target, args)
		})
		t0 := time.Now()

	selectLoop:
		for {
			select {
			case err := <-done:
				t1 := time.Now()
				delta := t1.Sub(t0)

				if err != nil {
					verb.Printf("Finished in %v (error)", delta)
					logErr(err)
				} else {
					verb.Printf("Finished in %v", delta)
				}

				break selectLoop

			case event := <-events:
				if !isRelevantPath(event.Path()) {
					verb.Printf("Ignoring %v", event)
					continue selectLoop
				}
				verb.Printf("Stopping due to %v", event)
				cancel()
				continue rerunLoop
			}
		}

		for event := range events {
			if !isRelevantPath(event.Path()) {
				verb.Printf("Ignoring %v", event)
				continue
			}
			verb.Printf("Stopping due to %v", event)
			cancel()
			continue rerunLoop
		}
	}
}

func runTarget(ctx context.Context, target string, args []string) error {
	abs, err := filepath.Abs(target)
	if err != nil {
		return err
	}

	ext := filepath.Ext(abs)
	isDir := ext == ""
	binName := strings.TrimSuffix(filepath.Base(abs), ext)
	installable := false

	if isDir {
		gopath := os.Getenv("GOPATH")
		if gopath != "" {
			installable = isWithinPath(filepath.Join(gopath, "src"), abs)
		}
	}

	if installable {
		if *NAME != "" {
			return fmt.Errorf(`Option "-n" doesn't work with "go install"`)
		}

		t0 := time.Now()
		cmd := exec.CommandContext(ctx, "go", "install", target)
		pipeIo(cmd)
		verb.Println("Installing")
		err := cmd.Run()
		if err != nil {
			return err
		}

		t1 := time.Now()
		cmd = exec.CommandContext(ctx, binName, args...)
		pipeIo(cmd)
		verb.Printf("Installed in %v, running", t1.Sub(t0))
		return cmd.Run()
	}

	if *NAME != "" {
		binName = *NAME
	}

	t0 := time.Now()
	binPath := filepath.Join(TEMPDIR, binName)
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, target)
	pipeIo(cmd)
	verb.Println("Building")
	err = cmd.Run()
	if err != nil {
		return err
	}

	t1 := time.Now()
	cmd = exec.CommandContext(ctx, binPath, args...)
	pipeIo(cmd)
	verb.Printf("Built in %v, running", t1.Sub(t0))
	return cmd.Run()
}

func gogo(fun func() error) chan error {
	out := make(chan error, 1)
	go func() {
		err := fun()
		if err != nil {
			out <- err
		}
		close(out)
	}()
	return out
}

func pipeIo(cmd *exec.Cmd) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}

// `strings.Split` without empty strings.
func stringSplit(input string, separator string) (out []string) {
	for _, str := range strings.Split(input, separator) {
		str = strings.TrimSpace(str)
		if str != "" {
			out = append(out, str)
		}
	}
	return
}

func isWithinPath(ancestor string, descendant string) bool {
	if len(ancestor) > len(descendant) {
		return false
	}
	var i int
	for ; i < len(ancestor); i++ {
		if ancestor[i] != descendant[i] {
			return false
		}
	}
	return len(ancestor) == len(descendant) || descendant[i] == os.PathSeparator
}

func initTempAndCleanup() {
	TEMPDIR = filepath.Join(os.TempDir(), fmt.Sprintf("gorun-%v", os.Getpid()))

	sigs := make(chan os.Signal, 1)
	// Might fail on Windows
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		// sig := <-sigs
		<-sigs
		signal.Stop(sigs)

		// verb.Printf("Received %q, deleting %v", sig, TEMPDIR)
		err := os.RemoveAll(TEMPDIR)
		if err != nil {
			verb.Printf("Failed to delete %v: %v", TEMPDIR, err)
		}

		// Is there a "proper" exit code for this?
		os.Exit(1)
	}()
}

func logErr(err error) {
	// In exec.ExitError, the error message is just a prefix+suffix of the
	// subprocess's stderr, which we already redirect to our stderr. So
	// there's no point logging it.
	exitErr, _ := err.(*exec.ExitError)
	if exitErr != nil {
		return
	}
	log.Print(err)
}

func isRelevantPath(path string) bool {
	return filepath.Ext(path) == ".go"
}
