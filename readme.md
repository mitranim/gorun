## Overview

`gorun` runs an entire Go package without an intermediary `go build`. It's a version of `go run` that works for directories.

## Installation

```sh
go get github.com/Mitranim/gorun
```

This will automatically get the code and compile the executable. Make sure your `$GOPATH/bin` is in your `$PATH` so the shell can discover it. For example, my `~/.profile` contains this:

```sh
export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin
```

## Usage

```sh
gorun .

gorun some-directory

# any additional arguments are passed to the program
gorun . arg0 arg1 arg2 ...

# usage info
gorun --help
```

It calls `go run`, listing the `.go` files from the given directory, skipping any `_test.go` files.

Proposals to add this functionality to `go run` keep being rejected:

  * https://github.com/golang/go/issues/5164
  * https://github.com/golang/go/issues/20432

Note: the current version only works on Unix. Modifications are welcome.

## Misc

I'm receptive to suggestions. If this package _almost_ satisfies you but needs changes, open an issue or chat me up. Contacts: https://mitranim.com/#contacts
