package lndir

/*

Derived from lndir.c, which includes the copyright notice below.

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
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"gopkg.in/src-d/go-billy.v3/osfs"
	"gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"
)

var isOSX = runtime.GOOS == "darwin"

type Config struct {
	Silent, IgnoreLinks, WithRevInfo, UseGitignore bool
	Logger                                         Logger
}

type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type directoryLinker struct {
	silent, ignoreLinks, withRevInfo, useGitignore bool
	gitignoreMatcher                               gitignore.Matcher
	currentPath                                    path
	logger                                         Logger
}

type userError struct {
	error
}

func newUserError(format string, v ...interface{}) userError {
	return userError{fmt.Errorf(format, v...)}
}

func IsUserError(err error) bool {
	_, isUserError := err.(userError)
	return isUserError
}

func Lndir(fromPath, toPath string, config Config) error {
	logger := config.Logger

	if logger == nil {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	linker := directoryLinker{
		silent:      config.Silent,
		ignoreLinks: config.IgnoreLinks,
		withRevInfo: config.WithRevInfo,
		logger:      logger,
	}

	sourcePath, sourceErr := newPath(fromPath)
	if sourceErr != nil {
		return sourceErr
	}

	if config.UseGitignore {
		absPath, _ := filepath.Abs(sourcePath.String())
		fs := osfs.New(absPath)
		if patterns, err := gitignore.ReadPatterns(fs, []string{}); err != nil {
			return fmt.Errorf("%s: Cannot read gitignore patterns: %s", absPath, err)
		} else {
			linker.gitignoreMatcher = gitignore.NewMatcher(patterns)
		}
	}

	var fromDir, toDir os.FileInfo
	var err error
	if toDir, err = os.Stat(toPath); err != nil {
		return err
	} else if !toDir.IsDir() {
		return newUserError("%s: Not a directory", toPath)
	}
	if err := os.Chdir(toPath); err != nil {
		return err
	}
	if fromDir, err = os.Stat(fromPath); err != nil {
		return err
	} else if !fromDir.IsDir() {
		return newUserError("%s: Not a directory", fromPath)
	}

	return linker.processDirectory(sourcePath, fromDir, toDir, len(sourcePath.List()))
}

func newPath(pathStr string) (path, error) {
	if pathStr == "" {
		return nil, fmt.Errorf("empty path: %s", pathStr)
	}
	segments := strings.Split(pathStr, string(filepath.Separator))
	if filepath.IsAbs(pathStr) {
		return append([]string{"/"}, segments...), nil
	} else {
		return segments, nil
	}
}

func (l *directoryLinker) logPrintf(format string, v ...interface{}) {
	l.logFileName()
	l.logger.Printf(format, v...)
}

func (l *directoryLinker) logError(msg string, err error) {
	l.logFileName()
	if msg != "" {
		l.logger.Printf("%s", msg+":")
	}
	l.logger.Println(err)
}

func (l *directoryLinker) logFileName() {
	l.logger.Printf("%s:\n", l.currentPath)
}

func equivalent(lname path, rname path) bool {
	return filepath.Clean(lname.String()) == filepath.Clean(rname.String())
}

func (l *directoryLinker) processSubdir(subdirName string, parentPath path, subdirInfo os.FileInfo, relativeDepth int) (err error) {
	if subdirName == "." || subdirName == ".." {
		return
	}

	if !l.withRevInfo && isRevInfo(subdirName) {
		return
	}

	// These are maintained for printing the path before an error
	originalPath := l.currentPath
	l.currentPath = parentPath
	if !l.silent {
		fmt.Printf("%s:\n", parentPath)
	}

	// Restore these when the method is done
	defer func() {
		l.currentPath = originalPath
	}()

	var targetInfo os.FileInfo
	if targetInfo, err = os.Stat(subdirName); err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(subdirName, os.FileMode(0777)); err == nil {
				if targetInfo, err = os.Stat(subdirName); err != nil {
					l.logError(subdirName, err)
					return
				}
			}
		} else {
			l.logError(subdirName, err)
			return
		}
	}

	_, err = os.Readlink(subdirName)
	if err == nil {
		l.logPrintf("%s: is a link instead of a directory", subdirName)
		return
	}

	if err = os.Chdir(subdirName); err != nil {
		l.logError(subdirName, err)
		return
	}
	defer func() {
		err = os.Chdir("..")
	}()

	srcPath := parentPath
	if !srcPath.isAbs() {
		relativeDepth += 1
		parentRoot, _ := newPath("..")
		srcPath = append(parentRoot, parentPath...)
	}

	err = l.processDirectory(srcPath, subdirInfo, targetInfo, relativeDepth)
	return
}

func (l *directoryLinker) processDirectory(sourceDirPath path, sourceDir os.FileInfo, targetDir os.FileInfo, baseDepth int) error {
	if os.SameFile(sourceDir, targetDir) {
		return newUserError("%s: From and to directories are identical!", sourceDirPath)
	}

	var err error
	var f *os.File
	if f, err = os.Open(sourceDirPath.String()); err != nil {
		return fmt.Errorf("%s: Cannot open directory: %s", sourceDirPath, err)
	}

	// Determine the maximum number of directories we might see
	dirsLeft := math.MaxInt32
	if s, err := f.Stat(); err != nil {
		return fmt.Errorf("%s: Cannot stat: %s", f.Name(), err)
	} else if stat, ok := s.Sys().(*syscall.Stat_t); ok && stat != nil {
		// Apparently, if this is 1, we have no clue about how many subdirectories there are in this directory
		if stat.Nlink != 1 {
			dirsLeft = int(stat.Nlink)
		}
	}

	var children []string
	if children, err = f.Readdirnames(0); err != nil {
		return fmt.Errorf("%s: Cannot readdir: %s", f.Name(), err)
	}
	f.Close()

	for _, name := range children {
		if strings.HasSuffix(name, "~") {
			continue
		}

		if isOSX && (name == ".DS_Store" || name == "._.DS_Store") {
			continue
		}

		sourcePath := append(sourceDirPath, name)

		isDir := false

		// Optimization to skip these checks once all directory entries have been processed
		var childInfo os.FileInfo
		if dirsLeft > 0 {
			if childInfo, err = os.Lstat(sourcePath.String()); err != nil {
				l.logError(sourcePath.String(), err)
				continue
			}

			isDir = childInfo.IsDir()
			if isDir {
				dirsLeft -= 1
			}
		}

		if l.gitignoreMatcher != nil && l.gitignoreMatcher.Match(sourcePath.List()[baseDepth:], isDir) {
			continue
		}

		if isDir {
			l.processSubdir(name, sourcePath, childInfo, baseDepth)
			continue
		}

		// The option to ignore links exists mostly because
		//   checking for them slows us down by 10-20%.
		//   But it is off by default because this really is a useful check.
		var sourceSymlinkPath path
		if !l.ignoreLinks {
			// see if the file in the base tree was a symlink
			sourceSymlinkPath = readlink(sourcePath)
		}

		namePath, err := newPath(name)
		if err != nil {
			l.logPrintf("error processing %s: %s", name, err)
			continue
		}
		existingSymlinkPath := readlink(namePath)
		if existingSymlinkPath != nil {
			// Link exists in new tree.  Print message if it doesn't match.
			expectedSymlinkPath := sourcePath
			if sourceSymlinkPath != nil {
				expectedSymlinkPath = sourceSymlinkPath
			}
			if !equivalent(existingSymlinkPath, expectedSymlinkPath) {
				l.logPrintf("%s: %s", name, existingSymlinkPath)
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
				l.logError(name, err)
			}
		}
	}
	return nil
}

func isRevInfo(name string) bool {
	return name == ".git" || name == ".hg" || name == "BigKeeper" || name == "RCS" || name == "SCCS" || name == "CVS" || name == "CVS.adm" || name == ".svn"
}

func readlink(p path) path {
	if src, err := os.Readlink(p.String()); err == nil {
		srcPath, _ := newPath(src)
		return srcPath
	} else {
		return nil
	}
}
