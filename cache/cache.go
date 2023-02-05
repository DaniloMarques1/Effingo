package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/danilomarques1/effingo/folder"
)

type Cache struct {
	DuplicatedHashes []string            `json:"duplicated_hashes"`
	Locations        map[string][]string `json:"locations"`
	RootPath         string              `json:"root_path"`
}

type CacheWriter interface {
	Save(*Cache) error
	Read() (*Cache, error)
	Evict() error
}

type FileCacheWriter struct {
	cacheFileName string
}

func NewCacheWriter() (CacheWriter, error) {
	c := &FileCacheWriter{}
	dirName, err := folder.GetEffingoFolderPath()
	if err != nil {
		return nil, err
	}

	c.cacheFileName = filepath.Join(dirName, ".effingo_cache")
	return c, nil
}

// Save the cache to cacheFileName as a json
func (c FileCacheWriter) Save(cache *Cache) error {
	b, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	if err := os.WriteFile(c.cacheFileName, b, os.ModePerm); err != nil {
		return err
	}

	return nil
}

// Read the cacheFileName and parses inside a Cache struct.
func (c FileCacheWriter) Read() (*Cache, error) {
	file, err := os.Open(c.cacheFileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	cacheModificationTime := fileInfo.ModTime().Unix()
	if c.hasCacheExpired(cacheModificationTime) {
		return nil, errors.New("Cache expired")
	}
	cache := &Cache{}
	if err := json.NewDecoder(file).Decode(cache); err != nil {
		return nil, err
	}

	return cache, nil
}

// Evict remove the cached file
func (c FileCacheWriter) Evict() error {
	return os.Remove(c.cacheFileName)
}

// hasCacheExpired returns true if the cache has
// been expired (it's at least two minutes old)
func (c FileCacheWriter) hasCacheExpired(cacheTime int64) bool {
	cacheLimitTime := time.Now().Unix() - 120
	return cacheTime < cacheLimitTime
}
