package yamlp

import (
	"errors"
	"fmt"

	"github.com/heimdalr/dag"
	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

type Exports struct {
	files   map[string]*exportFile
	exports map[string]*Node
}

type exportFile struct {
	fileName string
	refs     []*Node
	exports  []*Node
}

func newExportFile(fileName string) *exportFile {
	return &exportFile{
		fileName: fileName,
		refs:     []*Node{},
		exports:  []*Node{},
	}
}

func (ef *exportFile) ID() string {
	return ef.fileName
}

func (ef *exportFile) HasExport() bool {
	return len(ef.exports) > 0
}

func (ef *exportFile) GetImports() []string {
	imports := []string{}
	for _, n := range ef.refs {
		imports = append(imports, n.GetImports()...)
	}
	for _, n := range ef.exports {
		imports = append(imports, n.GetImports()...)
	}

	return imports
}

func (e Exports) Push(n *Node) error {
	switch n.Kind {
	case Ref:
		return e.pushRef(n)
	case Export:
		return e.pushExport(n)
	}

	return errors.New("cannot push node to exports, must be ref or export kind")
}
func (e Exports) pushRef(n *Node) error {
	if _, exists := e.files[n.File]; !exists {
		e.files[n.File] = newExportFile(n.File)
	}

	if e.files[n.File].HasExport() {
		return errors.New("refs must be declared before exports")
	}

	e.files[n.File].refs = append(e.files[n.File].refs, n)

	return nil
}

func (e Exports) pushExport(n *Node) error {
	if _, exists := e.exports[n.Name]; exists {
		return fmt.Errorf("'%s' export already exists", n.Name)
	}
	if _, exists := e.files[n.File]; !exists {
		e.files[n.File] = newExportFile(n.File)
	}

	e.files[n.File].exports = append(e.files[n.File].exports, n)
	e.exports[n.Name] = n

	return nil
}
func (e Exports) getExports(names []string) map[string]*Node {
	exports := map[string]*Node{}

	for _, name := range names {
		//TODO validation?
		exports[name] = e.exports[name]
	}

	return exports
}
func (e Exports) resolve() (*ContextNode, error) {
	files, err := e.getResolveOrder()
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		var prevRef *ContextNode
		for _, ref := range e.files[file].refs {
			exports := e.getExports(ref.GetImports())
			err := ref.Resolve(prevRef, exports)
			if err != nil {
				return nil, err
			}

			prevRef = NewContextNode(ref.CandidateNode)
		}
		for _, export := range e.files[file].exports {
			exports := e.getExports(export.GetImports())
			err := export.Resolve(prevRef, exports)
			if err != nil {
				return nil, err
			}
		}
	}

	if dot, exists := e.exports["."]; exists {
		return NewContextNode(dot.CandidateNode), nil
	}

	return NewContextNode(createExportMapNode(e.exports)), nil
}

func (e Exports) getResolveOrder() ([]string, error) {
	d := dag.NewDAG()

	for _, ef := range e.files {
		dstId, _ := d.AddVertex(ef)

		for _, importName := range ef.GetImports() {
			depNode, exists := e.exports[importName]
			if !exists {
				return nil, fmt.Errorf("can't import '%s', does not exists", importName)
			}
			srcId, _ := d.AddVertex(e.files[depNode.File])
			err := d.AddEdge(srcId, dstId)
			if err != nil {
				// TODO detect error type (EdgeLoopError) and return custom error
				return nil, err
			}
		}
	}

	v := &refVisitor{}

	d.OrderedWalk(v)

	return v.Ids, nil
}

func createExportMapNode(refs map[string]*Node) *yqlib.CandidateNode {
	refMap := &yqlib.CandidateNode{
		Kind: yqlib.MappingNode,
	}

	for name, ref := range refs {
		refMap.AddKeyValueChild(&yqlib.CandidateNode{
			Kind:  yqlib.ScalarNode,
			Value: name,
		}, ref.CandidateNode)
	}

	return refMap
}

type refVisitor struct {
	Ids []string
}

func (pv *refVisitor) Visit(v dag.Vertexer) {
	id, _ := v.Vertex()
	pv.Ids = append(pv.Ids, id)
}
