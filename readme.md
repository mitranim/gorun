## Overview

Minimal CLI tool for Go development. Can:

  * run an entire Go package with a single command, without `go build`
  * optionally watch and rerun on file changes

Works on MacOS, should work on Linux. Pull requests for Windows are welcome.

## Why

The standard `go run` command works for single files, but not directories. Using `go build` in development quickly gets annoying. With this tool, you just `gorun .`.

When developing a long-running program, like a server, you typically want to rerun on code changes. Other tools exist, like [`realize`](https://github.com/oxequa/realize), but they have an **earth-shattering** amounts of bells and whistles and unwanted features. I just want to watch and rerun! Silently, too!

Differences from `realize`:

  * small and simple
  * no extraneous logging
  * no config files
  * no garbage in working directory
  * no background CPU usage, or very little of it
  * doesn't mess with file paths in error reports

## Installation

```sh
go get -u github.com/Mitranim/gorun
```

This will automatically get the code and compile the executable. Make sure your `$GOPATH/bin` is in your `$PATH` so the shell can discover it. For example, my `~/.profile` contains this:

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

## Misc

Proposals to add directory support to `go run` have been rejected multiple times:

  * https://github.com/golang/go/issues/5164
  * https://github.com/golang/go/issues/20432

Seems like one has finally been accepted:

  * https://github.com/golang/go/issues/22726

I'm receptive to suggestions. If this tool _almost_ satisfies you but needs changes, open an issue or chat me up. Contacts: https://mitranim.com/#contacts
