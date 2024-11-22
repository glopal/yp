package yplib

import (
	"container/list"
	"errors"
	"os"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/op/go-logging.v1"
)

type ContextNode struct {
	candidateNode *yqlib.CandidateNode
	decoded       interface{}
}

var cleanMergesExpr *yqlib.ExpressionNode

func init() {
	// disable yqlib debug logging
	leveled := logging.AddModuleLevel(logging.NewLogBackend(os.Stderr, "", 0))
	leveled.SetLevel(logging.ERROR, "")
	yqlib.GetLogger().SetBackend(leveled)

	yqlib.InitExpressionParser()
	cleanMergesExpr, _ = yqlib.ExpressionParser.ParseExpression(`(... | select(tag == "!!merge")) |= (. tag =  "!!str" | . style = "folded")`)
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

func NewOutContext(nodes *Nodes) (*ContextNode, map[string]*ContextNode, *loadOptions) {
	return &ContextNode{
			candidateNode: &yqlib.CandidateNode{
				Kind:    yqlib.SequenceNode,
				Content: nodes.CandidateNodes(),
			},
		}, map[string]*ContextNode{
			"ctx": nodes.exports.contextNode,
		}, nodes.opts
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

	cleaned, err := cleanMerges(n.candidateNode)
	if err != nil {
		return nil, err
	}

	yn, err := cleaned.MarshalYAML()
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

func cleanMerges(n *yqlib.CandidateNode) (*yqlib.CandidateNode, error) {
	context, err := yqlib.NewDataTreeNavigator().GetMatchingNodes(createMergeContext(n), cleanMergesExpr)
	if err != nil {
		return nil, err
	}
	if context.MatchingNodes.Len() == 0 {
		return nil, errors.New("Unable to resolve")
	}

	cleaned, ok := context.MatchingNodes.Front().Value.(*yqlib.CandidateNode)
	if !ok {
		return nil, errors.New("failed to clean merges")
	}

	return cleaned, nil
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
