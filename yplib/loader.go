package yplib

import (
	"errors"
)

type NodeLoader struct {
	paths    []string
	opts     *loadOptions
	nodes    *Nodes
	err      error
	resolved bool
}

func Load(paths ...string) *NodeLoader {
	return &NodeLoader{
		paths: paths,
		opts:  defaultLoadOptions(),
	}
}

func WithOptions(opts ...func(*loadOptions)) *NodeLoader {
	return (&NodeLoader{
		paths: []string{},
		opts:  defaultLoadOptions(),
	}).Options(opts...)
}

func (l *NodeLoader) Load(paths ...string) *NodeLoader {
	l.paths = append(l.paths, paths...)

	return l
}

func (l *NodeLoader) Options(opts ...func(*loadOptions)) *NodeLoader {
	for _, o := range opts {
		o(l.opts)
	}

	return l
}

func (l *NodeLoader) load() {
	if l.nodes != nil || l.err != nil {
		return
	}

	files := []NamedReader{}

	for _, path := range l.paths {
		file, err := getNamedReader(path, l.opts)
		if err != nil {
			l.err = err
			return
		}

		fileInfo, err := file.Stat()
		if err != nil {
			l.err = err
			return
		}

		if fileInfo.IsDir() {
			readers, err := getDirReaders(path, l.opts)
			if err != nil {
				l.err = err
				return
			}

			files = append(files, readers...)
		} else {
			files = append(files, file)
		}
	}

	l.nodes, l.err = loadReaders(files...)
	if l.err == nil {
		l.nodes.opts = l.opts
	}
}

func (l *NodeLoader) Nodes() (*Nodes, error) {
	l.load()
	return l.nodes, l.err
}

func (l *NodeLoader) Resolve() *NodeLoader {
	l.load()
	if l.err != nil || l.resolved {
		return l
	}

	err := l.nodes.Resolve()
	if err != nil {
		l.err = err
		return l
	}

	l.resolved = true

	return l
}

func (l *NodeLoader) Error() error {
	return l.err
}

func (l *NodeLoader) Out() error {
	if err := l.Resolve().Error(); err != nil {
		return err
	}

	return l.nodes.Out()
}

func (l *NodeLoader) Decode(i any) error {
	if err := l.Resolve().Error(); err != nil {
		return err
	}

	if len(l.nodes.outNodes) != 1 {
		return errors.New("decoder operates on a single out node list, might change this later")
	}

	yn, err := l.nodes.outNodes[0].node.MarshalYAML()
	if err != nil {
		return err
	}

	return yn.Decode(i)
}
