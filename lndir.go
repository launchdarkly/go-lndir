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
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

var (
	silent      = false
	ignoreLinks = false
	withRevInfo = false
)

var isOSX = runtime.GOOS == "darwin"

var (
	rcurdir path
	curdir  path
)

type path []string

func (p path) isAbs() bool {
	return p[0] == "/"
}

func (p path) String() string {
	return filepath.Join(p...)
}

func NewPath(pathStr string) path {
	if pathStr == "" {
		quit(1, "Bad path: %s", pathStr)
	}
	segments := strings.Split(pathStr, string(filepath.Separator))
	if filepath.IsAbs(pathStr) {
		return append([]string{"/"}, segments...)
	} else {
		return segments
	}
}

func quit(code int, fmtStr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
	os.Stderr.Write([]byte("\n"))
	os.Exit(code)
}

func quiterr(code int, msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(code)
}

func perror(msg string, err error) {
	if msg != "" {
		fmt.Fprint(os.Stderr, msg+":")
	}
	fmt.Fprintln(os.Stderr, err)
}

func msg(fmtStr string, args ...interface{}) {
	if curdir != nil {
		fmt.Fprintf(os.Stderr, "%s:\n", curdir)
		curdir = nil
	}
	fmt.Fprintf(os.Stderr, fmtStr, args...)
	os.Stderr.Write([]byte("\n"))
}

func mperror(msg string, err error) {
	if curdir != nil {
		fmt.Fprintf(os.Stderr, "%s:\n", curdir)
		curdir = nil
	}
	perror(msg, err)
}

func equivalent(lname path, rname path) bool {
	return filepath.Clean(lname.String()) == filepath.Clean(rname.String())
}

func processSubdir(subdirName string, parentPath path, subdirInfo os.FileInfo) {
	if subdirName == "." || subdirName == ".." {
		return
	}

	if !withRevInfo && isRevInfo(subdirName) {
		return
	}

	// These are maintained for printing the path (once and only once) before an error
	ocurdir := rcurdir
	rcurdir = parentPath
	if silent {
		curdir = parentPath
	} else {
		fmt.Printf("%s:\n", parentPath)
	}

	// Restore these when the method is done
	defer func() {
		rcurdir = ocurdir
		curdir = ocurdir
	}()

	var err error
	var targetInfo os.FileInfo
	if targetInfo, err = os.Stat(subdirName); err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(subdirName, os.FileMode(0777)); err == nil {
				if targetInfo, err = os.Stat(subdirName); err != nil {
					mperror(subdirName, err)
					return
				}
			}
		} else {
			mperror(subdirName, err)
			return
		}
	}

	_, err = os.Readlink(subdirName)
	if err == nil {
		msg("%s: is a link instead of a directory", subdirName)
		return
	}

	if err = os.Chdir(subdirName); err != nil {
		mperror(subdirName, err)
		return
	}
	defer func() {
		if err = os.Chdir(".."); err != nil {
			quiterr(1, "..")
		}
	}()

	srcPath := parentPath
	if !srcPath.isAbs() {
		srcPath = append(NewPath(".."), parentPath...)
	}

	processDirectory(srcPath, subdirInfo, targetInfo)
}

