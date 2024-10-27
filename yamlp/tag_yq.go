package yamlp

import (
	"container/list"
	"fmt"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	yqlib.InitExpressionParser()

	AddTagResolver("!yq", yqResolver)
}

func yqResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	fmt.Println(rc.Target.GetPath())
	inputCandidates := list.New()
	inputCandidates.PushBack(rc.Target)

	yqctx := yqlib.Context{
		MatchingNodes: inputCandidates,
		Variables:     createVariables(rc.Node.CandidateNode, rc.Refs),
	}

	expr, err := yqlib.ExpressionParser.ParseExpression(rc.Target.Value)
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
		return nil, fmt.Errorf("yq expression error (%s): failed to marshal CandidateNode", rc.Target.Value)
	}

	return nn, nil
}

func createVariables(root *yqlib.CandidateNode, refs map[string]*Node) map[string]*list.List {
	vars := map[string]*list.List{}

	for ref, node := range refs {
		if node.GetResolveCount() == 0 {
			node.Resolve(refs)
		}

		l := list.New()
		l.PushBack(node.CandidateNode)

		vars[ref] = l
	}

	l := list.New()
	l.PushBack(root)

	vars["_"] = l

	return vars
}
