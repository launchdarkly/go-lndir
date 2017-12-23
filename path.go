package lndir

import "path/filepath"

type path []string

func (p path) isAbs() bool {
	return p[0] == "/"
}

func (p path) String() string {
	return filepath.Join(p...)
}

func (p path) List() []string {
	return []string(p)
}
