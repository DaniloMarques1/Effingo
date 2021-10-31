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
)

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

	filesHashes := make(map[string][]string)
	computeFileHashes(basePath, entries, filesHashes)

	if flag == "-r" {
		removeDuplicates(filesHashes)
	} else {
		printDuplicates(filesHashes)
	}
}

func usage() {
	fmt.Println("To use effingo you need to provide the path that will be analysed:")
	fmt.Println("\teffingo ./path/to/dir")
	fmt.Println()
	fmt.Println("If no flags were given, effingo will search and print the duplicate files.")
	fmt.Println("If you want to remove the duplicate files, you need to provide a -r flag:")
	fmt.Println("\teffingo ./path/to/dir -r")
	fmt.Println()
}

// will return the first argument given to the program
func getArgs() (string, string, error) {
	if len(os.Args) <= 1 {
		return "", "", errors.New(NoFileSystemGiven)
	}

	var flag string
	if len(os.Args) > 2 {
		flag = os.Args[2]
		if flag != "-r" {
			return "", "", errors.New(InvalidFlag)
		}
	}

	return os.Args[1], flag, nil
}

func computeFileHashes(basePath string, entries []os.DirEntry, filesHashes map[string][]string) {
	for _, entry := range entries {
		fullPath := fmt.Sprintf("%s/%s", basePath, entry.Name())
		if !entry.IsDir() {
			bytes, err := getBytesFromFile(fullPath)
			if err != nil {
				log.Fatal(err) // TODO
			}

			hash := computeHash(bytes)
			locations, ok := filesHashes[hash]
			if !ok {
				filesHashes[hash] = []string{fullPath}
			} else {
				locations = append(locations, fullPath)
				filesHashes[hash] = locations
			}
		} else {
			subEntries, err := os.ReadDir(fullPath)
			if err != nil {
				log.Fatal(err) // TODO
			}
			computeFileHashes(fullPath, subEntries, filesHashes)
		}
	}
}

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

func computeHash(bytes []byte) string {
	hash := sha256.Sum256(bytes)
	hex_hash := fmt.Sprintf("%x", hash)
	return hex_hash
}

func printDuplicates(filesHashes map[string][]string) {
	for _, locations := range filesHashes {
		if len(locations) > 1 {
			fmt.Println("Those files are duplicated")
			for _, fileName := range locations {
				fmt.Printf("\t%v\n", fileName)
			}
		}
	}
}

func removeDuplicates(filesHashse map[string][]string) {
	// TODO remove duplicates
	fmt.Println("About to remove duplicates...")
}
