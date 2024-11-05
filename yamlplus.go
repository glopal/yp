package main

import (
	"os"

	"github.com/glopal/yamlplus/yamlp"
)

func main() {
	nodes, err := yamlp.LoadDir("fixtures/refs", yamlp.OmitLeadingUnderscore())
	if err != nil {
		panic(err)
	}

	err = nodes.Resolve()
	if err != nil {
		panic(err)
	}

	nodes.PrettyPrintYaml(os.Stdout)

}
