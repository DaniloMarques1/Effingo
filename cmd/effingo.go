package cmd

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	CacheFileName = ".effingo_cache" // needs a better location
)

type DirTraverser struct {
	basePath         string
	ignoreCache      bool
	shouldRemove     bool
	dirsPaths        []string
	duplicatedHashes []string

	wg sync.WaitGroup

	mu        sync.Mutex
	locations map[string][]string // a map of hash and locations
}

func NewDirTraverser(basePath string, ignoreCache, shouldRemove bool) (*DirTraverser, error) {
	d := &DirTraverser{}

	isDir, err := d.isDir(basePath)
	if err != nil {
		return nil, err
	}

	if !isDir {
		return nil, errors.New("You should provide a valid directory")
	}

	d.basePath = basePath
	d.locations = make(map[string][]string)
	d.ignoreCache = ignoreCache
	d.shouldRemove = shouldRemove
	d.dirsPaths = make([]string, 0)
	d.duplicatedHashes = make([]string, 0)

	return d, nil
}

func (d *DirTraverser) Run() error {
	d.traverse(d.basePath)
	d.wg.Wait()
	log.Printf("Locations = %#v\n", d.locations)
	log.Printf("Locations = %v\n", len(d.locations))
	d.saveCache() // TODO use the cache file

	if d.shouldRemove {
		d.removeDuplicates()
	}

	return nil
}

func (d *DirTraverser) traverse(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("Read dir error = %v\n", err)
		return
	}

	for {
		d.computeEntries(entries, path)
		if len(d.dirsPaths) == 0 {
			log.Printf("No inner dirs")
			break
		}

		path = d.popEntry()
		entries, err = os.ReadDir(path)
		if err != nil {
			log.Printf("Read inner dir error = %v\n", err)
			return
		}
	}
}

func (d *DirTraverser) popEntry() string {
	path := d.dirsPaths[0]
	if len(d.dirsPaths) == 1 {
		d.dirsPaths = []string{}
		return path
	}

	d.dirsPaths = d.dirsPaths[1:]
	return path
}

func (d *DirTraverser) computeEntries(entries []os.DirEntry, path string) {
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			d.dirsPaths = append(d.dirsPaths, entryPath)
		} else {
			d.wg.Add(1)
			go d.computeHash(entryPath)
		}
	}
}

// TODO: need to stop using recursion
//func (d *DirTraverser) traverse2(path string) {
//	entries, err := os.ReadDir(path)
//	if err != nil {
//		log.Printf("Read dir error = %v\n", err)
//		return
//	}
//
//	for _, entry := range entries {
//		if entry.IsDir() {
//			d.traverse(filepath.Join(path, entry.Name()))
//		} else {
//			d.wg.Add(1)
//			go d.computeHash(filepath.Join(path, entry.Name()))
//		}
//	}
//}

// TODO: ignoring the errors for now. maybe keep track of them?
func (d *DirTraverser) computeHash(fileName string) {
	defer d.wg.Done()

	b, err := os.ReadFile(fileName)
	if err != nil {
		log.Printf("Read file error %v\n", err)
		return
	}

	hash := sha256.Sum256(b)
	hexHash := fmt.Sprintf("%x", hash)

	d.mu.Lock()
	defer d.mu.Unlock()
	v, ok := d.locations[hexHash]

	if !ok {
		v = make([]string, 1)
		v[0] = fileName
	} else {
		if len(v) == 1 {
			d.duplicatedHashes = append(d.duplicatedHashes, hexHash)
		}
		v = append(v, fileName)
	}

	d.locations[hexHash] = v
}

func (d *DirTraverser) removeDuplicates() {
	for _, hash := range d.duplicatedHashes {
		location := d.locations[hash]
		for len(location) > 1 {
			var path string
			location, path = d.popLocation(location)
			if err := os.Remove(path); err != nil {
				log.Printf("Error removing duplicated %v\n", err)
			}
		}
	}
}

// pops the first index of the location array and return its path to be removed
func (d *DirTraverser) popLocation(location []string) ([]string, string) {
	if len(location) <= 1 {
		return location, ""
	}

	lastIdx := len(location) - 1
	return location[0:lastIdx], location[lastIdx]
}

func (d *DirTraverser) saveCache() {
	b, err := json.Marshal(d.locations)
	if err != nil {
		log.Printf("Error marshal locations %v\n", err)
	}
	if err := os.WriteFile(CacheFileName, b, os.ModePerm); err != nil {
		log.Printf("Error saving locations %v\n", err)
	}
}

func (d *DirTraverser) isDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Printf("Error stat %v\n", err)
		return false, err
	}
	return fileInfo.IsDir(), nil
}
