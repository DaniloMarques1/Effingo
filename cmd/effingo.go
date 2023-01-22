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
	"time"
)

const (
	CacheFileName = ".effingo_cache" // TODO: needs a better location
)

type DirTraverser struct {
	basePath         string
	ignoreCache      bool
	shouldRemove     bool
	dirsPaths        []string // keep track of all subdirs inside basePath
	duplicatedHashes []string

	wg sync.WaitGroup

	mu        sync.Mutex
	locations map[string][]string // a map of hash and locations
}

type Cache struct {
	DuplicatedHashes []string            `json:"duplicated_hashes"`
	Locations        map[string][]string `json:"locations"`
	BasePath         string              `json:"base_path"`
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
	useCache, cache := d.shouldUseCache()
	if useCache {
		d.locations = cache.Locations
		d.duplicatedHashes = cache.DuplicatedHashes
	} else {
		log.Printf("Got here\n")
		// clean the cache file
		d.wg.Add(1)
		go d.removeCacheFile()
		d.traverse(d.basePath)
		d.wg.Wait()
		d.saveCache()
	}

	//log.Printf("Locations = %#v\n", d.locations)
	log.Printf("Locations = %v\n", len(d.locations))

	if d.shouldRemove {
		d.removeDuplicates()
	} else {
		// TODO: print the location of duplicated files
		d.printDuplicates()
	}

	return nil
}

func (d *DirTraverser) shouldUseCache() (bool, *Cache) {
	if d.ignoreCache {
		return false, nil
	}

	file, err := os.Open(CacheFileName)
	if err != nil {
		log.Printf("Error reading cache file %v\n", err)
		return false, nil
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Error stating cache file %v\n", err)
		return false, nil
	}

	cacheModificationTime := fileInfo.ModTime().Unix()
	if d.hasCacheExpired(cacheModificationTime) {
		return false, nil
	}

	cache := &Cache{}
	if err := json.NewDecoder(file).Decode(cache); err != nil {
		log.Printf("Error decoding cache file %v\n", err)
		return false, nil
	}

	// if cached base path is different than the current base path we should not use the cache
	if cache.BasePath != d.basePath {
		log.Printf("Different base paths %v\n", err)
		return false, nil
	}

	return true, cache
}

func (d *DirTraverser) hasCacheExpired(cacheTime int64) bool {
	cacheLimitTime := time.Now().Unix() - 120
	return cacheTime > cacheLimitTime
}

func (d *DirTraverser) removeCacheFile() error {
	defer d.wg.Done()
	return os.Remove(CacheFileName)
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

// receives a directories entries and will loop over
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

// TODO: ignoring the errors for now. maybe keep track of them?
// compute the file hash. this method will be executed
// on a different goroutine, thus it needs to lock the
// locations before accessing
func (d *DirTraverser) computeHash(fileName string) {
	defer d.wg.Done()

	b, err := os.ReadFile(fileName)
	if err != nil {
		log.Printf("Read file %v error %v\n", fileName, err)
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

// we remove the duplicate files that we found
func (d *DirTraverser) removeDuplicates() {
	for _, hash := range d.duplicatedHashes {
		location := d.locations[hash]
		for len(location) > 1 {
			var path string
			location, path = d.popLocation(location)
			log.Printf("Removing the file %v\n", path)
			if err := os.Remove(path); err != nil {
				log.Printf("Error removing duplicated %v\n", err)
			}
		}
	}
}

func (d *DirTraverser) printDuplicates() {
	for _, hash := range d.duplicatedHashes {
		location := d.locations[hash]
		for _, path := range location {
			fmt.Printf("- %v\n", path)
		}
		fmt.Println()
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

// we create a cache entry that can be used if we run the effingo again
func (d *DirTraverser) saveCache() {
	cache := &Cache{
		Locations:        d.locations,
		DuplicatedHashes: d.duplicatedHashes,
		BasePath:         d.basePath,
	}
	b, err := json.Marshal(cache)
	if err != nil {
		log.Printf("Error marshal locations %v\n", err)
		return
	}

	if err := os.WriteFile(CacheFileName, b, os.ModePerm); err != nil {
		log.Printf("Error saving locations %v\n", err)
	}
}

// returns true if the path is a directory
func (d *DirTraverser) isDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Printf("Error stat %v\n", err)
		return false, err
	}
	return fileInfo.IsDir(), nil
}
