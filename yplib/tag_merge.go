package yplib

import (
	"container/list"
	"errors"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

var mergeMapNewExpr *yqlib.ExpressionNode

func init() {
	AddTagResolver("merge", mergeResolver)
	mergeMapNewExpr, _ = yqlib.ExpressionParser.ParseExpression(`. *n $rhs`)

}

func mergeResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	if container := rc.Target.Parent.Parent; container != nil &&
		container.Kind == yqlib.SequenceNode &&
		len(rc.Target.Parent.Content) == 2 {
		err := seqMerge(rc.Target)

		return nil, err
	}

	err := mapMerge(rc.Target)

	return nil, err
}

func seqMerge(target *yqlib.CandidateNode) error {
	seqNode := target.Parent.Parent

	var newContent []*yqlib.CandidateNode
	var startIndex int

	for i, n := range seqNode.Content {
		if n.Kind != yqlib.MappingNode || len(n.Content) == 0 {
			continue
		}
		if n.Content[0].Value == target.Value {
			if n.Content[1].Kind != yqlib.SequenceNode {
				return errors.New("<< merge value must be a seq in this context")
			}

			newContent = n.Content[1].Content
			startIndex = i
			continue
		}
	}

	n := seqNode.CopyWithoutContent()

	n.Content = append(n.Content, seqNode.Content[:startIndex]...)
	n.Content = append(n.Content, newContent...)

	if startIndex+1 < len(seqNode.Content) {
		n.Content = append(n.Content, seqNode.Content[startIndex+1:]...)
	}

	seqNode.Content = n.Content

	return nil
}

func mapMerge(target *yqlib.CandidateNode) error {
	mapNode := target.Parent

	var newContent *yqlib.CandidateNode
	var startIndex int

	afterKeys := map[string]struct{}{}
	newPairs := map[string]*yqlib.CandidateNode{}

	for i := 0; i < len(mapNode.Content); i += 2 {
		k := mapNode.Content[i]
		v := mapNode.Content[i+1]

		if k.Value == target.Value {
			if v.Kind != yqlib.MappingNode {
				return errors.New("<< merge value must be a map")
			}

			newContent = v
			startIndex = i
			break
		}
	}

	if len(mapNode.Content) >= startIndex+2 {
		for i := startIndex + 2; i < len(mapNode.Content); i += 2 {
			afterKeys[mapNode.Content[i].Value] = struct{}{}
		}
		tmp := newContent.CopyWithoutContent()

		for i := 0; i < len(newContent.Content); i += 2 {
			k := newContent.Content[i]
			v := newContent.Content[i+1]

			if _, exists := afterKeys[k.Value]; !exists {
				tmp.Content = append(tmp.Content, k, v)
				newPairs[k.Value] = v
			}
		}

		newContent = tmp
	}

	n := mapNode.CopyWithoutContent()

	for i := 0; i < startIndex; i += 2 {
		k := mapNode.Content[i]
		v := mapNode.Content[i+1]

		if val, exists := newPairs[k.Value]; exists {
			v, err := mergeNewFields(v, val)
			if err != nil {
				return err
			}
			n.Content = append(n.Content, k, v)

			delete(newPairs, k.Value)
		} else {
			n.Content = append(n.Content, k, v)
		}
	}

	for i := 0; i < len(newContent.Content); i += 2 {
		k := newContent.Content[i]
		v := newContent.Content[i+1]

		if _, exists := newPairs[k.Value]; exists {
			n.Content = append(n.Content, k, v)
		}
	}

	n.Content = append(n.Content, mapNode.Content[startIndex+2:]...)

	mapNode.Content = n.Content

	return nil
}

func mergeNewFields(lhs, rhs *yqlib.CandidateNode) (*yqlib.CandidateNode, error) {
	if rhs.Kind != yqlib.MappingNode {
		return lhs, nil
	}
	context, err := yqlib.NewDataTreeNavigator().GetMatchingNodes(createMergeNewContext(lhs, rhs), mergeMapNewExpr)
	if err != nil {
		return nil, err
	}

	if context.MatchingNodes.Len() == 0 {
		return nil, errors.New("failed to merge")
	}

	return context.MatchingNodes.Front().Value.(*yqlib.CandidateNode), nil
}

func createMergeNewContext(lhs, rhs *yqlib.CandidateNode) yqlib.Context {
	lhsList := list.New()
	lhsList.PushBack(lhs)

	vars := map[string]*list.List{}
	rhsList := list.New()
	rhsList.PushBack(rhs)
	vars["rhs"] = rhsList

	return yqlib.Context{
		MatchingNodes: lhsList,
		Variables:     vars,
	}
}
