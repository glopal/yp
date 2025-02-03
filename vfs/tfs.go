package vfs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type TestFs struct {
	Input     *VFS[string] `yaml:"input"`
	Output    *VFS[string] `yaml:"output"`
	syncSuite func() error
}

func NewTestFs() (*TestFs, error) {
	t := &TestFs{
		Input:  NewVFS[string](),
		Output: NewVFS[string](),
	}

	return t, t.Input.InitMemMapFs()
}
func (t *TestFs) SetSyncHook(syncHook func() error) {
	t.syncSuite = syncHook

	t.Input.OnPushDir = func(path string) error {
		err := t.Input.Fs.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
		return syncHook()
	}
	t.Input.OnPush = func(p *orderedmap.Pair[string, FD[string]]) error {
		fmt.Println(t.Input)
		err := afero.WriteFile(t.Input.Fs, p.Key, []byte(p.Value.content), 0755)
		if err != nil {
			return err
		}

		return syncHook()
	}
	t.Input.OnRename = func(oldPath, newPath string) error {
		err := t.Input.Fs.Rename(oldPath, newPath)
		if err != nil {
			return err
		}

		return syncHook()
	}
	t.Input.OnDelete = func(path string) error {
		err := t.Input.Fs.RemoveAll(path)
		if err != nil {
			return err
		}

		return syncHook()
	}

	t.Output.OnPushDir = t.Input.OnPushDir
	t.Output.OnPush = t.Input.OnPush
	t.Output.OnRename = t.Input.OnRename
	t.Output.OnDelete = t.Input.OnDelete
}

func (ts *TestFs) SetOutput(outputFs *VFS[string]) error {
	ts.Output = outputFs
	return ts.syncSuite()
}

func (t *TestFs) UnmarshalYAML(node *yaml.Node) error {
	type alias TestFs
	tfs, err := NewTestFs()
	if err != nil {
		return err
	}
	tmp := alias(*tfs)

	err = node.Decode(&tmp)
	if err != nil {
		return err
	}

	*t = TestFs(tmp)

	return t.Input.InitMemMapFs()
}

func (t *TestFs) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Input  map[string]JsTreeNode `json:"input"`
		Output map[string]JsTreeNode `json:"output"`
	}{
		Input:  t.Input.ToJsTreeMap(),
		Output: t.Output.ToJsTreeMap(),
	})
}

type TestSuiteFs struct {
	*VFS[*TestFs]
	root string
}

func NewTestSuiteFs(dir string) (*TestSuiteFs, error) {
	tfs := NewVFS[*TestFs]()

	osPath := func(p string) string {
		return filepath.Join(dir, p)
	}
	tfs.OnPushDir = func(path string) error {
		return os.MkdirAll(osPath(path), 0755)
	}

	tfs.OnPush = func(p *orderedmap.Pair[string, FD[*TestFs]]) error {
		data, err := yaml.Marshal(p.Value.content)
		if err != nil {
			return err
		}

		return os.WriteFile(osPath(p.Key), data, 0755)
	}

	tfs.OnRename = func(oldPath, newPath string) error {
		return os.Rename(osPath(oldPath), osPath(newPath))
	}

	tfs.OnDelete = func(path string) error {
		return os.RemoveAll(osPath(path))
	}

	tfs.OnChild = func(key string) error {
		return tfs.OnPush(tfs.GetPair(key))
	}

	tfs.DecorateFileNode = func(node *JsTreeNode) {
		node.Text = strings.TrimSuffix(node.Text, ".yml")
	}

	err := tfs.UnmarshalDir(dir, len(strings.Split(dir, "/")))
	if err != nil {
		return nil, err
	}

	return &TestSuiteFs{
		tfs,
		dir,
	}, nil
}

func (ts *TestSuiteFs) DecorateFileNode(node *JsTreeNode) {
	node.Text = strings.TrimSuffix(node.Text, ".yml")
}
