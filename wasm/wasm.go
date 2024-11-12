//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	"syscall/js"

	"github.com/glopal/yamlplus/yamlp"
)

func main() {
	js.Global().Set("yamlp", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		reader := yamlp.NewMockReader("example.yml", args[0].String())
		nodes, err := yamlp.Load(reader)
		if err != nil {
			return err
		}

		err = nodes.Resolve()
		if err != nil {
			return err
		}

		b := bytes.NewBuffer([]byte{})
		nodes.PrettyPrintYaml(b)
		return b.String()
	}))

	select {}
}
