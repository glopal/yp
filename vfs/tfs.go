package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type TestFs struct {
	Input    *VFS[*Content] `yaml:"input"`
	Expected *VFS[*Content] `yaml:"expected"`
}

func NewTestFs() *TestFs {
	return &TestFs{
		Input:    NewVFS[*Content](),
		Expected: NewVFS[*Content](),
	}
}
func (t *TestFs) New() *TestFs {
	return NewTestFs()
}

// func (t *TestFs) UnmarshalYAML(node *yaml.Node) error {
// 	t = NewTestFs()

// 	return nil
// }

func (t *TestFs) Decode(data []byte) (*TestFs, error) {
	nt := NewTestFs()

	return nt, yaml.Unmarshal(data, nt)
}
func (t *TestFs) Encode() ([]byte, error) {
	return yaml.Marshal(t)
}
func (t *TestFs) IsDir() bool {
	return t.Input == nil
}

type TestSuiteFs struct {
	*VFS[*TestFs]
	root string
}

func NewTestSuiteFs(dir string) (*TestSuiteFs, error) {
	tfs := NewVFS[*TestFs]()

	err := tfs.UnmarshalDir(dir, len(strings.Split(dir, "/")))
	if err != nil {
		return nil, err
	}
	return &TestSuiteFs{
		tfs,
		dir,
	}, nil
}

func (ts *TestSuiteFs) CreateDir(path string) error {
	if err := os.MkdirAll(ts.path(path), 0755); err != nil {
		return err
	}

	ts.Push(path, &TestFs{})

	return nil
}

func (ts *TestSuiteFs) Rename(oldPath, newPath string) error {
	if err := os.Rename(ts.path(oldPath), ts.path(newPath)); err != nil {
		return err
	}

	return ts.VFS.Rename(oldPath, newPath)
}

func (ts *TestSuiteFs) Write(item *TestFs, path string) error {
	data, err := item.Encode()
	if err != nil {
		return err
	}

	parent := getParentKey(path)
	if !ts.Exists(parent) {
		err = ts.CreateDir(parent)
		if err != nil {
			return err
		}
	}

	if err := os.WriteFile(ts.path(path), data, 0755); err != nil {
		return err
	}

	ts.Push(path, item)

	return nil
}

func (ts *TestSuiteFs) WriteInput(testPath, inputPath, content string) error {
	t, exists := ts.Get(testPath)
	if !exists {
		t = NewTestFs()
	}

	c := &Content{&content}
	exists = t.Input.Push(inputPath, c)
	if exists {
		t.Input.Set(inputPath, c)
	}

	return ts.Write(t, testPath)
}

func (ts *TestSuiteFs) Delete(path string) error {
	if err := os.RemoveAll(ts.path(path)); err != nil {
		return err
	}

	ts.VFS.Delete(path)

	return nil
}

func (ts *TestSuiteFs) path(path string) string {
	return filepath.Join(ts.root, path)
}

type FD[T any] struct {
	IsDir   bool
	Content T
}

func (fd *FD[T]) UnmarshalYAML(node *yaml.Node) (err error) {
	return nil
}

type Content struct {
	*string
}

func (c *Content) UnmarshalYAML(node *yaml.Node) (err error) {
	fmt.Println(node.Kind, node.Tag, node.Value)
	return nil
}

func (c *Content) MarshalText() (text []byte, err error) {
	return []byte(*c.string), nil
}

// func (c *Content) MarshalYAML() (interface{}, error) {
// 	fmt.Println("MARSHAL: ", c)
// 	if c.string != nil {
// 		node := yaml.Node{
// 			Kind:  yaml.ScalarNode,
// 			Tag:   "!!str",
// 			Value: "TEST",
// 		}
// 		return &node, nil
// 	}

// 	return nil, nil
// }

func (c *Content) New() *Content {
	return &Content{}
}
func (c *Content) Decode(b []byte) (*Content, error) {
	str := string(b)

	return &Content{&str}, nil
}
func (c *Content) Encode() ([]byte, error) {
	return nil, nil
}
func (c *Content) IsDir() bool {
	return c == nil
}
