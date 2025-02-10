package yplib

import (
	"container/list"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"github.com/spf13/afero"
)

type Nodes struct {
	nodes    []*Node
	exports  *Exports
	out      *Node
	outNodes []OutNodes
	opts     *loadOptions
}

func NewNodes() *Nodes {
	return &Nodes{
		nodes: make([]*Node, 0),
		exports: &Exports{
			files:   map[string]*exportFile{},
			exports: map[string]*Node{},
		},
	}
}

func (ns *Nodes) Nodes() []*Node {
	return ns.nodes
}

func (ns *Nodes) CandidateNodes() []*yqlib.CandidateNode {
	cnodes := make([]*yqlib.CandidateNode, 0, len(ns.nodes))

	for _, n := range ns.nodes {
		cnodes = append(cnodes, n.CandidateNode)
	}

	return cnodes
}

func (ns *Nodes) Push(n *Node) error {
	if n.IsRefOrExport() {
		return ns.exports.Push(n)
	} else if n.Kind == Out {
		ns.out = n
	} else {
		ns.nodes = append(ns.nodes, n)
	}

	return nil
}

func (ns *Nodes) Resolve() error {
	exports, err := ns.exports.resolve(ns.opts)
	if err != nil {
		return err
	}

	for _, n := range ns.nodes {
		vars := map[string]*ContextNode{}
		vars["self"] = NewContextNode(n.CandidateNode)

		err = n.Resolve(exports, vars, ns.opts)
		if err != nil {
			return err
		}
	}

	outNodes, err := ns.resolveOut()
	if err != nil {
		return err
	}

	ns.outNodes = outNodes

	return nil
}

func mergeRefs(refs map[string]*Node, exports map[string]*Node) map[string]*Node {
	merged := map[string]*Node{}

	for k, v := range refs {
		merged[k] = v
	}
	for k, v := range exports {
		merged[k] = v
	}

	return merged
}

func (ns *Nodes) PrettyPrintYaml(w io.Writer) {
	for _, n := range ns.nodes {
		w.Write([]byte("---\n"))
		n.PrettyPrintYaml(w)
	}
}

func (ns *Nodes) Out() error {
	prefs := yqlib.NewDefaultYamlPreferences()
	prefs.UnwrapScalar = false
	prefs.PrintDocSeparators = true
	prefs.Indent = 2

	for _, out := range ns.outNodes {
		prefs.ColorsEnabled = false
		w := out.writer
		if w == nil {
			w = out.file
			prefs.ColorsEnabled = shouldColorize(out.file)
			defer out.file.Close()
		}
		l := list.New()

		printer := yqlib.NewPrinter(yqlib.NewYamlEncoder(prefs), yqlib.NewSinglePrinterWriter(w))

		for docIndex, cn := range out.node.Content {
			// setting the document index and parent to nil enables doc separator printing
			cn.SetDocument(uint(docIndex))
			cn.Parent = nil

			l.PushBack(cn)
		}

		err := printer.PrintResults(l)
		if err != nil {
			return err
		}
	}

	return nil
}

type OutNodes struct {
	node   *yqlib.CandidateNode
	file   afero.File
	writer io.Writer
}

func (ns *Nodes) resolveOut() ([]OutNodes, error) {
	if ns.out == nil {
		return []OutNodes{{
			node: &yqlib.CandidateNode{
				Kind:    yqlib.SequenceNode,
				Content: ns.CandidateNodes(),
			},
			file:   os.Stdout,
			writer: ns.opts.writer,
		}}, nil
	}

	err := ns.out.Resolve(NewOutContext(ns))
	if err != nil {
		return nil, err
	}

	if ns.out.CandidateNode.Kind > yqlib.MappingNode {
		return nil, errors.New("#out node must be a map or sequence")
	}

	if ns.out.CandidateNode.Kind == yqlib.SequenceNode {
		return []OutNodes{{
			node: ns.out.CandidateNode,
			file: os.Stdout,
		}}, nil
	}

	outNodes := make([]OutNodes, 0, len(ns.out.CandidateNode.Content)/2)
	os := ns.opts.os

	for i := 0; i < len(ns.out.CandidateNode.Content); i += 2 {
		path := ns.out.CandidateNode.Content[i].Value

		if path == "/dev/stdout" {
			outNodes = append(outNodes, ns.opts.getStdoutOutNodes(ns.out.CandidateNode.Content[i+1]))
			continue
		}
		if path != filepath.Base(path) {
			err := os.MkdirAll(filepath.Dir(path), 0755)
			if err != nil {
				return nil, err
			}
		}

		file, err := os.Create(path)
		if err != nil {
			return nil, err
		}

		outNodes = append(outNodes, OutNodes{
			file: file,
			node: ns.out.CandidateNode.Content[i+1],
		})
	}

	return outNodes, nil
}
