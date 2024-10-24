package tags

import (
	"container/list"
	"fmt"

	"github.com/glopal/go-yamlplus/yamlp"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	yqlib.InitExpressionParser()

	yamlp.AddTagResolver("!yq", yqResolver)
}

func yqResolver(n *yqlib.CandidateNode, nc yamlp.NodeContext, refs map[string]*yamlp.Node) (*yqlib.CandidateNode, error) {
	inputCandidates := list.New()
	inputCandidates.PushBack(n)

	yqctx := yqlib.Context{
		MatchingNodes: inputCandidates,
		Variables:     refsToVariables(refs),
	}

	expr, err := yqlib.ExpressionParser.ParseExpression(n.Value)
	if err != nil {
		return nil, err
	}

	context, err := yqlib.NewDataTreeNavigator().GetMatchingNodes(yqctx, expr)
	if err != nil {
		return nil, err
	}

	if context.MatchingNodes.Len() == 0 {
		return &yqlib.CandidateNode{
			Kind:  yqlib.ScalarNode,
			Style: yqlib.DoubleQuotedStyle,
			Tag:   "!!str",
		}, nil
	}

	nn, ok := context.MatchingNodes.Front().Value.(*yqlib.CandidateNode)
	if !ok {
		return nil, fmt.Errorf("yq expression error (%s): failed to marshal CandidateNode", n.Value)
	}

	return nn, nil
}

func refsToVariables(refs map[string]*yamlp.Node) map[string]*list.List {
	vars := map[string]*list.List{}

	for ref, node := range refs {
		if node.GetResolveCount() == 0 {
			node.Resolve(refs)
		}

		l := list.New()
		l.PushBack(node.CandidateNode)

		vars[ref] = l
	}

	return vars
}
