package vfs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type ParentSyncer interface {
	SetSyncHook(func() error)
}

type FD[T any] struct {
	isFile  bool
	content T
}

func (fd FD[T]) Bytes() ([]byte, error) {
	return []byte(fmt.Sprintf("%v", fd.content)), nil
}

func NewFile[T any](content T) FD[T] {
	return FD[T]{
		isFile:  true,
		content: content,
	}
}

func Decorate[T any](key string, item *FD[T], onChild func(string) error) {
	var i interface{} = item.content
	syncer, ok := i.(ParentSyncer)
	if !ok {
		return
	}

	syncer.SetSyncHook(func() error {
		return onChild(key)
	})

	content, _ := syncer.(T)
	item.content = content
}

func (fd FD[T]) IsDir() bool {
	return !fd.isFile
}

func (fd *FD[T]) UnmarshalYAML(node *yaml.Node) (err error) {
	if node.Kind == yaml.MappingNode && len(node.Content) == 0 {
		return nil
	}

	fd.isFile = true
	content := *new(T)

	err = node.Decode(&content)
	if err != nil {
		return err
	}
	fd.content = content

	return nil

}

func (fd FD[T]) MarshalYAML() (interface{}, error) {
	if fd.IsDir() {
		return yaml.Node{
			Kind: yaml.MappingNode,
		}, nil
	}

	var node yaml.Node
	err := node.Encode(fd.content)

	return node, err
}

type VFS[T any] struct {
	*orderedmap.OrderedMap[string, FD[T]]
	OnPushDir        func(string) error
	OnPush           func(*orderedmap.Pair[string, FD[T]]) error
	OnRename         func(string, string) error
	OnDelete         func(string) error
	OnChild          func(string) error
	DecorateFileNode func(node *JsTreeNode)
	Fs               afero.Fs
}

func NewVFS[T any]() *VFS[T] {
	return &VFS[T]{
		orderedmap.New[string, FD[T]](),
		func(path string) error {
			return nil
		},
		func(*orderedmap.Pair[string, FD[T]]) error {
			return nil
		},
		func(string, string) error {
			return nil
		},
		func(string) error {
			return nil
		},
		nil,
		func(node *JsTreeNode) {},
		nil,
	}
}

