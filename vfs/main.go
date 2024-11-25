package main

import (
	"fmt"
)

func main() {
	ts, err := NewTestSuiteFs("vfs/testdata")
	if err != nil {
		panic(err)
	}

	err = ts.WriteInput("tags/yq/simple.yml", "root/node.yml", "a: 1\nb: 5")
	if err != nil {
		panic(err)
	}

	// err = ts.Rename("foo/bbb", "foo/aaa")
	// if err != nil {
	// 	panic(err)
	// }

	// err = ts.Delete("test")
	// if err != nil {
	// 	panic(err)
	// }

	// err = ts.CreateDir("test/foo/bar")
	// if err != nil {
	// 	panic(err)
	// }

	// err = ts.CreateDir("test/foo/aaa")
	// if err != nil {
	// 	panic(err)
	// }

	// err = ts.CreateDir("test/foo/aaa/bar")
	// if err != nil {
	// 	panic(err)
	// }

	// err = ts.CreateDir("aaa")
	// if err != nil {
	// 	panic(err)
	// }

	// err = ts.CreateDir("zzz")
	// if err != nil {
	// 	panic(err)
	// }

	// err = ts.CreateDir("test/bbb")
	// if err != nil {
	// 	panic(err)
	// }

	// err = ts.Rename("test/foo", "foo")
	// if err != nil {
	// 	panic(err)
	// }
	// err = ts.Rename("test/bbb", "bbb")
	// if err != nil {
	// 	panic(err)
	// }
	// err = ts.Rename("foo/aaa", "foo/bbb")
	// if err != nil {
	// 	panic(err)
	// }

	data, err := ts.ToYaml()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
}
