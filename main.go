package main

import (
	"log"

	"github.com/TrianaLab/remake/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
