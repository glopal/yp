package main

import (
	"log"

	"github.com/glopal/yp/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
