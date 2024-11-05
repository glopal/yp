package yamlp

import "testing"

func TestLoadGlob(t *testing.T) {
	nodes, err := LoadDir("./fixtures")
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	expectedLen := 3
	actualLen := len(nodes.nodes)
	if expectedLen != actualLen {
		t.Fatalf("expected %d nodes, got %d", expectedLen, actualLen)
	}
}

// expect errors
// - refs need to come before exports in file
