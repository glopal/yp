package vfs

import (
	"bytes"
	"encoding/json"
	"iter"
	"os"
	"path/filepath"
	"strings"

	"github.com/glopal/yp/yplib"
	"github.com/spf13/afero"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type TestFs struct {
	Input     *VFS[string] `yaml:"input"`
	Output    *VFS[string] `yaml:"output"`
	Stdout    string       `yaml:"stdout,omitempty"`
	Err       string       `yaml:"err,omitempty"`
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

func (ts *TestFs) SetStdout(val string) error {
	ts.Stdout = val
	return ts.syncSuite()
}
func (ts *TestFs) SetErr(val string) error {
	ts.Err = val
	return ts.syncSuite()
}

func (ts *TestFs) SetOutput(outputFs *VFS[string], stdout, err string) error {
	ts.Output = outputFs
	ts.Stdout = stdout
	ts.Err = err
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

type testFs struct {
	Input  map[string]JsTreeNode `json:"input"`
	Output map[string]JsTreeNode `json:"output"`
	Stdout string                `json:"stdout,omitempty"`
	Err    string                `json:"err,omitempty"`
}

func (t *TestFs) MarshalJSON() ([]byte, error) {
	return json.Marshal(testFs{
		Input:  t.Input.ToJsTreeMap(),
		Output: t.Output.ToJsTreeMap(),
		Stdout: t.Stdout,
		Err:    t.Err,
	})
}

type YpOutput struct {
	Output *VFS[string] `json:"output"`
	Stdout string       `json:"stdout"`
	Err    string       `json:"err"`
}

func (t *TestFs) Run() (YpOutput, error) {
	output := YpOutput{}
	ofs := afero.NewMemMapFs()
	b := bytes.NewBuffer([]byte{})

	err := yplib.WithOptions(yplib.WithFS(afero.NewIOFS(t.Input.Fs)), yplib.WithOutputFS(ofs), yplib.WithWriter(b)).Load(".").Out()
	if err != nil {
		output.Err = err.Error()
	}

	output.Stdout = b.String()

	outputVfs, err := UnmarshalFs(afero.NewIOFS(ofs))
	if err != nil {
		return output, err
	}

	output.Output = outputVfs

	return output, nil
}

// func (t *TestFs) Validate() (YpOutput, error) {

// }

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

func (ts *TestSuiteFs) Tests() iter.Seq2[string, *TestFs] {
	return func(yield func(string, *TestFs) bool) {
		for pair := ts.Oldest(); pair != nil; pair = pair.Next() {
			if pair.Value.IsDir() {
				continue
			}
			if !yield(pair.Key, pair.Value.content) {
				return
			}
		}
	}
}

func (ts *TestSuiteFs) DecorateFileNode(node *JsTreeNode) {
	node.Text = strings.TrimSuffix(node.Text, ".yml")
}
