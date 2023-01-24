package cache

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"
)

const (
	CacheFileName = ".effingo_cache"
)

type Cache struct {
	DuplicatedHashes []string            `json:"duplicated_hashes"`
	Locations        map[string][]string `json:"locations"`
	RootPath         string              `json:"root_path"`
}

type CacheService interface {
	Save(*Cache) error
	Read() (*Cache, error)
	Evict() error
}

type FileCacheService struct {
	dirName string
}

func NewCacheService() (CacheService, error) {
	c := &FileCacheService{}
	dirName, err := c.createCacheFolder()
	if err != nil {
		return nil, err
	}
	return &FileCacheService{dirName: dirName}, nil
}

func (c *FileCacheService) createCacheFolder() (string, error) {
	curUser, err := user.Current()
	if err != nil {
		return "", err
	}

	system := runtime.GOOS

	var cachePath string
	switch system {
	case "windows":
		cachePath = filepath.Join(curUser.HomeDir, ".effingo")
	default:
		cachePath = filepath.Join(curUser.HomeDir, ".cache/effingo")
	}

	if err := os.Mkdir(cachePath, os.ModePerm); err != nil {
		if errors.Is(err, os.ErrExist) {
			return cachePath, nil
		}
		return "", err
	}

	return cachePath, nil
}

func (c FileCacheService) Save(cache *Cache) error {
	b, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	if err := os.WriteFile(c.fileName(), b, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func (c FileCacheService) fileName() string {
	return filepath.Join(c.dirName, CacheFileName)
}

func (c FileCacheService) Read() (*Cache, error) {
	file, err := os.Open(c.fileName())
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
func (c FileCacheService) Evict() error {
	return os.Remove(c.fileName())
}

func (c FileCacheService) hasCacheExpired(cacheTime int64) bool {
	cacheLimitTime := time.Now().Unix() - 120
	return cacheTime < cacheLimitTime
}
