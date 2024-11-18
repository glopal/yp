package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type Coder[T any] interface {
	Decode([]byte) error
	Encode() ([]byte, error)
	IsDir() bool
}

type VFS[T Coder[T]] struct {
	*orderedmap.OrderedMap[string, T]
}

type VFD struct {
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Content  *string                `yaml:"content,omitempty"`
}

func (vfd *VFD) IsDir() bool {
	return vfd.Content == nil
}

func NewVFS[T Coder[T]]() *VFS[T] {
	return &VFS[T]{orderedmap.New[string, T]()}
}

type TestSuiteFs struct {
	input    *VFS[*VFS[Content]]
	expected *VFS[*VFS[Content]]
}

func NewTestSuiteFs() *TestSuiteFs {
	return &TestSuiteFs{
		input:    NewVFS[*VFS[Content]](),
		expected: NewVFS[*VFS[Content]](),
	}
}

func (vfs *VFS[T]) Decode([]byte) error {
	return nil
}
func (vfs *VFS[T]) Encode() ([]byte, error) {
	return nil, nil
}
func (vfs *VFS[T]) IsDir() bool {
	return vfs.OrderedMap == nil
}

func (vfs *VFS[T]) Push(newKey string, val T) (T, bool) {
	if item, exists := vfs.Get(newKey); exists {
		return item, false
	}

	afterKey := ""

	for curKey := range vfs.KeysFromOldest() {
		val := strings.Compare(curKey, newKey)

		if val > 0 { // curKey > newKey
			break
		} else if val < 0 { // curKey < newKey
			afterKey = curKey
		}
	}

	v, ok := vfs.Set(newKey, val)
	if !ok {
		return v, false
	}

	var err error
	if afterKey == "" {
		err = vfs.MoveToFront(newKey)
	} else {
		err = vfs.MoveAfter(newKey, afterKey)
	}
	if err != nil {
		return *new(T), false
	}

	return v, true
}

func (vfs *VFS[T]) ToYaml() ([]byte, error) {
	return yaml.Marshal(vfs)
}

func (vfs *VFS[T]) UnmarshalYaml(file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, vfs)
}

func (vfs *VFS[T]) UnmarshalYamlString(data string) error {
	return yaml.Unmarshal([]byte(data), vfs)
}

func (vfs *VFS[T]) UnmarshalDir(dir string, trim int) error {
	cleanPath := filepath.Clean(dir)
	pathTokens := strings.Split(cleanPath, "/")
	if trim > len(pathTokens) {
		trim = len(pathTokens)
	}
	prefix := strings.Join(pathTokens[:trim], "/")

	if trimmedBaseTokens := strings.Split(strings.TrimPrefix(strings.TrimPrefix(dir, prefix), "/"), "/"); len(trimmedBaseTokens) > 1 {
		increment := ""
		for _, dirItem := range trimmedBaseTokens[:len(trimmedBaseTokens)-1] {
			if increment != "" {
				increment += "/"
			}
			increment += dirItem

			vfs.Set(increment, *new(T))
		}
	}

	err := filepath.WalkDir(filepath.Clean(dir), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		vPath := strings.TrimPrefix(strings.TrimPrefix(path, prefix), "/")

		if vPath == "" {
			return nil
		}

		if !d.IsDir() {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var item T

			err = item.Decode(data)
			if err != nil {
				return err
			}

			vfs.Set(vPath, item)
		} else {
			vfs.Set(vPath, *new(T))
		}

		return nil
	})

	return err
}

// func (vfs VFS) ToMemMapFs() (afero.Fs, error) {
// 	afs := afero.NewMemMapFs()

// 	for path, vfd := range vfs.omap.FromOldest() {
// 		if vfd.IsDir() {
// 			err := afs.MkdirAll(path, 0755)
// 			if err != nil {
// 				return nil, err
// 			}
// 		} else {
// 			if err := afero.WriteFile(afs, path, []byte(*vfd.Content), 0755); err != nil {
// 				return nil, err
// 			}
// 		}
// 	}

// 	return afs, nil
// }