func (vfs *VFS[T]) Get(key string) (T, bool) {
	item, exists := vfs.OrderedMap.Get(key)
	return item.content, exists
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

func (vfs *VFS[T]) AtOrOldest(key string) *orderedmap.Pair[string, FD[T]] {
	if key == "" {
		return vfs.Oldest()
	}

	pair := vfs.GetPair(key)
	if pair == nil {
		return vfs.Oldest()
	}

	return pair
}

func (vfs *VFS[T]) PushDir(path string) error {
	if path == "" || vfs.Exists(path) {
		return nil
	}

	paths := []string{}
	parent := path

	for {
		paths = append(paths, parent)
		parent = getParentKey(parent)
		if parent == "" {
			break
		}

		if vfs.Exists(parent) {
			break
		}
	}

	parent = vfs.getAfterKey(paths[len(paths)-1], parent)

	for i := len(paths) - 1; i >= 0; i-- {
		key := paths[i]
		vfs.Set(key, FD[T]{})

		if parent == "" {
			err := vfs.MoveToFront(key)
			if err != nil {
				return err
			}
		} else {
			err := vfs.MoveAfter(key, parent)
			if err != nil {
				return err
			}
		}

		parent = key
	}

	return vfs.OnPushDir(path)
}

func (vfs *VFS[T]) Push(newKey string, val T) error {
	if vfs.Exists(newKey) {
		pair := vfs.GetPair(newKey)
		pair.Value.content = val

		return vfs.OnPush(pair)
	}

	parentKey := getParentKey(newKey)
	vfs.PushDir(parentKey)

	parentKey = vfs.getAfterKey(newKey, parentKey)

	f := NewFile(val)
	Decorate(newKey, &f, vfs.OnChild)
	vfs.Set(newKey, f)

	if parentKey == "" {
		err := vfs.MoveToFront(newKey)
		if err != nil {
			return err
		}
	} else {
		err := vfs.MoveAfter(newKey, parentKey)
		if err != nil {
			return err
		}
	}

	return vfs.OnPush(vfs.GetPair(newKey))
}

// func (vfs *VFS[T]) Replace(key string, val *VFS[T]) error {

// 	return vfs.OnPush(vfs.GetPair(key))
// }

func (vfs *VFS[T]) getAfterKey(newKey string, startKey string) string {
	afterKey := ""
	start := vfs.AtOrOldest(startKey)
	if start == nil {
		return ""
	}
	if start.Next() == nil {
		return start.Key
	}

	for cur := vfs.AtOrOldest(startKey); cur != nil; cur = cur.Next() {
		val := strings.Compare(cur.Key, newKey)

		if val > 0 { // curKey > newKey
			break
		} else if val < 0 { // curKey < newKey
			afterKey = cur.Key
		}
	}

	return afterKey
}

func (vfs *VFS[T]) Delete(key string) error {
	pair := vfs.GetPair(key)
	if pair == nil {
		return nil
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

	return vfs.OnDelete(key)
}

func (vfs *VFS[T]) Rename(orgKey string, newKey string) error {
	if vfs.Exists(newKey) {
		return fmt.Errorf("can't rename '%s' to existing key '%s'", orgKey, newKey)
	}

	pair := vfs.GetPair(orgKey)
	if pair == nil {
		return fmt.Errorf("'%s' does not exists", orgKey)
	}

	parent := ""
	if prev := pair.Prev(); prev != nil {
		parent = prev.Key
	}

	deletes := []string{orgKey}
	newPairs := []*orderedmap.Pair[string, FD[T]]{{
		Key:   newKey,
		Value: pair.Value,
	}}

	if pair.Value.IsDir() {
		for cur := pair.Next(); cur != nil; cur = cur.Next() {
			if strings.HasPrefix(cur.Key, orgKey) {
				deletes = append(deletes, cur.Key)
				newPairs = append(newPairs, &orderedmap.Pair[string, FD[T]]{
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
		key := p.Key
		vfs.Set(key, p.Value)

		if parent == "" {
			err := vfs.MoveToFront(key)
			if err != nil {
				return err
			}
		} else {
			err := vfs.MoveAfter(key, parent)
			if err != nil {
				return err
			}
		}

		parent = key
	}

	return vfs.OnRename(orgKey, newKey)
}

func (vfs *VFS[T]) ToYaml() ([]byte, error) {
	return yaml.Marshal(vfs)
}

func UnmarshalFs(fsys fs.FS) (*VFS[string], error) {
	vfs := NewVFS[string]()
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}

		if !d.IsDir() {
			data, err := fs.ReadFile(fsys, path)
			if err != nil {
				return err
			}

			vfs.Set(path, FD[string]{isFile: true, content: string(data)})
		} else {
			vfs.Set(path, FD[string]{})
		}

		return nil
	})

	return vfs, err
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

			vfs.Set(increment, FD[T]{})
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

			item := NewFile(*new(T))

			err = yaml.Unmarshal(data, &item)
			// item, err = item.Decode(data)
			if err != nil {
				return err
			}

			Decorate(vPath, &item, vfs.OnChild)

			vfs.Set(vPath, item)
		} else {
			vfs.Set(vPath, FD[T]{})
		}

		return nil
	})

	return err
}

func (vfs *VFS[T]) InitMemMapFs() error {
	vfs.Fs = afero.NewMemMapFs()

	for path, vfd := range vfs.FromOldest() {
		if vfd.IsDir() {
			err := vfs.Fs.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		} else {
			data, err := vfd.Bytes()
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			if err := afero.WriteFile(vfs.Fs, path, data, 0755); err != nil {
				return err
			}
		}
	}

	return nil
}
