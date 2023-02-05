package main

import (
	"flag"
	"log"
	"os"

	"github.com/danilomarques1/effingo/cmd"
	"github.com/danilomarques1/effingo/folder"
)

var (
	ignoreCache  = flag.Bool("ignore-cache", false, "Ignore the use of the cached file")
	shouldRemove = flag.Bool("remove", false, "If should remove the files")
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	dir := flag.String("dir", cwd, "The directory to search for duplicate files")
	flag.Parse()

	if err := folder.CreateEffingoFolter(); err != nil {
		log.Fatal(err)
	}

	traverser, err := cmd.NewDirTraverser(*dir, *ignoreCache, *shouldRemove)
	if err != nil {
		log.Fatal(err)
	}

	traverser.Run()
}
