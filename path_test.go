package lndir

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	t.Parallel()

	specs := []struct {
		path           string
		expectedList   []string
		expectedString string
		isAbsolute     bool
	}{
		{".", []string{"."}, ".", false},
		{"..", []string{".."}, "..", false},
		{"file", []string{"file"}, "file", false},
		{"../file", []string{"..", "file"}, "../file", false},
		{"/file", []string{"/", "", "file"}, "/file", true},
	}

	for _, spec := range specs {
		t.Run(spec.path, func(t *testing.T) {
			path, err := newPath(spec.path)
			assert.NoError(t, err)
			assert.Equal(t, spec.expectedString, path.String())
			assert.Equal(t, spec.expectedList, path.List())
			assert.Equal(t, spec.isAbsolute, path.isAbs())
		})
	}

	t.Run("empty path generates error", func(t *testing.T) {
		_, err := newPath("")
		assert.Error(t, err)
	})
}
