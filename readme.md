## Overview

Minimal CLI tool for Go development. Can:

  * run an entire Go package with a single command, without `go build`
  * optionally watch and rerun on file changes

Works on MacOS, should work on Linux. Pull requests for Windows are welcome.

## Why

  * one command to build and run
  * easy to remember
  * can watch and rerun

The Go team has rejected multiple proposals to add directory support to `go run`: [[1]](https://github.com/golang/go/issues/5164)[[2]](https://github.com/golang/go/issues/20432). One has finally been accepted: [[3]](https://github.com/golang/go/issues/22726), but is not part of a stable Go release at the time of writing.

### Why not existing runners

Existing runners, like [`realize`](https://github.com/oxequa/realize), tend to:

  * have an **earth-shattering** amount of bells and whistles
  * have verbose logging you can't disable
  * have unnecessary delays in the file watcher
  * use CPU constantly
  * require config files
  * put garbage in the working directory
  * mess with file paths in error reports
  * have thousands of lines of code, so you can't fix them yourself

`gorun` doesn't.

## Installation

```sh
go get -u github.com/Mitranim/gorun
```

This will automatically compile the executable. Make sure `$GOPATH/bin` is in your `$PATH` so the shell can discover it. For example, my `~/.profile` contains this:

```sh
export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin
```

## Usage

```sh
# Run current directory
gorun .

# Run subdirectory
gorun ./src/go

# Specify process name
gorun -n=my-app .

# Watch and rerun
gorun -w .

# Any additional arguments are passed to the program
gorun . arg0 arg1 arg2 ...

# Usage info
gorun -h
```

## Changelog

### 2018-06-28

Now ignores non-`.go` files when watching, regardless of the watch pattern.

### 2018-06-15

Now uses `go install` when possible, falling back on `go build`.

When `gorun` uses `go build` and is stopped with `^C` or by closing a terminal tab, it immediately deletes its temporary directory with the binary.

After updating `gorun`, delete any leftover directories:

    find $TMPDIR -name "gorun-*" -delete

Verbose log now includes build duration.

## TODO

Consider stopping the child process with `SIGINT` to allow cleanup; don't forget a timeout.

## Alternatives

For general purpose file watching, consider these excellent tools:

  * https://github.com/emcrisostomo/fswatch
  * https://github.com/mattgreen/watchexec

Differences:

  * `gorun` builds and runs a Go directory, which can be fiddly and awkward otherwise.
  * Most general-purpose watchers don't support killing and restarting the child process; `watchexec` is one of the few exceptions.
  * For Go, remembering how to invoke `gorun` is much easier.

## License

https://en.wikipedia.org/wiki/WTFPL

## Misc

I'm receptive to suggestions. If this tool _almost_ satisfies you but needs changes, open an issue or chat me up. Contacts: https://mitranim.com/#contacts
