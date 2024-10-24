package main

import (
	"os"

	_ "github.com/glopal/go-yamlplus/tags"
	"github.com/glopal/go-yamlplus/yamlp"
)

func main() {
	nodes, err := yamlp.LoadDir("./fixtures")
	if err != nil {
		panic(err)
	}

	err = nodes.Resolve()
	if err != nil {
		panic(err)
	}

	nodes.PrettyPrintYaml(os.Stdout)

}
