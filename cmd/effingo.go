package cmd

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type DirTraverser struct {
	basePath string

	wg sync.WaitGroup

	mu        sync.Mutex
	locations map[string][]string // a map of hash and locations
}

func NewDirTraverser(basePath string) (*DirTraverser, error) {
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
	return d, nil
}

func (d *DirTraverser) Run() error {
	d.traverse(d.basePath)
	d.wg.Wait()
	log.Printf("Locations = %#v\n", d.locations)
	return nil
}

// TODO: need to stop using recursion
func (d *DirTraverser) traverse(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("Read dir error = %v\n", err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			d.traverse(filepath.Join(path, entry.Name()))
		} else {
			d.wg.Add(1)
			go d.hashFile(filepath.Join(path, entry.Name()))
		}
	}
}

// TODO: use goroutines!!!!
// TODO: ignoring the errors for now. maybe keep track of them?
// this functions is being executed on a different goroutine
func (d *DirTraverser) hashFile(fileName string) {
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
		v = append(v, fileName)
	}

	d.locations[hexHash] = v
}

func (d *DirTraverser) isDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Printf("Error stat %v\n", err)
		return false, err
	}
	return fileInfo.IsDir(), nil
}
