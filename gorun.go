package main

import (
	"io/ioutil"
	l "log"
	"os"
	"os/exec"
	"path"
	"strings"
)

const (
	EXECUTABLE = "go"

	RUN = "run"

	USAGE = `

    gorun .
    gorun <directory>
    gorun <directory> <... args>
    gorun --help`

	HELP = `gorun is a simple way to run a Go directory. Usage:` + USAGE

	MISSING_ARG = `Please specify a file or directory to run. Examples:` + USAGE
)

var log = l.New(os.Stderr, "", 0)

func main() {
	if !(len(os.Args) >= 2) {
		log.Fatal(MISSING_ARG)
	}
	target := os.Args[1]

	if target == "--help" || target == "-h" {
		log.Fatal(HELP)
	}

	stat, err := os.Stat(target)
	if err != nil {
		log.Fatal(err)
	}

	if stat.IsDir() {
		runDir(target, os.Args[2:])
	} else {
		runFile(target, os.Args[2:])
	}
}

func runDir(target string, args []string) {
	stats, err := ioutil.ReadDir(target)
	if err != nil {
		log.Fatal(err)
	}

	paths := []string{}
	for _, stat := range stats {
		pt := path.Join(target, stat.Name())
		if path.Ext(pt) != ".go" ||
			strings.HasSuffix(pt, "_test.go") {
			continue
		}
		if !stat.IsDir() {
			paths = append(paths, pt)
		}
	}
	if len(paths) == 0 {
		log.Fatalf("No .go files found in directory %s", target)
	}

	cmdArgs := []string{RUN}
	for _, path := range paths {
		cmdArgs = append(cmdArgs, path)
	}
	for _, arg := range args {
		cmdArgs = append(cmdArgs, arg)
	}

	runCmd(EXECUTABLE, cmdArgs)
}

func runFile(target string, args []string) {
	cmdArgs := []string{RUN, target}
	for _, arg := range args {
		cmdArgs = append(cmdArgs, arg)
	}
	runCmd(EXECUTABLE, cmdArgs)
}

func runCmd(executable string, args []string) {
	log.Printf("# Command: %s %s", executable, strings.Join(args, " "))

	cmd := exec.Command(executable, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		// In exec.ExitError, the error message is just a prefix+suffix of the
		// subprocess's stderr, which we already redirect to our stderr. So
		// there's no point logging it. Would be nice to reuse the subprocess's
		// exit code, but I found no way to get it.
		exitErr, _ := err.(*exec.ExitError)
		if exitErr != nil {
			os.Exit(1)
		}

		log.Fatal(err)
	}
}
