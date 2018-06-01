package main

import (
	"context"
	"flag"
	l "log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/rjeczalik/notify"
)

const DIR_PERMISSIONS = 0700

var (
	log     = l.New(os.Stderr, "[gorun] ", 0)
	TEMPDIR = filepath.Join(os.TempDir(), "gorun-"+strconv.FormatInt(int64(os.Getpid()), 10))
)

func main() {
	watch := flag.Bool("w", false, "Watch and rerun")
	name := flag.String("n", "", "Process name")
	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Println("Please specify a file or directory to run. Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	err := os.MkdirAll(TEMPDIR, DIR_PERMISSIONS)
	if err != nil {
		fatal(err)
	}

	target, args := flag.Args()[0], flag.Args()[1:]

	if !*watch {
		err := buildAndRun(context.Background(), *name, target, args)
		if err != nil {
			fatal(err)
		}
		return
	}

	events := make(chan notify.EventInfo, 1)
	// `dir/...` works as a glob pattern
	err = notify.Watch(filepath.Join(target, "..."), events, notify.All)
	if err != nil {
		fatal(err)
	}

	for {
		ctx, cancel := context.WithCancel(context.Background())
		done := gogo(func() error {
			return buildAndRun(ctx, *name, target, args)
		})

		select {
		case <-events:
			cancel()
		case err := <-done:
			if err != nil {
				logErr(err)
			}
			<-events
		}
	}
}

func buildAndRun(ctx context.Context, name string, target string, args []string) error {
	if name == "" {
		var err error
		name, err = binName(target)
		if err != nil {
			return err
		}
	}

	binpath := filepath.Join(TEMPDIR, name)
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binpath, target)
	pipeIo(cmd)
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, binpath, args...)
	pipeIo(cmd)
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

func binName(target string) (string, error) {
	if target == "." {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Base(wd), nil
	}
	return filepath.Base(target), nil
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
	log.Println(err)
}
