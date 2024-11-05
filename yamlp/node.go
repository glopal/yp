package yamlp

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

type Kind uint32

const (
	SequenceNode Kind = 1 << iota
	MappingNode
	ScalarNode
)

type Node struct {
	Dir           string
	File          string
	Name          string
	Kind          DocKind
	CandidateNode *yqlib.CandidateNode
	decoded       interface{}
	resolveCount  int
	resolved      bool
	tagNodes      []*tagNode
}

func NewNode(cn *yqlib.CandidateNode, file string) *Node {
	n := &Node{
		Dir:           filepath.Dir(file),
		File:          file,
		Name:          strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)),
		CandidateNode: cn,
		tagNodes:      getTagNodes(cn),
	}

	doc := determineDoc(cn)
	n.Kind = doc.Kind

	if n.Kind&(Ref|Export) > 0 {
		n.Name = doc.Val
	}

	return n
}

func (n *Node) Resolve(ctx *ContextNode, exports map[string]*Node) error {
	n.resolveCount += 1

	for _, tn := range n.tagNodes {
		nn, err := tagResolvers[tn.tag].Resolve(ResolveContext{
			Target:  tn.candidateNode,
			Ctx:     ctx,
			Node:    n,
			Imports: exports,
		})
		if err != nil {
			return err
		}

		if nn != nil {
			*tn.candidateNode = *nn
		}
	}

	n.resolved = true

	return nil
}

func (n *Node) GetImports() []string {
	imports := []string{}
	for _, tn := range n.tagNodes {
		if tn.tag == "import" {
			imports = append(imports, tn.candidateNode.Value)
		}
	}

	return imports
}

func (n *Node) Interface() (interface{}, error) {
	if n.decoded != nil {
		return n.decoded, nil
	}

	yn, err := n.CandidateNode.MarshalYAML()
	if err != nil {
		return nil, err
	}
	var i interface{}
	err = yn.Decode(&i)
	if err != nil {
		return nil, err
	}

	n.decoded = i

	return i, nil
}

func (n *Node) IsRef() bool {
	return n.Kind == Ref
}
func (n *Node) IsExport() bool {
	return n.Kind == Export
}
func (n *Node) IsRefOrExport() bool {
	return n.Kind&(Ref|Export) > 0
}

func (n *Node) IsResolved() bool {
	return n.resolved
}

func (n *Node) ID() string {
	return n.Kind.String() + "/" + n.Name
}

func (n *Node) GetResolveCount() int {
	return n.resolveCount
}

func (n *Node) PrettyPrintYaml(w io.Writer) {
	prefs := yqlib.NewDefaultYamlPreferences()
	prefs.UnwrapScalar = false
	prefs.ColorsEnabled = shouldColorize()
	prefs.Indent = 2
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
