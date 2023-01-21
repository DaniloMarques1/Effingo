package main

import (
	"flag"
	"log"

	"github.com/danilomarques1/effingo/cmd"
)

var (
	dir = flag.String("dir", ".", "The directory to search for duplicate files")
)

func main() {
	flag.Parse()

	traverser, err := cmd.NewDirTraverser(*dir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%v\n", traverser)

	traverser.Run()
}
