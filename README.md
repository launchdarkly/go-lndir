# go-lndir: A port of lndir to Go

[![Build Status](https://travis-ci.org/launchdarkly/go-lndir.svg?branch=master)](https://travis-ci.org/launchdarkly/go-lndir)

From the linux [man page](http://www.xfree86.org/4.3.0/lndir.1.html):

> The lndir program makes a shadow copy todir of a directory tree fromdir, except that the shadow is not populated with real files but instead with symbolic links pointing at the real files in the fromdir directory tree.

**Current stable library release:** gopkg.in/launchdarkly/go-lndir.v1

This project was originally derived from the C language source at [lndir.c](https://opensource.apple.com/source/X11misc/X11misc-10.1/lndir/lndir-1.0.1/lndir.c). 

`go-lndir` also introduces a `-gitignore` option that causes it to skip files and directories specified in .gitignore.

## Why?

The impetus to port this to Go was to make it available on OSX and to add support for ignoring files specified in `.gitignore`.  It is used by `github.com/launchdarkly/gogitix` to quickly clone a git workspace for in order to run pre-commit tests in a clean workspace.

## Installation

To install the command-line tool `go-lndir`, run:

```
go get -u github.com/launchdarkly/go-lndir/cmd/...
```

Then run it:

```
go-lndir <path to source directory from target directory> [target directory]
```

If the path you provide for the source directory is relative, then all of the generated links will also be relative.  

## Testing

1. Install `bats`.  On OSX, this can be done with ```brew install bats```.
2. Run `make`.

## Planned Improvements

Remaining planned changes:

1. Switching to the go logger instead of the output related code that was ported from C.

I don't have a sense of how important output and error code compatibility might be with existing `lndir` deploys.  I haven't broken it yet but if you need these things, please let me know.


  
