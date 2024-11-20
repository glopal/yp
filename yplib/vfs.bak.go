package yplib

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/fs"
// 	"os"
// 	"path/filepath"
// 	"strings"

// 	"github.com/spf13/afero"
// 	orderedmap "github.com/wk8/go-ordered-map/v2"
// 	"gopkg.in/yaml.v3"
// )

// type VFS struct {
// 	omap *orderedmap.OrderedMap[string, *VFD]
// }

// type VFD struct {
// 	Metadata map[string]interface{} `json:"metadata,omitempty"`
// 	Content  *string                `yaml:"content,omitempty"`
// }

// func (vfd *VFD) IsDir() bool {
// 	return vfd.Content == nil
// }

// func NewVFS() VFS {
// 	return VFS{
// 		omap: orderedmap.New[string, *VFD](),
// 	}
// }

// func (vfs VFS) Get(key string) (*VFD, bool) {
// 	return vfs.omap.Get(key)
// }

// func (vfs VFS) Set(key string, content *string) (*VFD, bool) {
// 	return vfs.omap.Set(key, &VFD{
// 		Content: content,
// 	})
// }

// func (vfs VFS) Push(newKey string, content *string) (*VFD, bool) {
// 	if vfd, exists := vfs.Get(newKey); exists {
// 		return vfd, false
// 	}

// 	afterKey := ""

// 	for curKey := range vfs.omap.KeysFromOldest() {
// 		val := strings.Compare(curKey, newKey)

// 		if val > 0 { // curKey > newKey
// 			break
// 		} else if val < 0 { // curKey < newKey
// 			afterKey = curKey
// 		}
// 	}

// 	vfd, ok := vfs.omap.Set(newKey, &VFD{
// 		Content: content,
// 	})
// 	if !ok {
// 		return nil, false
// 	}

// 	var err error
// 	if afterKey == "" {
// 		err = vfs.omap.MoveToFront(newKey)
// 	} else {
// 		err = vfs.omap.MoveAfter(newKey, afterKey)
// 	}
// 	if err != nil {
// 		return nil, false
// 	}

// 	return vfd, true
// }

// func (vfs VFS) ToYaml() ([]byte, error) {
// 	return yaml.Marshal(vfs.omap)
// }

// func (vfs VFS) UnmarshalYaml(file string) error {
// 	data, err := os.ReadFile(file)
// 	if err != nil {
// 		return err
// 	}

// 	return yaml.Unmarshal(data, vfs.omap)
// }

// func (vfs VFS) UnmarshalYamlString(data string) error {
// 	return yaml.Unmarshal([]byte(data), vfs.omap)
// }

// func (vfs VFS) UnmarshalDir(dir string, trim int) error {
// 	cleanPath := filepath.Clean(dir)
// 	pathTokens := strings.Split(cleanPath, "/")
// 	if trim > len(pathTokens) {
// 		trim = len(pathTokens)
// 	}
// 	prefix := strings.Join(pathTokens[:trim], "/")

// 	if trimmedBaseTokens := strings.Split(strings.TrimPrefix(strings.TrimPrefix(dir, prefix), "/"), "/"); len(trimmedBaseTokens) > 1 {
// 		increment := ""
// 		for _, dirItem := range trimmedBaseTokens[:len(trimmedBaseTokens)-1] {
// 			if increment != "" {
// 				increment += "/"
// 			}
// 			increment += dirItem

// 			vfs.omap.Set(increment, &VFD{})
// 		}
// 	}

// 	err := filepath.WalkDir(filepath.Clean(dir), func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		vPath := strings.TrimPrefix(strings.TrimPrefix(path, prefix), "/")

// 		if vPath == "" {
// 			return nil
// 		}

// 		vfd := &VFD{}
// 		if !d.IsDir() {
// 			data, err := os.ReadFile(path)
// 			if err != nil {
// 				return err
// 			}
// 			content := string(data)
// 			vfd.Content = &content
// 		}

// 		vfs.omap.Set(vPath, vfd)

// 		return nil
// 	})

// 	return err
// }

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
