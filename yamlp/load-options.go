package yamlp

import (
	"path/filepath"
	"strings"
)

type loadOptions struct {
	omitFunc func(string) bool
}

func defaultLoadOptions() *loadOptions {
	return &loadOptions{
		omitFunc: func(s string) bool {
			return false
		},
	}
}

func OmitLeadingUnderscore() func(*loadOptions) {
	return func(lo *loadOptions) {
		lo.omitFunc = func(path string) bool {
			return strings.HasPrefix(filepath.Base(path), "_")
		}
	}
}

func OmitDotFiles() func(*loadOptions) {
	return func(lo *loadOptions) {
		lo.omitFunc = func(path string) bool {
			return strings.HasPrefix(filepath.Base(path), ".")
		}
	}
}
