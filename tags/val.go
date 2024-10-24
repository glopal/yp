package tags

import (
	"github.com/glopal/go-yamlplus/yamlp"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	yamlp.AddTagResolver("!val", valResolver)
}

func valResolver(n *yqlib.CandidateNode, nc yamlp.NodeContext, refs map[string]*yamlp.Node) (*yqlib.CandidateNode, error) {
	nn := n.Copy()
	nn.Tag = "!!str"
	nn.Value = "val-" + n.Value
	nn.Style = 0
	return nn, nil
}
