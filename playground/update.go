package main

import (
	"fmt"

	"github.com/glopal/yp/vfs"
)

type Op string
type Target string

const (
	PUSH_DIR Op = "PUSH_DIR"
	PUSH     Op = "PUSH"
	RENAME   Op = "RENAME"
	DELETE   Op = "DELETE"
)

const (
	INPUT  Target = "input"
	OUTPUT Target = "output"
	STDOUT Target = "stdout"
	ERR    Target = "err"
)

type UpdateBody struct {
	Op      Op     `json:"op"`
	Id      string `json:"id"`
	OldId   string `json:"oldId"`
	Content any    `json:"content"`
}

type UpdateTestBody struct {
	UpdateBody
	Target   Target `json:"target"`
	ParentId string `json:"parentId"`
}

func (u UpdateBody) Update(ts *vfs.TestSuiteFs) error {
	switch u.Op {
	case PUSH_DIR:
		return ts.PushDir(u.Id)
	case PUSH:
		t, err := vfs.NewTestFs()
		if err != nil {
			return err
		}
		return ts.Push(u.Id, t)
	case RENAME:
		return ts.Rename(u.OldId, u.Id)
	case DELETE:
		return ts.Delete(u.Id)
	}

	return nil
}

func (u UpdateTestBody) Update(ts *vfs.TestSuiteFs) error {
	t, ok := ts.Get(u.ParentId)
	if !ok {
		return fmt.Errorf("failed to get testfs (%s)", u.ParentId)
	}
	var target *vfs.VFS[string]

	switch u.Target {
	case INPUT:
		target = t.Input
	case OUTPUT:
		target = t.Output
	case STDOUT:
		content, _ := u.Content.(string)
		return t.SetStdout(content)
	case ERR:
		content, _ := u.Content.(string)
		return t.SetErr(content)
	}

	switch u.Op {
	case PUSH_DIR:
		return target.PushDir(u.Id)
	case PUSH:
		content, _ := u.Content.(string)
		return target.Push(u.Id, content)
	case RENAME:
		return target.Rename(u.OldId, u.Id)
	case DELETE:
		return target.Delete(u.Id)
	}

	return nil
}
