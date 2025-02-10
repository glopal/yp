package test

import (
	"testing"

	"github.com/glopal/yp/vfs"
)

func Test_Acceptance(t *testing.T) {
	ts, err := vfs.NewTestSuiteFs("../testdata")
	if err != nil {
		t.Fatalf("failed to load testdata: \n%s", err.Error())
	}

	for path, tst := range ts.Tests() {
		t.Run(path, func(t *testing.T) {
			output, err := tst.Run()
			if err != nil {
				t.Errorf("failed to run '%s': \n%s", path, err.Error())
			}

			if !IsEqual(tst, output) {
				t.Errorf("not equal '%s'", path)
			}
		})

	}
}

func IsEqual(expected *vfs.TestFs, actual vfs.YpOutput) bool {
	if actual.Output.Len() != expected.Output.Len() {
		return false
	}
	if actual.Stdout != expected.Stdout {
		return false
	}
	if actual.Err != expected.Err {
		return false
	}

	for key, fd := range actual.Output.FromOldest() {
		efd, exists := expected.Output.OrderedMap.Get(key)
		if !exists {
			return false
		}

		if fd.IsDir() != efd.IsDir() {
			return false
		}

		if fd.Content() != efd.Content() {
			return false
		}
	}

	return true
}
