package yamlp

import (
	"os"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

type Kind uint32

const (
	SequenceNode Kind = 1 << iota
	MappingNode
	ScalarNode
)

type Node struct {
	CandidateNode *yqlib.CandidateNode
	NodeContext   NodeContext
	resolveCount  int
}

type NodeContext struct {
	Dir string
}

func (n *Node) GetTagNodes() []*tagNode {
	return getTagNodes(n.CandidateNode)
}

func (n *Node) Resolve(refs map[string]*Node) error {
	n.resolveCount += 1
	
	for _, tn := range n.GetTagNodes() {
		nn, err := tagResolvers[tn.tag].Resolve(tn.candidateNode, n.NodeContext, refs)
		if err != nil {
			return err
		}

		*tn.candidateNode = *nn
	}

	return nil
}

func (n *Node) GetResolveCount() int {
	return n.resolveCount
}

func (n *Node) PrettyPrintYaml(w *os.File) {
	prefs := yqlib.NewDefaultYamlPreferences()
	prefs.UnwrapScalar = false
	prefs.ColorsEnabled = shouldColorize()
	prefs.Indent = 4
	printer := yqlib.NewPrinter(yqlib.NewYamlEncoder(prefs), yqlib.NewSinglePrinterWriter(w))

	list, err := yqlib.NewAllAtOnceEvaluator().EvaluateNodes(".", n.CandidateNode)
	if err != nil {
		panic(err)
	}
	printer.PrintResults(list)
}

func shouldColorize() bool {
	colorsEnabled := false
	fileInfo, _ := os.Stdout.Stat()

	if (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		colorsEnabled = true
	}

	return colorsEnabled
}

// func deepCopyNode(node *yqlib.CandidateNode, cache map[*yqlib.CandidateNode]*yqlib.CandidateNode) *yqlib.CandidateNode {
// 	if n, ok := cache[node]; ok {
// 		return n
// 	}
// 	if cache == nil {
// 		cache = make(map[*yqlib.CandidateNode]*yqlib.CandidateNode)
// 	}
// 	copy := *node
// 	cache[node] = &copy
// 	copy.Content = nil
// 	for _, elem := range node.Content {
// 		copy.Content = append(copy.Content, deepCopyNode(elem, cache))
// 	}
// 	if node.Alias != nil {
// 		copy.Alias = deepCopyNode(node.Alias, cache)
// 	}
// 	return &copy
// }

// func (n *Node) getNodesByTag(node *yqlib.CandidateNode, tag string, allowedKinds Kind) []*yaml.Node {
// 	if node.Kind == yaml.ScalarNode && node.Tag == tag {
// 		return []*yaml.Node{node}
// 	}

// 	scalarNodes := []*yaml.Node{}

// 	if node.Kind <= yaml.MappingNode {
// 		for _, n := range node.Content {
// 			scalarNodes = append(scalarNodes, getScalarNodesByTag(n, tag)...)
// 		}
// 	}

// 	return scalarNodes
// }
