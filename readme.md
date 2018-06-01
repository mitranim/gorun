## Overview

Minimal CLI tool for Go development. Can:

  * run an entire Go package with a single command, without `go build`

  * optionally watch and rerun on file changes

Works on MacOS, should work on Linux. Pull requests for Windows are welcome.

## Why

The standard `go run` command works for single files, but not directories. Using `go build` in development quickly gets annoying. With this tool, you just `gorun .`.

When developing a long-running program, like a server, you typically want it to rerun on code changes. Other similar tools, like [`realize`](https://github.com/oxequa/realize), come with an **earth-shattering** amount of bells and whistles and unwanted features. I just want to watch and rerun! Silently, too!

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
gorun -n my-app .

# Watch and rerun
gorun -w .

# Any additional arguments are passed to the program
gorun . arg0 arg1 arg2 ...

# Usage info
gorun -h
```

Strangely, proposals to add directory support to `go run` keep being rejected:

  * https://github.com/golang/go/issues/5164
  * https://github.com/golang/go/issues/20432

## Misc

I'm receptive to suggestions. If this tool _almost_ satisfies you but needs changes, open an issue or chat me up. Contacts: https://mitranim.com/#contacts
