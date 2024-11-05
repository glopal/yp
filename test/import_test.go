package test

import (
	"bytes"
	"testing"
)

func TestImportNotFound(t *testing.T) {
	nodes, err := Load([]MockFile{
		{"", "!import foo"},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = nodes.Resolve()
	if err == nil {
		b := bytes.NewBuffer([]byte{})
		nodes.PrettyPrintYaml(b)
		t.Fatalf("expected error, got:\n%s", b.String())
	}
}
