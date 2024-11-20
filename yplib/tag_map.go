package yplib

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	AddTagResolver("map", mapResolver, yqlib.ScalarNode, yqlib.SequenceNode)
}

func mapResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	data, partial, err := getMapData(rc)
	if err != nil {
		return nil, err
	}

	partialNode := &Node{
		Dir:           rc.Node.Dir,
		File:          rc.Node.File,
		CandidateNode: partial,
		tagNodes:      getTagNodes(partial),
	}

	nn, err := NewContextNode(data).Reduce(partial.Kind, func(vars map[string]*ContextNode) (*yqlib.CandidateNode, error) {
		clone := partialNode.Clone()
		err := clone.Resolve(rc.Ctx, vars, rc.Opts)
		if err != nil {
			return nil, err
		}

		return clone.CandidateNode, nil
	})
	if err != nil {
		return nil, err
	}

	if rc.Target.IsMapKey {
		*rc.Target.Parent = *nn
		return nil, nil
	}

	return nn, nil
}

func getMapData(rc ResolveContext) (*yqlib.CandidateNode, *yqlib.CandidateNode, error) {
	if rc.Target.IsMapKey {
		var err error
		data := rc.Target
		if rc.Target.Kind == yqlib.ScalarNode {
			data, err = yq(createContext(rc), data.Value)
			if err != nil {
				return nil, nil, err
			}
		}

		partial := getMapKeyValue(rc.Target)
		if partial.Kind == yqlib.ScalarNode {
			partial, err = yq(createContext(rc), partial.Value)
			if err != nil {
				return nil, nil, err
			}
		}

		return data, partial, nil
	}

	tokens := strings.Split(strings.TrimSpace(rc.Target.Value), " ")
	if len(tokens) != 2 {
		return nil, nil, errors.New("expected 2 expressions")
	}

	data, err := yq(createContext(rc), tokens[0])
	if err != nil {
		return nil, nil, err
	}

	partial, err := yq(createContext(rc), tokens[1])
	if err != nil {
		return nil, nil, err
	}

	return data, partial, nil
}

func getMapKeyValue(key *yqlib.CandidateNode) *yqlib.CandidateNode {
	for i, n := range key.Parent.Content {
		if n.Value == key.Value {
			return key.Parent.Content[i+1]
		}
	}

	return nil
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
