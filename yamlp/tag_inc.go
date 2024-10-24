package yamlp

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	AddTagResolver("!inc/file", incFileResolver)
	AddTagResolver("!inc/file/flatten", incFileFlattenResolver)
	AddTagResolver("!inc/files", incFilesResolver, yqlib.SequenceNode)
}

func incFileResolver(n *yqlib.CandidateNode, nc NodeContext, refs map[string]*Node) (*yqlib.CandidateNode, error) {
	return resolveFile(nc.Dir, n.Value)
}

func incFileFlattenResolver(n *yqlib.CandidateNode, nc NodeContext, refs map[string]*Node) (*yqlib.CandidateNode, error) {
	if n.Parent.Kind != yqlib.SequenceNode {
		return nil, errors.New("!inc/file/flatten must be used inside a sequence")
	}

	index, _ := strconv.Atoi(n.Key.Value)

	nn, err := resolveFile(nc.Dir, n.Value)
	if err != nil {
		return nil, err
	}

	if nn.Kind != yqlib.SequenceNode {
		return nil, errors.New("!inc/file/flatten must return a sequence")
	}

	contents := make([]*yqlib.CandidateNode, 0, len(n.Parent.Content)+len(nn.Content)-1)
	contents = append(contents, n.Parent.Content[:index]...)
	contents = append(contents, nn.Content...)
	contents = append(contents, n.Parent.Content[index+1:]...)

	for i, c := range contents {
		c.Key.Value = fmt.Sprintf("%v", i)
	}

	n.Parent.Content = contents
	return nn.Content[0], nil

}

func incFilesResolver(n *yqlib.CandidateNode, nc NodeContext, refs map[string]*Node) (*yqlib.CandidateNode, error) {
	for i, cn := range n.Content {
		if cn.Tag != "!!str" {
			return nil, fmt.Errorf("!!inc/files[%d] is not !!str (%s)", i, cn.Value)
		}

		nn, err := resolveFile(nc.Dir, cn.Value)
		if err != nil {
			return nil, err
		}

		n.Content[i] = nn

	}

	n.Style = 0
	n.Tag = "!!seq"

	return n, nil
}

func resolveFile(dir, relPath string) (*yqlib.CandidateNode, error) {
	var err error
	ymlFilePath := relPath

	if !filepath.IsAbs(ymlFilePath) {
		ymlFilePath = filepath.Join(dir, ymlFilePath)
	}

	ns, err := LoadFile(ymlFilePath)
	if err != nil {
		return nil, err
	}

	nodes := ns.Nodes()

	if len(nodes) > 1 {
		return nil, fmt.Errorf("multi-doc yaml files cannot be included (%s)", ymlFilePath)
	}

	if len(nodes) == 0 {
		return &yqlib.CandidateNode{
			Kind:  yqlib.ScalarNode,
			Style: yqlib.DoubleQuotedStyle,
			Tag:   "!!str",
		}, nil
	}

	return nodes[0].CandidateNode, nil
}
