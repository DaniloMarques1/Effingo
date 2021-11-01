package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	NoFileSystemGiven = "You need to provide a file system path"
	InvalidFlag       = "The given flag is invalid"
	InvalidUsage      = "Invalid usage"
)

type Flag struct {
	shouldRemove  bool
	includeHidden bool
}

func main() {
	basePath, flag, err := getArgs()
	if err != nil {
		usage()
		log.Fatal(err) // TODO
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		log.Fatal(err) // TODO
	}

	hashes := make(map[string][]string)
	computeFileHashes(basePath, entries, hashes, flag.includeHidden)

	if flag.shouldRemove {
		removeDuplicates(hashes)
	} else {
		printDuplicates(hashes)
	}
}

// help function
func usage() {
	fmt.Println("To use effingo you need to provide the path that will be analysed:")
	fmt.Println("\teffingo ./path/to/dir")
	fmt.Println()
	fmt.Println("If no flags were given, effingo will search and print the duplicate files.")
	fmt.Println("If you want to remove the duplicate files, you need to provide a -r flag:")
	fmt.Println("\teffingo ./path/to/dir -r")
	fmt.Println("If you want to include the hidden files in the seach add the -i flag:")
	fmt.Println("\teffingo ./path/to/dir -i")

	fmt.Println()
}

// will return the first argument given to the program
func getArgs() (string, *Flag, error) {
	if len(os.Args) <= 1 {
		return "", nil, errors.New(NoFileSystemGiven)
	}
	if len(os.Args) > 4 {
		return "", nil, errors.New(InvalidUsage)
	}

	var flag Flag
	for _, f := range os.Args[2:] {
		if f == "-r" {
			flag.shouldRemove = true
		} else if f == "-i" {
			flag.includeHidden = true
		}
	}

	return os.Args[1], &flag, nil
}

// traversy the entries of the given file system and populate the filehashes map
func computeFileHashes(basePath string, entries []os.DirEntry, hashes map[string][]string, includeHidden bool) {
	for _, entry := range entries {
		fullPath := fmt.Sprintf("%s/%s", basePath, entry.Name())
		if entry.IsDir() {
			// TODO finds a way to do it without recursion to prevent call stack problem
			if entry.Name()[0] == '.' && !includeHidden {
				continue
			}

			subEntries, err := os.ReadDir(fullPath)
			if err != nil {
				log.Fatal(err) // TODO
			}
			computeFileHashes(fullPath, subEntries, hashes, includeHidden)

		} else {
			bytes, err := getBytesFromFile(fullPath)
			if err != nil {
				log.Fatal(err) // TODO
			}

			hash := computeHash(bytes)
			locations, ok := hashes[hash]
			if !ok {
				hashes[hash] = []string{fullPath}
			} else {
				locations = append(locations, fullPath)
				hashes[hash] = locations
			}
		}
	}
}

// reads the fileName file and return its bytes
func getBytesFromFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// return the sha256 hash of bytes
func computeHash(bytes []byte) string {
	hash := sha256.Sum256(bytes)
	hex_hash := fmt.Sprintf("%x", hash)
	return hex_hash
}

// iterates the file hashes and print the files names
// that are duplicated
func printDuplicates(hashes map[string][]string) {
	for _, locations := range hashes {
		if len(locations) > 1 {
			fmt.Println("Those files are duplicated")
			for _, fileName := range locations {
				fmt.Printf("\t%v\n", fileName)
			}
		}
	}
}

// iterates the file hashes and remove the files
// that are duplicated
func removeDuplicates(hashes map[string][]string) {
	for _, locations := range hashes {
		if len(locations) > 1 {
			for _, fileName := range locations[1:] {
				fmt.Printf("Removing duplicate file %v\n", fileName)
				err := os.Remove(fileName)
				if err != nil {
					log.Fatal(err)
				}
			}
			fmt.Printf("Remaining %v\n\n", locations[0])
		}
	}
}
