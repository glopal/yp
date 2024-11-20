package yplib

import (
	"container/list"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

type ContextNode struct {
	candidateNode *yqlib.CandidateNode
	decoded       interface{}
}

func NewContextNode(n *yqlib.CandidateNode) *ContextNode {
	if n == nil {
		n = &yqlib.CandidateNode{
			Kind: yqlib.ScalarNode,
			Tag:  "!!null",
		}
	}
	return &ContextNode{
		candidateNode: n,
	}
}

func NewOutContextNode(nodes *Nodes) *ContextNode {
	ctx := &ContextNode{
		candidateNode: &yqlib.CandidateNode{
			Kind: yqlib.MappingNode,
		},
	}

	ctx.candidateNode.AddKeyValueChild(&yqlib.CandidateNode{
		Kind:  yqlib.ScalarNode,
		Value: "nodes",
	}, &yqlib.CandidateNode{
		Kind:    yqlib.SequenceNode,
		Content: nodes.CandidateNodes(),
	})

	ctx.candidateNode.AddKeyValueChild(&yqlib.CandidateNode{
		Kind:  yqlib.ScalarNode,
		Value: "ctx",
	}, nodes.exports.contextNode.candidateNode)

	return ctx
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

func (n *ContextNode) Reduce(kind yqlib.Kind, iter func(vars map[string]*ContextNode) (*yqlib.CandidateNode, error)) (*yqlib.CandidateNode, error) {
	var initial *yqlib.CandidateNode
	var update func(*yqlib.CandidateNode, *yqlib.CandidateNode) error

	switch kind {
	case yqlib.MappingNode:
		initial = newMapNode()
		update = updateMapNode
	default:
		initial = newSeqNode()
		update = updateSeqNode
	}

	for _, v := range n.candidateNode.Content {
		node, err := iter(map[string]*ContextNode{
			"v": NewContextNode(v),
		})
		if err != nil {
			return nil, err
		}

		err = update(initial, node)
		if err != nil {
			return nil, err
		}
	}

	return initial, nil
}

func newSeqNode() *yqlib.CandidateNode {
	return &yqlib.CandidateNode{
		Kind:    yqlib.SequenceNode,
		Content: []*yqlib.CandidateNode{},
	}
}

func updateSeqNode(seq *yqlib.CandidateNode, item *yqlib.CandidateNode) error {
	seq.Content = append(seq.Content, item.Content...)
	return nil
}

func newMapNode() *yqlib.CandidateNode {
	return &yqlib.CandidateNode{
		Kind:    yqlib.MappingNode,
		Content: []*yqlib.CandidateNode{},
	}
}

// func mapNodeReducer() (*yqlib.CandidateNode, func(*yqlib.CandidateNode, *yqlib.CandidateNode)) {
// 	initial := newMapNode()

//		return initial, func(mapNode *yqlib.CandidateNode, item *yqlib.CandidateNode) {
//			mapNode.AddChildren(item.Content)
//		}
//	}
func updateMapNode(mapNode *yqlib.CandidateNode, item *yqlib.CandidateNode) error {
	return yqlib.NewDataTreeNavigator().DeeplyAssign(createMergeContext(mapNode), []interface{}{}, item)
}
