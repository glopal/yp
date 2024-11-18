package main

import (
	"errors"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

var (
	ErrFileClosed = errors.New("File is closed")
	// ErrOutOfRange        = errors.New("out of range")
	// ErrTooLarge          = errors.New("too large")
	ErrFileNotFound = os.ErrNotExist
	// ErrFileExists        = os.ErrExist
	// ErrDestinationExists = os.ErrExist
)

type File struct {
	// atomic requires 64-bit alignment for struct field access
	at           int64
	readDirCount int64
	closed       bool
	fileData     *FileData
}

type FileInfo struct {
	*FileData
}

type FileData struct {
	sync.Mutex
	name string
	data []byte
	dir  bool
}

func (f *File) Open() error {
	atomic.StoreInt64(&f.at, 0)
	atomic.StoreInt64(&f.readDirCount, 0)
	f.fileData.Lock()
	f.closed = false
	f.fileData.Unlock()
	return nil
}

func (f *File) Close() error {
	f.fileData.Lock()
	f.closed = true
	f.fileData.Unlock()
	return nil
}

func (f *File) Read(b []byte) (n int, err error) {
	f.fileData.Lock()
	defer f.fileData.Unlock()
	if f.closed {
		return 0, ErrFileClosed
	}
	if len(b) > 0 && int(f.at) == len(f.fileData.data) {
		return 0, io.EOF
	}
	if int(f.at) > len(f.fileData.data) {
		return 0, io.ErrUnexpectedEOF
	}
	if len(f.fileData.data)-int(f.at) >= len(b) {
		n = len(b)
	} else {
		n = len(f.fileData.data) - int(f.at)
	}
	copy(b, f.fileData.data[f.at:f.at+int64(n)])
	atomic.AddInt64(&f.at, int64(n))
	return
}
