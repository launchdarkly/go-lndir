# go-lndir: A port of lndir to golang

This is a port of https://opensource.apple.com/source/X11misc/X11misc-10.1/lndir/lndir-1.0.1/lndir.c.  It creates a copy of the directory struture that contains links to the files in the original directory structure.

It's not well tested so *use at your own peril*.  If lndir has a test suite, I couldn't find it and this began as a verbatim port of C-code.

To use it, run:

```
go get -u github.com/launchdarkly/go-lndir
go install github.com/launchdarkly/go-lndir
```

Then run it:

```
go-lndir <path to source directory from target directory> [target directory]
```

If the path you provide for the source directory is relative, then all of the generated links will also be relative.  

## Testing

Install `bats`.  On OSX, this can be done with ```brew install bats```.

## Plans

Remaining planned changes:

1. Switching to the go logger instead of the output related code that was ported from C.

I don't have a sense of how important output and error code compatibility might be with existing lndir deploys.  I haven't broken it yet but if you need these things, please let me know.


  
