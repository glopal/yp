package yamlp

import (
	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

var tagResolvers = map[string]TagResolver{}

type TagResolver struct {
	AllowedKind yqlib.Kind
	Resolve     ResolveFunc
}

type ResolveFunc func(*yqlib.CandidateNode, NodeContext, map[string]*Node) (*yqlib.CandidateNode, error)

func AddTagResolver(tag string, resolver ResolveFunc, allowedKinds ...yqlib.Kind) {
	var allowedKind yqlib.Kind

	for _, k := range allowedKinds {
		allowedKind |= k
	}
	if allowedKind == 0 {
		allowedKind = yqlib.ScalarNode
	}
	tagResolvers[tag] = TagResolver{
		AllowedKind: allowedKind,
		Resolve:     resolver,
	}
}

type tagNode struct {
	tag           string
	candidateNode *yqlib.CandidateNode
}

func getTagNodes(node *yqlib.CandidateNode) []*tagNode {
	if resolver, exists := tagResolvers[node.Tag]; exists && node.Kind&resolver.AllowedKind > 0 {
		return []*tagNode{{node.Tag, node}}
	}

	tagNodes := []*tagNode{}

	if node.Kind <= yqlib.MappingNode {
		for _, n := range node.Content {
			tagNodes = append(tagNodes, getTagNodes(n)...)
		}
	}

	return tagNodes
}
