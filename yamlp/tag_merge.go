package yamlp

import (
	"errors"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	AddTagResolver("merge", mergeResolver)
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
	n.Content = append(n.Content, seqNode.Content[startIndex+2:]...)

	seqNode.Content = n.Content

	return nil
}

func mapMerge(target *yqlib.CandidateNode) error {
	mapNode := target.Parent

	var newContent *yqlib.CandidateNode
	var startIndex int

	for i := 0; i < len(mapNode.Content); i += 2 {
		k := mapNode.Content[i]
		v := mapNode.Content[i+1]

		if k.Value == target.Value {
			if v.Kind != yqlib.MappingNode {
				return errors.New("<< merge value must be a map")
			}

			newContent = v
			startIndex = i
			continue
		}
	}

	n := mapNode.CopyWithoutContent()

	n.Content = append(n.Content, mapNode.Content[:startIndex]...)
	n.Content = append(n.Content, newContent.Content...)
	n.Content = append(n.Content, mapNode.Content[startIndex+2:]...)

	mapNode.Content = n.Content

	return nil
}
