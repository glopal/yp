package main

import orderedmap "github.com/wk8/go-ordered-map/v2"

type TFS[T any] struct {
	*orderedmap.OrderedMap[string, T]
}

type Content struct {
	data *string
}

func (c Content) Decode([]byte) error {
	return nil
}
func (c Content) Encode() ([]byte, error) {
	return nil, nil
}
func (c Content) IsDir() bool {
	return c.data == nil
}

func NewTFS[T any]() *TFS[T] {
	return &TFS[T]{orderedmap.New[string, T]()}
}
func (tfs *TFS[T]) GetByKey(key string) (T, bool) {
	return tfs.Get(key)
}
