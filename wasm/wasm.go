//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"syscall/js"

	"github.com/glopal/yp/yplib"
	"github.com/spf13/afero"
)

func main() {

	js.Global().Set("yamlp", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		afs := afero.NewMemMapFs()
		err := afero.WriteFile(afs, "example.yml", []byte(args[0].String()), 0755)
		if err != nil {
			return err
		}

		b := bytes.NewBuffer([]byte{})
		err = yplib.Load("example.yml").Options(yplib.WithFS(afero.NewIOFS(afs)), yplib.WithWriter(b)).Out()
		if err != nil {
			return err
		}

		return b.String()
	}))

	select {}
}
