package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

const (
	CacheExpired = "Cache file expired"
	CacheFile    = ".cache"
	CacheFileSec = 300 // five minutes in seconds
)

// helper messages
const (
	ShouldRemoveMessage  = "Indicates if should remove the duplicated files"
	IncludeHiddenMessage = "Indicates if should include all files"
	DirMessage           = "The file system to look for duplicate files"
	HelpMessage          = "Show the usage of effingo"
)

var (
	shouldRemove  bool
	includeHidden bool
	dir           string
	help          bool
)

func Run() {
	parseFlags()

	if help {
		usage()
		return
	}

	hashes := make(map[string][]string)
	if cached, err := readCacheFile(); err == nil {
		// if no errors was returned
		hashes = cached
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			log.Fatal(err) // TODO
		}

		computeFileHashes(dir, entries, hashes)
		saveCache(hashes) // ignoring possible errors
	}

	if shouldRemove {
		removeDuplicates(hashes)
	} else {
		printDuplicates(hashes)
	}
}

// read the .cache file and see if it still a valid cache file
func readCacheFile() (map[string][]string, error) {
	fileInfo, err := os.Stat(CacheFile)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	modTime := fileInfo.ModTime()
	fileModTime := modTime.Unix() + CacheFileSec
	if now > fileModTime {
		return nil, errors.New(CacheExpired)
	}

	b, err := os.ReadFile(CacheFile)
	if err != nil {
		return nil, err
	}
	hashes := make(map[string][]string)
	if err := json.Unmarshal(b, &hashes); err != nil {
		return nil, err
	}

	return hashes, nil
}

// parse command line flags
func parseFlags() {
	flag.BoolVar(&shouldRemove, "r", false, ShouldRemoveMessage)
	flag.BoolVar(&shouldRemove, "remove", false, ShouldRemoveMessage)

	flag.BoolVar(&includeHidden, "a", false, IncludeHiddenMessage)
	flag.BoolVar(&includeHidden, "all", false, IncludeHiddenMessage)

	flag.StringVar(&dir, "d", ".", DirMessage)
	flag.StringVar(&dir, "dir", ".", DirMessage)

	flag.BoolVar(&help, "help", false, HelpMessage)

	flag.Parse()
}

// traversy the entries of the given file system and populate the filehashes map
func computeFileHashes(basePath string, entries []os.DirEntry,
	hashes map[string][]string) {
	for _, entry := range entries {
		fullPath := fmt.Sprintf("%s/%s", basePath, entry.Name())
		if entry.IsDir() {
			// TODO finds a way to do it without recursion to prevent call stack problem
			if entry.Name()[0] == '.' && !includeHidden {
				continue
			}

			subEntries, err := os.ReadDir(fullPath)
			if err != nil {
				fmt.Printf("There was a problem when trying to open the directory %v\n", fullPath)
				continue
			}
			computeFileHashes(fullPath, subEntries, hashes)

		} else {
			bytes, err := getBytesFromFile(fullPath)
			if err != nil {
				fmt.Printf("There was a problem when trying to open the file %v\n", fullPath)
				continue
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
			fmt.Println("These files are duplicated")
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

func saveCache(hashes map[string][]string) error {
	b, err := json.Marshal(hashes)
	if err != nil {
		return err
	}
	if err := os.WriteFile(CacheFile, b, 0777); err != nil {
		return err
	}
	return nil
}

// help function
func usage() {
	fmt.Println("To use effingo you need to provide the path that will be analysed:")
	fmt.Println("\teffingo -d ./path/to/dir")
	fmt.Println()
	fmt.Println("If no flags were given, effingo will search and print the duplicate files.")
	fmt.Println("If you want to remove the duplicate files, you need to provide a -r flag:")
	fmt.Println("\teffingo ./path/to/dir -r")
	fmt.Println("If you want to include the hidden files in the seach add the -i flag:")
	fmt.Println("\teffingo ./path/to/dir -i")
}
