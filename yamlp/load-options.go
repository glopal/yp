package yamlp

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type loadOptions struct {
	omitFunc func(string) bool
	fs       fs.FS
}

func defaultLoadOptions() *loadOptions {
	return &loadOptions{
		omitFunc: func(s string) bool {
			return false
		},
		fs: os.DirFS("."),
	}
}

func WithFS(fsys fs.FS) func(*loadOptions) {
	return func(lo *loadOptions) {
		lo.fs = fsys
	}
}

func OmitLeadingUnderscore() func(*loadOptions) {
	return func(lo *loadOptions) {
		lo.omitFunc = func(path string) bool {
			return strings.HasPrefix(filepath.Base(path), "_")
		}
	}
}

func OmitDotFiles() func(*loadOptions) {
	return func(lo *loadOptions) {
		lo.omitFunc = func(path string) bool {
			return strings.HasPrefix(filepath.Base(path), ".")
		}
	}
}
func (lo *loadOptions) isOSFS() bool {
	return lo.fs == nil
}

func (lo *loadOptions) walkDir(root string, walkFunc fs.WalkDirFunc) error {
	if lo.isOSFS() {
		return filepath.WalkDir(root, walkFunc)
	}

	return fs.WalkDir(lo.fs, root, walkFunc)
}
