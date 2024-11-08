package yamlp

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	AddTagResolver("map", mapResolver)
}

func mapResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	tokens := strings.Split(strings.TrimSpace(rc.Target.Value), " ")
	if len(tokens) != 2 {
		return nil, errors.New("expected 2 expressions")
	}

	partial, err := yq(createContext(rc), tokens[0])
	if err != nil {
		return nil, err
	}
	data, err := yq(createContext(rc), tokens[1])
	if err != nil {
		return nil, err
	}

	partialNode := &Node{
		Dir:           rc.Node.Dir,
		File:          rc.Node.File,
		CandidateNode: partial,
		tagNodes:      getTagNodes(partial),
	}

	seq := newSeqNode()
	NewContextNode(data).Reduce(seq, func(vars map[string]*ContextNode) (*yqlib.CandidateNode, error) {
		clone := partialNode.Clone()
		err := clone.Resolve(rc.Ctx, vars)
		if err != nil {
			return nil, err
		}

		return clone.CandidateNode, nil
	}, updateSeqNode)

	return seq, nil
}

func yq(ctx yqlib.Context, expr string) (*yqlib.CandidateNode, error) {
	e, err := yqlib.ExpressionParser.ParseExpression(expr)
	if err != nil {
		return nil, err
	}

	context, err := yqlib.NewDataTreeNavigator().GetMatchingNodes(ctx, e)
	if err != nil {
		return nil, err
	}
	if context.MatchingNodes.Len() == 0 {
		return nil, errors.New("Unable to resolve")
	}

	node, ok := context.MatchingNodes.Front().Value.(*yqlib.CandidateNode)
	if !ok {
		return nil, fmt.Errorf("yq expression error (%s): failed to marshal CandidateNode", expr)
	}

	return node, nil
}
