package main

import "fmt"

func main() {
	tfs := NewTFS()
	content := "bla"
	tfs.Set("test", &VFD{Content: &content})

	vfd, ok := tfs.GetByKey("test")
	if !ok {
		panic("bla")
	}

	fmt.Println(*vfd.Content)
}
