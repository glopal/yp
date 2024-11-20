package yplib

import (
	"io"
)

type MockReader struct {
	name string
	s    string
	i    int64
}

func NewMockReader(name, value string) *MockReader {
	return &MockReader{name, value, 0}
}

func (nsr *MockReader) Name() string {
	return nsr.name
}

func (r *MockReader) Read(b []byte) (n int, err error) {
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n = copy(b, r.s[r.i:])
	r.i += int64(n)
	return
}
