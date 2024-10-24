package yamlp

import "os"

type Nodes struct {
	nodes []*Node
	refs  map[string]*Node
}

func NewNodes() *Nodes {
	return &Nodes{
		nodes: make([]*Node, 0),
		refs:  map[string]*Node{},
	}
}

func (ns *Nodes) Append(n *Nodes) {
	ns.nodes = append(ns.nodes, n.nodes...)

	for k, v := range n.refs {
		ns.refs[k] = v
	}
}

func (ns *Nodes) Nodes() []*Node {
	return ns.nodes
}

func (ns *Nodes) Resolve() error {
	var err error

	// for _, n := range ns.refs {
	// 	err = n.Resolve(ns.refs)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	for _, n := range ns.nodes {
		err = n.Resolve(ns.refs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ns *Nodes) PrettyPrintYaml(w *os.File) {
	for _, n := range ns.nodes {
		w.Write([]byte("---\n"))
		n.PrettyPrintYaml(w)
	}
}
