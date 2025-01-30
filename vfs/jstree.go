package vfs

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

type JsTree []JsTreeNode

type JsTreeNode struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Parent string `json:"parent"`
	Data   any    `json:"data,omitempty"`
	Text   string `json:"text"`
}

func (vfs *VFS[T]) ToJsTreeMap() map[string]JsTreeNode {
	tree := vfs.ToJsTree()
	treeMap := make(map[string]JsTreeNode, len(tree))

	for _, node := range tree {
		treeMap[node.Id] = node
	}

	return treeMap
}

func (vfs *VFS[T]) ToJsTree() JsTree {
	tree := make(JsTree, 0, vfs.Len())

	for path, fd := range vfs.FromOldest() {
		parent := "#"
		pathTokens := strings.Split(path, "/")
		if len(pathTokens) > 1 {
			parent = strings.Join(pathTokens[:len(pathTokens)-1], "/")
		}

		node := JsTreeNode{
			Id:     path,
			Type:   "dir",
			Text:   filepath.Base(path),
			Parent: parent,
		}
		if !fd.IsDir() {
			node.Type = "file"
			node.Data = fd.content

			vfs.DecorateFileNode(&node)
		}

		tree = append(tree, node)
	}

	return tree
}

func (vfs *VFS[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(vfs.ToJsTree())
}
