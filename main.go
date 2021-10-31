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
)

func main() {
	basePath, err := getFsFromArgs()
	if err != nil {
		log.Fatal(err) // TODO
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		log.Fatal(err) // TODO
	}

	filesHashes := make(map[string][]string)
	computeFileHashes(basePath, entries, filesHashes)
	fmt.Println(filesHashes)
}

// will return the first argument given to the program
func getFsFromArgs() (string, error) {
	if len(os.Args) == 1 {
		return "", errors.New(NoFileSystemGiven)
	}

	return os.Args[1], nil
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
