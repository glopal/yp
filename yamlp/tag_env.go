package yamlp

import (
	"errors"
	"fmt"
	"os"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	AddTagResolver("env", envResolver, yqlib.ScalarNode)
	AddTagResolver("env/map", envMapResolver, yqlib.SequenceNode, yqlib.ScalarNode)
}

func envResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	return &yqlib.CandidateNode{
		Kind:  yqlib.ScalarNode,
		Tag:   "!!str",
		Value: os.Getenv(rc.Target.Value),
	}, nil
}

func envMapResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	if rc.Target.IsMapKey {
		return envMapKeyResolver(rc)
	}

	n := &yqlib.CandidateNode{
		Kind: yqlib.MappingNode,
	}

	for _, item := range rc.Target.Content {
		k, v, err := getEnv(item)
		if err != nil {
			return nil, err
		}
		n.AddKeyValueChild(&yqlib.CandidateNode{
			Kind:  yqlib.ScalarNode,
			Tag:   "!!str",
			Value: k,
		}, &yqlib.CandidateNode{
			Kind:  yqlib.ScalarNode,
			Tag:   "!!str",
			Value: v,
		})
	}

	return n, nil
}

func envMapKeyResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	mapNode := rc.Target.Parent

	exisitingKeyIndices := map[string]int{}
	var envVars *yqlib.CandidateNode
	var envVarsIndex int

	for i := 0; i < len(mapNode.Content); i += 2 {
		k := mapNode.Content[i]
		v := mapNode.Content[i+1]

		if k.Value == rc.Target.Value {
			if v.Kind != yqlib.SequenceNode {
				return nil, errors.New("!env/map when used as map key must have a sequence value")
			}

			envVars = v
			envVarsIndex = i
			continue
		}

		exisitingKeyIndices[k.Value] = i
	}

	n := mapNode.CopyWithoutContent()

	n.Content = append(n.Content, mapNode.Content[:envVarsIndex]...)

	for _, envVar := range envVars.Content {
		if envVar.Kind != yqlib.ScalarNode {
			return nil, errors.New("!env/map values must be scalar")
		}

		k, v, err := getEnv(envVar)
		if err != nil {
			return nil, err
		}

		if i, exists := exisitingKeyIndices[k]; exists {
			if v != "" {
				mapNode.Content[i+1].Value = v
			}
			continue
		}
		n.AddKeyValueChild(&yqlib.CandidateNode{
			Kind:  yqlib.ScalarNode,
			Tag:   "!!str",
			Value: k,
		}, &yqlib.CandidateNode{
			Kind:  yqlib.ScalarNode,
			Tag:   "!!str",
			Value: v,
		})
	}
	n.Content = append(n.Content, mapNode.Content[envVarsIndex+2:]...)

	mapNode.Content = n.Content

	return nil, nil
}

func getEnv(n *yqlib.CandidateNode) (string, string, error) {
	k := n.Value
	v := os.Getenv(k)
	var err error
	if v == "" && cleanTag(n.Tag) == "must" {
		err = fmt.Errorf("required environment variable '%s' not set", k)
	}

	return k, v, err
}
