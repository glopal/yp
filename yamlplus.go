package main

import (
	"fmt"

	"github.com/glopal/yamlplus/yamlp"
)

func main() {
	// nodes, err := yamlp.LoadDir("fixtures/map", yamlp.OmitLeadingUnderscore())
	// if err != nil {
	// 	panic(err)
	// }

	// err = nodes.Resolve()
	// if err != nil {
	// 	panic(err)
	// }

	// nodes.PrettyPrintYaml(os.Stdout)

	// fsNode, err := yamlp.UnmarshalDir("fixtures/map")
	// if err != nil {
	// 	panic(err)
	// }

	// y, err := fsNode.ToYaml()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(y))

	// fsys, err := fsNode.ToMemMapFs()
	// if err != nil {
	// 	panic(err)
	// }

	vfs := yamlp.NewVFS()

	// err := vfs.UnmarshalDir("fixtures/map", 1)
	// if err != nil {
	// 	panic(err)
	// }
	err := vfs.UnmarshalYaml("tests/tag_map/basic-usage.yml")
	if err != nil {
		panic(err)
	}

	data, err := vfs.ToYaml()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	// fsys, err := vfs.ToMemMapFs()
	// if err != nil {
	// 	panic(err)
	// }

	// nodes, err := yamlp.LoadDirFS(fsys, ".")
	// if err != nil {
	// 	panic(err)
	// }

	// err = nodes.Resolve()
	// if err != nil {
	// 	panic(err)
	// }

	// nodes.PrettyPrintYaml(os.Stdout)

	// tree, err := vfs.ToJtree()
	// if err != nil {
	// 	panic(err)
	// }
	// treeJson, err := tree.Json()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(string(treeJson))

}
