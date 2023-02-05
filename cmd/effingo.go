package cmd

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/danilomarques1/effingo/writer"
)

type DirTraverser struct {
	rootPath         string
	ignoreCache      bool
	shouldRemove     bool
	subDirsPaths     []string // keep track of all subdirs inside the rootPath
	duplicatedHashes []string

	cacheWriter writer.CacheWriter
	logWriter   writer.LogWriter

	wg sync.WaitGroup

	mu     sync.Mutex
	hashes map[string][]string // a map of hash and locations
}

func NewDirTraverser(rootPath string, ignoreCache, shouldRemove bool) (*DirTraverser, error) {
	d := &DirTraverser{}

	isDir, err := d.isDir(rootPath)
	if err != nil || !isDir {
		return nil, errors.New("You should provide a valid directory")
	}

	cacheWriter, err := writer.NewCacheWriter()
	if err != nil {
		return nil, err
	}
	logWriter, err := writer.NewLogWriter()
	if err != nil {
		return nil, err
	}

	d.cacheWriter = cacheWriter
	d.logWriter = logWriter

	d.rootPath = rootPath
	d.ignoreCache = ignoreCache
	d.shouldRemove = shouldRemove

	d.hashes = make(map[string][]string)
	d.subDirsPaths = make([]string, 0)
	d.duplicatedHashes = make([]string, 0)

	return d, nil
}

func (d *DirTraverser) Run() error {
	defer d.logWriter.Flush()

	cache, ok := d.readCacheFile()
	if ok {
		d.hashes = cache.Locations
		d.duplicatedHashes = cache.DuplicatedHashes
	} else {
		// clean the cache file
		d.cacheWriter.Evict()
		d.traverse(d.rootPath)
		d.wg.Wait()
		d.saveCache()
	}

	if d.shouldRemove {
		d.cacheWriter.Evict()
		d.removeDuplicates()
	} else {
		if len(d.duplicatedHashes) == 0 {
			fmt.Printf("You have no duplicated files\n")
		} else {
			d.printDuplicates()
		}
	}

	return nil
}

// receives the last time the cache file was modified (in seconds)
// and returns false if this time is bigger than now - 2 minutes
// meaning the cache did not expired
func (d *DirTraverser) hasCacheExpired(cacheTime int64) bool {
	cacheLimitTime := time.Now().Unix() - 120
	return cacheTime < cacheLimitTime
}

func (d *DirTraverser) traverse(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		d.logWriter.Err(fmt.Sprintf("Read dir error = %v\n", err))
		return
	}

	for {
		d.computeEntries(entries, path)
		if len(d.subDirsPaths) == 0 {
			break
		}

		path = d.popEntry()
		entries, err = os.ReadDir(path)
		if err != nil {
			d.logWriter.Err(fmt.Sprintf("Read inner dir error = %v\n", err))
			return
		}
	}
}

// remove and returns the last path of subDirsPaths
func (d *DirTraverser) popEntry() string {
	lastIdx := len(d.subDirsPaths) - 1
	path := d.subDirsPaths[lastIdx]
	if len(d.subDirsPaths) == 1 {
		d.subDirsPaths = []string{}
		return path
	}

	d.subDirsPaths = d.subDirsPaths[0:lastIdx]
	return path
}

// receives a directories entries and will loop over
func (d *DirTraverser) computeEntries(entries []os.DirEntry, path string) {
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			d.subDirsPaths = append(d.subDirsPaths, entryPath)
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
		d.logWriter.Err(fmt.Sprintf("Read file %v error %v\n", fileName, err))
		return
	}

	hash := sha256.Sum256(b)
	hexHash := fmt.Sprintf("%x", hash)

	d.mu.Lock()
	defer d.mu.Unlock()
	v, ok := d.hashes[hexHash]

	if !ok {
		v = make([]string, 1)
		v[0] = fileName
	} else {
		if len(v) == 1 {
			d.duplicatedHashes = append(d.duplicatedHashes, hexHash)
		}
		v = append(v, fileName)
	}

	d.hashes[hexHash] = v
}

// we remove the duplicate files that we found
func (d *DirTraverser) removeDuplicates() {
	for _, hash := range d.duplicatedHashes {
		location := d.hashes[hash]
		for len(location) > 1 {
			var path string
			location, path = d.popLocation(location)
			fmt.Printf("Removing the file %v\n", path)
			if err := os.Remove(path); err != nil {
				d.logWriter.Err(fmt.Sprintf("Error removing duplicated %v\n", err))
			}
		}
	}
}

func (d *DirTraverser) printDuplicates() {
	for _, hash := range d.duplicatedHashes {
		location := d.hashes[hash]
		for _, path := range location {
			fmt.Printf("- %v\n", path)
		}
		fmt.Println()
	}
}

// pops the last index of the location array and return its path to be removed
func (d *DirTraverser) popLocation(location []string) ([]string, string) {
	if len(location) <= 1 {
		return location, ""
	}

	lastIdx := len(location) - 1
	return location[0:lastIdx], location[lastIdx]
}

// we create a cache entry that can be used if we run the effingo again
func (d *DirTraverser) saveCache() {
	c := &writer.Cache{
		Locations:        d.hashes,
		DuplicatedHashes: d.duplicatedHashes,
		RootPath:         d.rootPath,
	}

	if err := d.cacheWriter.Save(c); err != nil {
		d.logWriter.Err(fmt.Sprintf("Error saving cache %v\n", err))
	}
}

// will return the cached locations if there is one
// if there is no cache it will nil and false
func (d *DirTraverser) readCacheFile() (*writer.Cache, bool) {
	if d.ignoreCache {
		return nil, false
	}

	cache, err := d.cacheWriter.Read()
	if err != nil {
		return nil, false
	}

	// if cached base path is different than the current base path we should not use the cache
	if cache.RootPath != d.rootPath {
		d.logWriter.Err(fmt.Sprintf("Different base paths %v\n", err))
		return nil, false
	}

	return cache, true
}

// returns true if the path is a directory
func (d *DirTraverser) isDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		d.logWriter.Err(fmt.Sprintf("Error stat %v\n", err))
		return false, err
	}
	return fileInfo.IsDir(), nil
}