// func (vfs VFS) ToJtreeMap() (*orderedmap.OrderedMap[string, *JtreeNode], error) {
// 	tree, err := vfs.ToJtree()
// 	if err != nil {
// 		return nil, err
// 	}

// 	treeMap := orderedmap.New[string, *JtreeNode]()

// 	for _, node := range tree {
// 		treeMap.Set(node.Id, node)
// 	}

// 	return treeMap, nil
// }
// func (vfs VFS) ToJtree() (Jtree, error) {
// 	tree := Jtree{}

// 	for path, vfd := range vfs.omap.FromOldest() {
// 		parent := "#"
// 		pathTokens := strings.Split(path, "/")
// 		if len(pathTokens) > 1 {
// 			parent = strings.Join(pathTokens[:len(pathTokens)-1], "/")
// 		}

// 		nodeType := "dir"
// 		content := ""
// 		if !vfd.IsDir() {
// 			nodeType = "file"
// 			content = *vfd.Content
// 		}

// 		tree = append(tree, &JtreeNode{
// 			Id:     path,
// 			Type:   nodeType,
// 			Text:   filepath.Base(path),
// 			Parent: parent,
// 			Data: NodeData{
// 				Metadata: vfd.Metadata,
// 				Content:  content,
// 			},
// 		})
// 	}

// 	return tree, nil
// }

// func (t Jtree) ContentToTree() Jtree {
// 	for _, node := range t {
// 		if node.Type == "file" {
// 			vfs := NewVFS()
// 			err := vfs.UnmarshalYamlString(node.Data.Content)
// 			if err != nil {
// 				fmt.Println(err)
// 				continue
// 			}

// 			nt, err := vfs.ToJtreeMap()
// 			if err != nil {
// 				fmt.Println(err)
// 				continue
// 			}

// 			node.Data.Tree = nt
// 			// node.Data.Content = ""
// 		}
// 	}

// 	return t
// }

// func (t Jtree) ToVFS() (VFS, error) {
// 	vfs := NewVFS()

// 	for _, n := range t {
// 		if n.Type == "dir" {
// 			vfs.Set(n.Id, nil)
// 			continue
// 		}

// 		content := n.Data.Content

// 		if n.Data.Tree != nil {
// 			subfs, err := n.Data.Jtree().ToVFS()
// 			if err != nil {
// 				return vfs, err
// 			}

// 			data, err := subfs.ToYaml()
// 			if err != nil {
// 				return vfs, err
// 			}
// 			content = string(data)
// 		}
// 		vfs.Set(n.Id, &content)
// 	}

// 	return vfs, nil
// }

// type Jtree []*JtreeNode

// func (jt Jtree) Json() ([]byte, error) {
// 	return json.Marshal(jt)
// }

// type JtreeNode struct {
// 	Id     string `json:"id"`
// 	Type   string `json:"type"`
// 	Parent string `json:"parent"`
// 	// Metadata map[string]interface{} `json:"metadata,omitempty"`
// 	// Content  string                 `json:"content,omitempty"`
// 	Data NodeData `json:"data,omitempty"`
// 	Text string   `json:"text"`
// }

// type NodeData struct {
// 	Tree     *orderedmap.OrderedMap[string, *JtreeNode] `json:"tree,omitempty"`
// 	Content  string                                     `json:"content,omitempty"`
// 	Metadata map[string]interface{}                     `json:"metadata,omitempty"`
// }

// func (nd NodeData) Jtree() Jtree {
// 	tree := Jtree{}

// 	for _, n := range nd.Tree.FromOldest() {
// 		tree = append(tree, n)
// 	}

// 	return tree
// }

// // Alternative format of the node (id & parent are required)
// // {
// // 	id          : "string" // required
// // 	parent      : "string" // required
// // 	text        : "string" // node text
// // 	icon        : "string" // string for custom
// // 	state       : {
// // 	  opened    : boolean  // is the node open
// // 	  disabled  : boolean  // is the node disabled
// // 	  selected  : boolean  // is the node selected
// // 	},
// // 	li_attr     : {}  // attributes for the generated LI node
// // 	a_attr      : {}  // attributes for the generated A node
// //   }
