package test

import (
	"bytes"
	"os"
	"testing"
)

func TestEnvTagSimple(t *testing.T) {
	input := `
ENV: !env ENV
NAME: !env NAME
MISSING: !env MISSING
`
	expected := `
---
ENV: tst
NAME: yamlp
MISSING: ""	
`

	os.Setenv("ENV", "tst")
	os.Setenv("NAME", "yamlp")

	nodes, err := Load([]MockFile{
		{"", input},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = nodes.Resolve()
	if err != nil {
		t.Fatal(err)
	}

	Assert(nodes, expected, t)
}

func TestEnvMap(t *testing.T) {
	input := `
vars: !env/map
	- ENV
	- NAME
	- MISSING
`
	expected := `
---
vars:
	ENV: tst
	NAME: yamlp
	MISSING: ""
`

	os.Setenv("ENV", "tst")
	os.Setenv("NAME", "yamlp")

	nodes, err := Load([]MockFile{
		{"", input},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = nodes.Resolve()
	if err != nil {
		t.Fatal(err)
	}

	Assert(nodes, expected, t)
}

func TestEnvMapAsKey(t *testing.T) {
	input := `
!env/map _:
	- ENV
	- NAME
	- REPLACE
	- MISSING
STATIC: foo
MISSING: default
REPLACE: default
`
	expected := `
---
ENV: tst
NAME: yamlp
STATIC: foo
MISSING: default
REPLACE: new
`

	os.Setenv("ENV", "tst")
	os.Setenv("NAME", "yamlp")
	os.Setenv("REPLACE", "new")

	nodes, err := Load([]MockFile{
		{"", input},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = nodes.Resolve()
	if err != nil {
		t.Fatal(err)
	}

	Assert(nodes, expected, t)
}

func TestEnvMapAsKeyValueMustBeSequence(t *testing.T) {
	input := `
!env/map _: INVALID_KIND
`
	nodes, err := Load([]MockFile{
		{"", input},
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
func TestEnvMapAsKeySequenceValuesMustBeScalar(t *testing.T) {
	input := `
!env/map _:
	- ENV
	- a: 1
`
	nodes, err := Load([]MockFile{
		{"", input},
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
