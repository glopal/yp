package yplib

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"github.com/spf13/afero"
)

type loadOptions struct {
	omitFunc func(string) bool
	ifs      fs.FS
	os       afero.Fs
	writer   io.Writer
}

func defaultLoadOptions() *loadOptions {
	return &loadOptions{
		omitFunc: func(s string) bool {
			return false
		},
		os: afero.NewOsFs(),
	}
}

func WithFS(fsys fs.FS) func(*loadOptions) {
	return func(lo *loadOptions) {
		lo.ifs = fsys
	}
}

func WithOutputFS(fsys afero.Fs) func(*loadOptions) {
	return func(lo *loadOptions) {
		lo.os = fsys
	}
}

func WithWriter(w io.Writer) func(*loadOptions) {
	return func(lo *loadOptions) {
		lo.writer = w
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
func (lo *loadOptions) isInputOS() bool {
	return lo.ifs == nil
}
func (lo *loadOptions) getStdoutOutNodes(cnode *yqlib.CandidateNode) OutNodes {
	out := OutNodes{
		node: cnode,
	}

	if lo.writer != nil {
		out.writer = lo.writer
	} else {
		out.file = os.Stdout
	}

	return out
}

func (lo *loadOptions) walkDir(root string, walkFunc fs.WalkDirFunc) error {
	if lo.isInputOS() {
		return filepath.WalkDir(root, walkFunc)
	}

	return fs.WalkDir(lo.ifs, root, walkFunc)
}
