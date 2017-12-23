package main

/*

Derived from lndir.c, which includes the above copyright notice.

Copyright (c) 1990, 1998 The Open Group

Permission to use, copy, modify, distribute, and sell this software and its
documentation for any purpose is hereby granted without fee, provided that
the above copyright notice appear in all copies and that both that
copyright notice and this permission notice appear in supporting
documentation.

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
OPEN GROUP BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

Except as contained in this notice, the name of The Open Group shall not be
used in advertising or otherwise to promote the sale, use or other dealings
in this Software without prior written authorization from The Open Group.

*/

import (
	"flag"
	"fmt"
	"os"

	lndir "github.com/launchdarkly/go-lndir"
)

func main() {
	config := lndir.Config{}

	flag.BoolVar(&config.Silent, "silent", false, "suppress output")
	flag.BoolVar(&config.IgnoreLinks, "ignorelinks", false, "Don't give links special treatment")
	flag.BoolVar(&config.WithRevInfo, "withrevinfo", false, "Include revision directories (.git, etc)")
	flag.BoolVar(&config.UseGitignore, "gitignore", false, "Exclude files listed in ,gitignore files")

	flag.Parse()

	if flag.NArg() < 1 && flag.NArg() > 2 {
		flag.Usage()
		os.Exit(1)
	}

	fromPath := flag.Arg(0)
	toPath := flag.Arg(1)
	if toPath == "" {
		toPath = "."
	}

	err := lndir.Lndir(fromPath, toPath, config)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if lndir.IsUserError(err) {
			os.Exit(2)
		} else {
			os.Exit(1)
		}
	}
}
