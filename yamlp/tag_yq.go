package yamlp

import (
	"container/list"
	"errors"
	"fmt"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	yqlib.InitExpressionParser()

	AddTagResolver("!yq", yqResolver)
}

func yqResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	expr, err := yqlib.ExpressionParser.ParseExpression(rc.Target.Value)
	if err != nil {
		return nil, err
	}

	context, err := yqlib.NewDataTreeNavigator().GetMatchingNodes(createContext(rc), expr)
	if err != nil {
		return nil, err
	}

	if context.MatchingNodes.Len() == 0 {
		if rc.Node.IsRef() {
			fmt.Println(rc.Node.Kind.String(), rc.Node.File)
			return nil, errors.New("Unable to resolve")
		}

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

func createContext(rc ResolveContext) yqlib.Context {
	inputCandidates := list.New()
	inputCandidates.PushBack(rc.Ctx.candidateNode)

	return yqlib.Context{
		MatchingNodes: inputCandidates,
		Variables:     createVariables(rc),
	}
}

func createVariables(rc ResolveContext) map[string]*list.List {
	vars := map[string]*list.List{}

	l := list.New()
	l.PushBack(&yqlib.CandidateNode{
		Kind:  yqlib.ScalarNode,
		Tag:   "!!str",
		Value: rc.Node.Dir,
	})

	vars["DIR"] = l

	return vars
}
