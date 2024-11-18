package yamlp

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/op/go-logging.v1"
)

type NamedReader interface {
	io.Reader
	Name() string
}

type FsFileWrapper struct {
	fs.File
	name string
}

func (f FsFileWrapper) Name() string {
	return f.name
}

var decoder = yqlib.NewYamlDecoder(yqlib.YamlPreferences{
	Indent:                      2,
	ColorsEnabled:               false,
	LeadingContentPreProcessing: false,
	PrintDocSeparators:          true,
	UnwrapScalar:                true,
	EvaluateTogether:            false,
})

func init() {
	// disable yqlib debug logging
	leveled := logging.AddModuleLevel(logging.NewLogBackend(os.Stderr, "", 0))
	leveled.SetLevel(logging.ERROR, "")
	yqlib.GetLogger().SetBackend(leveled)
}

func LoadDirFS(fsys fs.FS, dir string, opts ...func(*loadOptions)) (*Nodes, error) {
	options := defaultLoadOptions()
	for _, o := range opts {
		o(options)
	}

	files := []NamedReader{}

	err := fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !IsYamlFile(path) || options.omitFunc(path) {
			return nil
		}

		f, err := fsys.Open(path)
		if err != nil {
			return err
		}

		files = append(files, FsFileWrapper{f, path})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return LoadReaders(files...)
}

func LoadDir(dir string, opts ...func(*loadOptions)) (*Nodes, error) {
	readers, err := getDirReaders(dir, opts...)
	if err != nil {
		return nil, err
	}

	return LoadReaders(readers...)
}

func getDirReaders(dir string, opts ...func(*loadOptions)) ([]NamedReader, error) {
	options := defaultLoadOptions()
	for _, o := range opts {
		o(options)
	}

	files := []NamedReader{}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !IsYamlFile(path) || options.omitFunc(path) {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		files = append(files, f)

		return nil
	})

	return files, err
}

func LoadFile(file string) (*Nodes, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	return LoadReaders(f)
}

func Load(paths []string, opts ...func(*loadOptions)) (*Nodes, error) {
	files := []NamedReader{}

	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		fileInfo, err := file.Stat()
		if err != nil {
			return nil, err
		}

		if fileInfo.IsDir() {
			readers, err := getDirReaders(path)
			if err != nil {
				return nil, err
			}

			files = append(files, readers...)
		} else {
			files = append(files, file)
		}
	}

	return LoadReaders(files...)
}
func LoadReaders(files ...NamedReader) (*Nodes, error) {
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

func IsYamlFile(file string) bool {
	ext := filepath.Ext(file)
	return ext == ".yml" || ext == ".yaml"
}
