package yamlp

import (
	"io"
)

type Nodes struct {
	nodes   []*Node
	exports Exports
}

func NewNodes() *Nodes {
	return &Nodes{
		nodes: make([]*Node, 0),
		exports: Exports{
			files:   map[string]*exportFile{},
			exports: map[string]*Node{},
		},
	}
}

func (ns *Nodes) Nodes() []*Node {
	return ns.nodes
}

func (ns *Nodes) Push(n *Node) error {
	if n.IsRefOrExport() {
		return ns.exports.Push(n)
	} else {
		ns.nodes = append(ns.nodes, n)
	}

	return nil
}

func (ns *Nodes) Resolve() error {
	exports, err := ns.exports.resolve()
	if err != nil {
		return err
	}

	for _, n := range ns.nodes {
		err = n.Resolve(exports, nil)
		if err != nil {
			return err
		}
	}

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

// func (ns *Nodes) getRefExportOrder() {
// 	for _, refs := range ns.refs {
// 		for name, ref := range refs {
// 			ref.GetImports()
// 		}
// 	}
// }

func (ns *Nodes) PrettyPrintYaml(w io.Writer) {
	for _, n := range ns.nodes {
		w.Write([]byte("---\n"))
		n.PrettyPrintYaml(w)
	}
}
