package main

import (
	"log"

	"github.com/TrianaLab/remake/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	remake := cmd.NewRemakeCommand()
	err := doc.GenMarkdownTree(remake, "./docs")
	if err != nil {
		log.Fatal(err)
	}
}
