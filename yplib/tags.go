package yplib

import (
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

var tagResolvers = map[string]TagResolver{}

type TagResolver struct {
	AllowedKind yqlib.Kind
	Resolve     ResolveFunc
}

type ResolveFunc func(ResolveContext) (*yqlib.CandidateNode, error)

type ResolveContext struct {
	Target *yqlib.CandidateNode
	Ctx    *ContextNode
	Node   *Node
	Vars   map[string]*ContextNode
	Opts   *loadOptions
}

func AddTagResolver(tag string, resolver ResolveFunc, allowedKinds ...yqlib.Kind) {
	var allowedKind yqlib.Kind

	for _, k := range allowedKinds {
		allowedKind |= k
	}
	if allowedKind == 0 {
		allowedKind = yqlib.ScalarNode
	}
	tagResolvers[cleanTag(tag)] = TagResolver{
		AllowedKind: allowedKind,
		Resolve:     resolver,
	}
}

type tagNode struct {
	tag           string
	candidateNode *yqlib.CandidateNode
}

func getTagNodes(node *yqlib.CandidateNode) []*tagNode {
	tag := cleanTag(node.Tag)
	if resolver, exists := tagResolvers[tag]; exists && node.Kind&resolver.AllowedKind > 0 {
		return []*tagNode{{tag, node}}
	}

	tagNodes := []*tagNode{}

	if node.Kind <= yqlib.MappingNode {
		skip := false
		for _, n := range node.Content {
			if skip {
				skip = false
				continue
			}
			if cleanTag(n.Tag) == "_" {
				if n.IsMapKey {
					if n.Value == "<<" {
						n.Tag = "!!merge"
					}

				}
				skip = true
				continue
			}
			tagNodes = append(tagNodes, getTagNodes(n)...)
		}
	}

	return tagNodes
}

func cleanTag(tag string) string {
	return strings.TrimLeft(tag, "!")
}

func CreateScalarNode(tag, value string) *yqlib.CandidateNode {
	return &yqlib.CandidateNode{
		Kind:  yqlib.ScalarNode,
		Tag:   tag,
		Value: value,
	}
}

func CreateNullNode() *yqlib.CandidateNode {
	return CreateScalarNode("!!null", "null")
}
