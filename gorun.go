package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	l "log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rjeczalik/notify"
)

const (
	DIR_PERMISSIONS = 0700

	EXAMPLES = `
  gorun .
  gorun ./dir
  gorun ./dir arg0 arg1 arg2 ...

  gorun -w               ./dir
  gorun -n=process-name  ./dir
  gorun -p=./dir,./dir0  ./dir
  gorun -v -w -n=name    ./dir
`
)

var (
	OUT     = flag.CommandLine.Output()
	VERBOSE = l.New(ioutil.Discard, "[gorun] ", 0)
	TEMPDIR = filepath.Join(os.TempDir(), "gorun-"+strconv.FormatInt(int64(os.Getpid()), 10))
)

func printUsage() {
	fmt.Fprintf(OUT, "\n%s\nOptions:\n\n", EXAMPLES)
	flag.PrintDefaults()
}

func main() {
	watch := flag.Bool("w", false, "Watch and rerun")
	name := flag.String("n", "", "Process name")
	verbose := flag.Bool("v", false, "Verbose logging")
	pattern := flag.String("p", "", "Comma-separated watch patterns. Implies -w. Use ... for a wildcard.")

	// For implicit "-h"
	flag.Usage = func() {
		fmt.Fprintf(OUT, `Usage of %s:`, os.Args[0])
		printUsage()
	}

	flag.Parse()

	if *verbose {
		VERBOSE.SetOutput(OUT)
	}

	if len(flag.Args()) == 0 {
		fmt.Fprintf(OUT, `Please specify a file or directory to run. Usage:`)
		printUsage()
		os.Exit(1)
	}

	err := os.MkdirAll(TEMPDIR, DIR_PERMISSIONS)
	if err != nil {
		fatal(err)
	}

	patterns := splitPattern(*pattern)
	if len(patterns) > 0 {
		*watch = true
	}

	target, args := flag.Args()[0], flag.Args()[1:]

	// Single run
	if !*watch {
		err := buildAndRun(context.Background(), *name, target, args)
		if err != nil {
			fatal(err)
		}
		return
	}

	// Watch

	if len(patterns) == 0 {
		patterns = []string{filepath.Join(target, "...")}
	}

	events := make(chan notify.EventInfo, 1)
	for _, pattern := range patterns {
		VERBOSE.Printf("Watching pattern %v", pattern)
		err = notify.Watch(pattern, events, notify.All)
		if err != nil {
			fatal(err)
		}
	}

	for {
		ctx, cancel := context.WithCancel(context.Background())
		done := gogo(func() error {
			return buildAndRun(ctx, *name, target, args)
		})
		t0 := time.Now()

		select {
		case <-events:
			VERBOSE.Printf("Stopping")
			cancel()

		case err := <-done:
			t1 := time.Now()
			delta := t1.Sub(t0)

			if err != nil {
				VERBOSE.Printf("Finished in %v (error)", delta)
				logErr(err)
			} else {
				VERBOSE.Printf("Finished in %v", delta)
			}

			<-events
		}
	}
}

func buildAndRun(ctx context.Context, name string, target string, args []string) error {
	if name == "" {
		var err error
		name, err = chooseBinName(target)
		if err != nil {
			return err
		}
	}

	binpath := filepath.Join(TEMPDIR, name)
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binpath, target)
	pipeIo(cmd)

	VERBOSE.Println("Building")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, binpath, args...)
	pipeIo(cmd)

	VERBOSE.Println("Running")
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

// Chooses the name for the binary to build. This determines the name of the
// child process. Can be set manually with the -n flag.
func chooseBinName(target string) (string, error) {
	if target == "." {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Base(wd), nil
	}
	return filepath.Base(target), nil
}

func splitPattern(pattern string) (out []string) {
	for _, str := range strings.Split(pattern, ",") {
		str = strings.TrimSpace(str)
		if str != "" {
			out = append(out, str)
		}
	}
	return
}

func fatal(err error) {
	logErr(err)
	os.Exit(1)
}

func logErr(err error) {
	// In exec.ExitError, the error message is just a prefix+suffix of the
	// subprocess's stderr, which we already redirect to our stderr. So
	// there's no point logging it.
	exitErr, _ := err.(*exec.ExitError)
	if exitErr != nil {
		return
	}
	fmt.Fprintln(OUT, err)
}
