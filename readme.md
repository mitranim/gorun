## Overview

Minimal CLI tool for Go development. Can:

  * run an entire Go package with a single command, without `go build`
  * optionally watch and rerun on file changes

Works on MacOS, should work on Linux. Pull requests for Windows are welcome.

## Why

  * one command to build and run
  * easy to remember
  * can watch and rerun

### Why not existing tools

Existing tools, like [`realize`](https://github.com/oxequa/realize), tend to:

  * have an **earth-shattering** amounts of bells and whistles
  * have verbose logging you can't disable
  * have unnecessary delays in the file watcher
  * use CPU constantly
  * require config files
  * put garbage in the working directory
  * mess with file paths in error reports
  * be large and complex, so you can't fix them yourself
  * have 1000 open issues, causing an unresponsive maintainer

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

### 2018-06-15

Now uses `go install` when possible, falling back on `go build`.

When `gorun` uses `go build` and is stopped with `^C` or by closing a terminal tab, it immediately deletes its temporary directory with the binary.

After updating `gorun`, delete any leftover directories:

    find $TMPDIR -name "gorun-*" -delete

Verbose log now includes build duration.

## TODO

Consider stopping the child process with `SIGINT` to allow cleanup. Must have a timeout.

## Misc

Proposals to add directory support to `go run` have been rejected multiple times:

  * https://github.com/golang/go/issues/5164
  * https://github.com/golang/go/issues/20432

Seems like one has finally been accepted:

  * https://github.com/golang/go/issues/22726

I'm receptive to suggestions. If this tool _almost_ satisfies you but needs changes, open an issue or chat me up. Contacts: https://mitranim.com/#contacts
