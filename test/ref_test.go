package test

import (
	"bytes"
	"os"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/glopal/yamlplus/yamlp"
)

func TestMain(m *testing.M) {
	approvals.UseFolder("approvals")

	os.Exit(m.Run())
}
func TestRefsSimple(t *testing.T) {
	refYml := `
--- #ref/one
a: 1
b: 2
c: 3
--- #ref/two
one: !yq $one.a
---
!yq $two
	`

	nodes, err := Load([]MockFile{
		{"ref.yml", refYml},
	})
	if err != nil {
		t.Fatal()
	}

	err = nodes.Resolve()
	if err != nil {
		t.Fatal(err)
	}

	approvals.Verify(t, OutReader(nodes), approvals.Options().WithExtension(".yml"))
}

func TestRefsAreScopedToDocForRefResolution(t *testing.T) {
	nodes, err := yamlp.LoadDir("fixtures/refs/scoped")
	if err != nil {
		t.Fatal()
	}

	err = nodes.Resolve()
	if err == nil {
		b := bytes.NewBuffer([]byte{})
		nodes.PrettyPrintYaml(b)
		t.Fatalf("expected error, got:\n%s", b.String())
	}
}
