package test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/glopal/go-yamlplus/yamlp"
)

type MockFile struct {
	Name string
	Body string
}

type mockFileReader struct {
	name string
	s    string
	i    int64
}

func (nsr *mockFileReader) Name() string {
	return nsr.name
}

func (r *mockFileReader) Read(b []byte) (n int, err error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n = copy(b, r.s[r.i:])
	r.i += int64(n)
	return
}

func toMockFileReaders(files []MockFile) []yamlp.NamedReader {
	readers := make([]yamlp.NamedReader, 0, len(files))
	for _, file := range files {
		readers = append(readers, &mockFileReader{file.Name, cleanTabs(file.Body), 0})
	}

	return readers
}

func cleanTabs(str string) string {
	return strings.ReplaceAll(str, "\t", "    ")
}

func Load(files []MockFile) (*yamlp.Nodes, error) {
	return yamlp.Load(toMockFileReaders(files)...)
}

func OutReader(nodes *yamlp.Nodes) io.Reader {
	r, w := io.Pipe()
	go func() {
		nodes.PrettyPrintYaml(w)
		w.Close()
	}()

	return r
}

func Assert(nodes *yamlp.Nodes, expected string, t *testing.T) {
	expected = cleanTabs(strings.TrimSpace(expected))
	b := bytes.NewBuffer([]byte{})
	nodes.PrettyPrintYaml(b)

	actual := strings.TrimSpace(b.String())

	if actual != expected {
		t.Fatalf("\n# expected \n%s\n\n# got\n%s", expected, actual)
	}
}
