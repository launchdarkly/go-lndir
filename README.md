# go-lndir: A port of lndir to golang

This is a port of https://opensource.apple.com/source/X11misc/X11misc-10.1/lndir/lndir-1.0.1/lndir.c.

It's not well-tested so use at your own peril.  If lndir has a test suite, I didn't see it and this began as a verbatim port of C-code.

## Plans

There are a couple of planned changes:

1. Gitignore support
2. Switching to the go logger instead of the output related code that was ported from C.

I don't have a sense of how important output and error code compatibility might be with existing lndir deploys.  I haven't broken it yet but if you need these things, please let me know.


  
