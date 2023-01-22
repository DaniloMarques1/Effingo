package main

import (
	"flag"
	"log"

	"github.com/danilomarques1/effingo/cmd"
)

var (
	dir          = flag.String("dir", ".", "The directory to search for duplicate files")
	ignoreCache  = flag.Bool("ignore-cache", false, "Ignore the use of the cached file")
	shouldRemove = flag.Bool("remove", false, "If should remove the files")
)

func main() {
	flag.Parse()

	traverser, err := cmd.NewDirTraverser(*dir, *ignoreCache, *shouldRemove)
	if err != nil {
		log.Fatal(err)
	}

	traverser.Run()
}