func processDirectory(sourceDirPath path, sourceDir os.FileInfo, targetDir os.FileInfo) int {
	if os.SameFile(sourceDir, targetDir) {
		msg("%s: From and to directories are identical!", sourceDirPath)
		return 1
	}

	var err error
	var f *os.File
	if f, err = os.Open(sourceDirPath.String()); err != nil {
		msg("%s: Cannot opendir: %s", sourceDirPath.String(), err)
		return 1
	}

	// Determine the maximum number of directories we might see
	dirsLeft := math.MaxInt32
	if s, err := f.Stat(); err != nil {
		msg("%s: Cannot stat", err)
		return 1
	} else if stat, ok := s.Sys().(*syscall.Stat_t); ok && stat != nil {
		// Apparently, if this is 1, we have no clue about how many subdirectories there are in this directory
		if stat.Nlink != 1 {
			dirsLeft = int(stat.Nlink)
		}
	}

	var children []string
	if children, err = f.Readdirnames(0); err != nil {
		msg("%s: Cannot readdir", err)
		return 1
	}
	f.Close()

	for _, name := range children {
		if strings.HasSuffix(name, "~") {
			continue
		}

		if isOSX && (name == ".DS_Store" || name == "._DS_Store") {
			continue
		}

		sourcePath := append(sourceDirPath, name)

		// Optimization to skip these checks once all directory entries have been processed
		if dirsLeft > 0 {
			var childInfo os.FileInfo
			if childInfo, err = os.Lstat(sourcePath.String()); err != nil {
				mperror(sourcePath.String(), err)
				continue
			}

			if childInfo.IsDir() {
				processSubdir(name, sourcePath, childInfo)
				dirsLeft -= 1
				continue
			}
		}

		// The option to ignore links exists mostly because
		//   checking for them slows us down by 10-20%.
		//   But it is off by default because this really is a useful check.
		var sourceSymlinkPath path
		if !ignoreLinks {
			// see if the file in the base tree was a symlink
			sourceSymlinkPath = readlink(sourcePath)
		}

		existingSymlinkPath := readlink(NewPath(name))
		if existingSymlinkPath != nil {
			// Link exists in new tree.  Print message if it doesn't match.
			expectedSymlinkPath := sourcePath
			if sourceSymlinkPath != nil {
				expectedSymlinkPath = sourceSymlinkPath
			}
			if !equivalent(existingSymlinkPath, expectedSymlinkPath) {
				msg("%s: %s", name, existingSymlinkPath)
			}
		} else {
			newSymlinkPath := sourcePath
			if sourceSymlinkPath != nil {
				newSymlinkPath = sourceSymlinkPath

				if sourcePath[0] == ".." && sourceSymlinkPath[0] == ".." {
					//	It becomes very tricky here. We have
					//	  ../../bar/foo symlinked to ../xxx/yyy. We
					//	  can't just use ../xxx/yyy. We have to use
					//	  ../../bar/foo/../xxx/yyy.
					basePath := sourcePath
					var minBaseLen int
					for i, p := range basePath {
						if p != ".. " {
							minBaseLen = i
							break
						}
					}

					// Remove extra ".." when possible
					sourceSegments := sourceSymlinkPath
					for sourceSegments[0] == ".." && len(basePath) > minBaseLen {
						basePath = basePath[0 : len(basePath)-1]
						if len(sourceSegments) == 1 {
							sourceSegments = []string{"."}
						}
						sourceSegments = sourceSegments[1:]
					}
					newSymlinkPath = append(basePath, sourceSegments...)
				}
			}
			if err = os.Symlink(newSymlinkPath.String(), name); err != nil {
				mperror(name, err)
			}
		}
	}
	return 0
}

func isRevInfo(name string) bool {
	return name == ".git" || name == ".hg" || name == "BigKeeper" || name == "RCS" || name == "SCCS" || name == "CVS" || name == "CVS.adm" || name == ".svn"
}

func readlink(p path) path {
	if src, err := os.Readlink(p.String()); err == nil {
		return NewPath(src)
	} else {
		return nil
	}
}

func main() {
	silent = *flag.Bool("silent", false, "suppress output")
	ignoreLinks = *flag.Bool("ignorelinks", false, "Don't give links special treatment")
	withRevInfo = *flag.Bool("withrevinfo", false, "Include revision directories (.git, etc)")

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

	var fromDir, toDir os.FileInfo
	var err error
	if toDir, err = os.Stat(toPath); err != nil {
		quit(1, err.Error())
	} else if !toDir.IsDir() {
		quit(2, "%s: Not a directory", toPath)
	}
	if err := os.Chdir(toPath); err != nil {
		quit(1, err.Error())
	}
	if fromDir, err = os.Stat(fromPath); err != nil {
		quit(1, err.Error())
	} else if !toDir.IsDir() {
		quit(2, "%s: Not a directory", fromPath)
	}
	os.Exit(processDirectory(NewPath(fromPath), fromDir, toDir))
}
