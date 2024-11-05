package yamlp

import "github.com/mikefarah/yq/v4/pkg/yqlib"

type ContextNode struct {
	candidateNode *yqlib.CandidateNode
	decoded       interface{}
}

func NewContextNode(n *yqlib.CandidateNode) *ContextNode {
	return &ContextNode{
		candidateNode: n,
	}
}

func (n *ContextNode) Interface() (interface{}, error) {
	if n.decoded != nil {
		return n.decoded, nil
	}

	yn, err := n.candidateNode.MarshalYAML()
	if err != nil {
		return nil, err
	}
	var i interface{}
	err = yn.Decode(&i)
	if err != nil {
		return nil, err
	}

	n.decoded = i

	return i, nil
}
