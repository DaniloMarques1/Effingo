package writer

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"
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
	dirName, err := getEffingoFolderPath()
	if err != nil {
		return nil, err
	}

	c.cacheFileName = filepath.Join(dirName, ".effingo_cache")
	return c, nil
}

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
		log.Printf("Has expired\n")
		return nil, errors.New("Cache expired")
	}
	cache := &Cache{}
	if err := json.NewDecoder(file).Decode(cache); err != nil {
		log.Printf("Error decoding cache file %v\n", err)
		return nil, err
	}

	return cache, nil
}

// remove the cached file
func (c FileCacheWriter) Evict() error {
	return os.Remove(c.cacheFileName)
}

func (c FileCacheWriter) hasCacheExpired(cacheTime int64) bool {
	cacheLimitTime := time.Now().Unix() - 120
	return cacheTime < cacheLimitTime
}
