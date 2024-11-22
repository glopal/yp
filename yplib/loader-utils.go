package yplib

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

var decoder = yqlib.NewYamlDecoder(yqlib.YamlPreferences{
	Indent:                      2,
	ColorsEnabled:               false,
	LeadingContentPreProcessing: false,
	PrintDocSeparators:          true,
	UnwrapScalar:                true,
	EvaluateTogether:            false,
})

type NamedReader interface {
	fs.File
	Name() string
}

type FsFileWrapper struct {
	fs.File
	name string
}

func (f FsFileWrapper) Name() string {
	return f.name
}

func getNamedReader(path string, opts *loadOptions) (NamedReader, error) {
	if opts.isOSFS() {
		return os.Open(path)
	}

	f, err := opts.fs.Open(path)
	if err != nil {
		return nil, err
	}

	return FsFileWrapper{f, path}, nil
}

func getDirReaders(dir string, opts *loadOptions) ([]NamedReader, error) {
	files := []NamedReader{}

	err := opts.walkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if opts.omitFunc(path) {
			if d.IsDir() {
				return fs.SkipDir
			}

			return nil
		}

		if d.IsDir() || !IsYamlFile(path) {
			return nil
		}

		f, err := getNamedReader(path, opts)
		if err != nil {
			return err
		}

		files = append(files, f)

		return nil
	})

	return files, err
}

func loadReaders(files ...NamedReader) (*Nodes, error) {
	nodes := NewNodes()

	for _, f := range files {
		err := decoder.Init(f)
		if err != nil {
			return nil, err
		}

		for {
			node, err := decoder.Decode()
			if err != nil {
				// break the loop in case of EOF
				if errors.Is(err, io.EOF) {
					break
				} else {
					return nil, err
				}
			}

			err = nodes.Push(NewNode(node, f.Name()))
			if err != nil {
				return nil, err
			}
		}
	}

	return nodes, nil
}

func LoadFile(file string, opts *loadOptions) (*Nodes, error) {
	f, err := getNamedReader(file, opts)
	if err != nil {
		return nil, err
	}

	return loadReaders(f)
}

func IsYamlFile(file string) bool {
	ext := filepath.Ext(file)
	return ext == ".yml" || ext == ".yaml"
}
