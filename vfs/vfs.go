package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type WriteThruVFS[T Coder[T]] interface {
	CreateDir(string) error
	Rename(string, string) error
	Write(T, string) error
	Delete(string) error
}

type Coder[T any] interface {
	New() T
	Decode([]byte) (T, error)
	Encode() ([]byte, error)
	IsDir() bool
}

type VFS[T Coder[T]] struct {
	*orderedmap.OrderedMap[string, T]
}

func NewVFS[T Coder[T]]() *VFS[T] {
	return &VFS[T]{orderedmap.New[string, T]()}
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

func (vfs *VFS[T]) Exists(key string) bool {
	_, exists := vfs.Get(key)
	return exists
}

func getParentKey(key string) string {
	pathItems := strings.Split(key, "/")
	if len(pathItems) <= 1 {
		return ""
	}

	return strings.Join(pathItems[:len(pathItems)-1], "/")
}
func (vfs *VFS[T]) AtOrOldest(key string) *orderedmap.Pair[string, T] {
	pair := vfs.AfterOrOldest(key)
	prev := pair.Prev()

	if prev != nil {
		return prev
	}

	return pair
}

func (vfs *VFS[T]) AfterOrOldest(key string) *orderedmap.Pair[string, T] {
	if key == "" {
		return vfs.Oldest()
	}

	pair := vfs.GetPair(key)
	if pair == nil {
		return vfs.Oldest()
	}

	return pair.Next()
}

func (vfs *VFS[T]) Push(newKey string, val T) bool {
	if vfs.Exists(newKey) {
		return true
	}

	// recursively create parent directories if they don't exist
	parentKey := getParentKey(newKey)
	if parentKey != "" && !vfs.Exists(parentKey) {
		vfs.Push(parentKey, *new(T))
	}

	for cur := vfs.AfterOrOldest(parentKey); cur != nil; cur = cur.Next() {
		val := strings.Compare(cur.Key, newKey)

		if val > 0 { // curKey > newKey
			break
		} else if val < 0 { // curKey < newKey
			parentKey = cur.Key
		}
	}

	vfs.Set(newKey, val)

	if parentKey == "" {
		vfs.MoveToFront(newKey)
	} else {
		vfs.MoveAfter(newKey, parentKey)
	}

	return false
}

func (vfs *VFS[T]) Delete(key string) {
	pair := vfs.GetPair(key)
	if pair == nil {
		return
	}

	keyToDeletes := []string{key}

	for cur := pair.Next(); cur != nil; cur = cur.Next() {
		if !strings.HasPrefix(cur.Key, key) {
			break
		}

		keyToDeletes = append(keyToDeletes, cur.Key)
	}

	for _, k := range keyToDeletes {
		vfs.OrderedMap.Delete(k)
	}
}

func (vfs *VFS[T]) Rename(orgKey string, newKey string) error {
	if vfs.Exists(newKey) {
		return fmt.Errorf("can't rename '%s' to existing key '%s'", orgKey, newKey)
	}

	pair := vfs.GetPair(orgKey)
	if pair == nil {
		return fmt.Errorf("'%s' does not exists", orgKey)
	}

	deletes := []string{orgKey}
	newPairs := []*orderedmap.Pair[string, T]{{
		Key:   newKey,
		Value: pair.Value,
	}}

	if pair.Value.IsDir() {
		for cur := pair.Next(); cur != nil; cur = cur.Next() {
			if strings.HasPrefix(cur.Key, orgKey) {
				deletes = append(deletes, cur.Key)
				newPairs = append(newPairs, &orderedmap.Pair[string, T]{
					Key:   newKey + strings.TrimPrefix(cur.Key, orgKey),
					Value: cur.Value,
				})
			}
		}
	}

	for _, d := range deletes {
		vfs.OrderedMap.Delete(d)
	}

	for _, p := range newPairs {
		vfs.Push(p.Key, p.Value)
	}

	return nil
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

			item := *new(T)
			item = item.New()

			err = yaml.Unmarshal(data, &item)
			// item, err = item.Decode(data)
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
