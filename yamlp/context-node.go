package yamlp

import (
	"container/list"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

type ContextNode struct {
	candidateNode *yqlib.CandidateNode
	decoded       interface{}
}

func NewContextNode(n *yqlib.CandidateNode) *ContextNode {
	return &ContextNode{
		candidateNode: n,
	}
}

func (n *ContextNode) Merge(rhs *yqlib.CandidateNode) (*ContextNode, error) {
	err := yqlib.NewDataTreeNavigator().DeeplyAssign(createMergeContext(n.candidateNode), []interface{}{}, rhs)
	if err != nil {
		return nil, err
	}

	return &ContextNode{
		candidateNode: n.candidateNode,
	}, nil
}

func createMergeContext(lhs *yqlib.CandidateNode) yqlib.Context {
	inputCandidates := list.New()
	inputCandidates.PushBack(lhs)

	return yqlib.Context{
		MatchingNodes: inputCandidates,
		Variables:     map[string]*list.List{},
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

func (n *ContextNode) ForEachNode(iter func(vars map[string]*ContextNode)) {
	for _, v := range n.candidateNode.Content {
		iter(map[string]*ContextNode{
			"v": NewContextNode(v),
		})
	}
}

func (n *ContextNode) Reduce(initial *yqlib.CandidateNode, iter func(vars map[string]*ContextNode) (*yqlib.CandidateNode, error), update func(*yqlib.CandidateNode, *yqlib.CandidateNode)) error {
	for _, v := range n.candidateNode.Content {
		node, err := iter(map[string]*ContextNode{
			"v": NewContextNode(v),
		})
		if err != nil {
			return err
		}

		update(initial, node)
	}

	return nil
}

func newSeqNode() *yqlib.CandidateNode {
	return &yqlib.CandidateNode{
		Kind:    yqlib.SequenceNode,
		Content: []*yqlib.CandidateNode{},
	}
}

func updateSeqNode(seq *yqlib.CandidateNode, item *yqlib.CandidateNode) {
	seq.Content = append(seq.Content, item)
}
