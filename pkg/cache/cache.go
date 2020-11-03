package cache

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/TheoBrigitte/evansky/pkg/list"
	"github.com/TheoBrigitte/evansky/pkg/scan"
)

const (
	ProjectCacheDir = "evansky"

	fileschecksumFilename = "fileschecksum.json"
	scanFilename          = "scan.json"
)

// NewMultiple return a slice of all available cache entries found.
func NewMultiple() ([]Cache, error) {
	userCache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	projectCache := path.Join(userCache, ProjectCacheDir)

	files, err := ioutil.ReadDir(projectCache)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var caches []Cache
	for _, file := range files {
		if file.IsDir() {
			d := path.Join(projectCache, file.Name())
			c := Cache{
				dir:           d,
				fileschecksum: path.Join(d, fileschecksumFilename),
				scan:          path.Join(d, scanFilename),
			}

			log.Debugf("cache: %s\n", c.dir)
			caches = append(caches, c)
		} else {
			log.Warnf("invalid file %s in cache directory (%s)\n", file.Name(), projectCache)
		}
	}

	return caches, nil
}

// New return cache entry for given dir directory.
func New(dir string) (*Cache, error) {
	userCache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	projectCache := path.Join(userCache, ProjectCacheDir)

	d := path.Join(projectCache, dir)
	c := &Cache{
		dir:           d,
		fileschecksum: path.Join(d, fileschecksumFilename),
		scan:          path.Join(d, scanFilename),
	}

	log.Debugf("cache: %s\n", c.dir)

	return c, nil
}

// Status check cache status.
// e.g. filename=1234/5678
//      DoNotExist: /cache/path/1234/fileschecksum missing
//      Changed:    /cache/path/1234/fileschecksum content != checksum
//      UpToDate:   /cache/path/1234/fileschecksum content == checksum
func (c *Cache) Status(checksum string) (status Status, err error) {
	defer func() {
		log.Debugf("cache status=%s\n", status)
	}()

	f, err := os.Open(c.fileschecksum)
	if os.IsNotExist(err) {
		status = DoNotExists
		return status, nil
	}
	if err != nil {
		return "", nil
	}
	defer f.Close()

	var result list.Result
	err = json.NewDecoder(f).Decode(&result)
	if err != nil {
		return "", nil
	}

	if result.FilesChecksum != checksum {
		status = Changed
		return status, nil
	}

	status = UpToDate
	return status, nil
}

// StoreList store result under filechecksum file.
func (c *Cache) StoreList(result *list.Result) error {
	err := os.MkdirAll(path.Dir(c.fileschecksum), 0755)
	if err != nil {
		return err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.fileschecksum, data, 0755)
	if err != nil {
		return err
	}

	log.Debugf("stored list\n")

	return nil
}

// StoreScan store results under scan file.
func (c *Cache) StoreScan(results *scan.Results) error {
	err := os.MkdirAll(path.Dir(c.scan), 0755)
	if err != nil {
		return err
	}

	data, err := json.Marshal(results)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(c.scan, data, 0755)
	if err != nil {
		return err
	}

	log.Debugf("stored list\n")

	return nil
}

// GetScan return cached scan results.
func (c *Cache) GetScan() (*scan.Results, error) {
	var results scan.Results

	f, err := os.Open(c.scan)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&results)
	if err != nil {
		return nil, err
	}

	log.Debugf("scan contains %d file(s), %d result(s), %s complete\n", results.Total, results.Found, results.CompletePercentage())

	return &results, nil
}

// GetList return cached list results.
func (c *Cache) GetList() (*list.Result, error) {
	var result list.Result

	f, err := os.Open(c.fileschecksum)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&result)
	if err != nil {
		return nil, err
	}

	log.Debugf("list contains %d file(s)\n", result.Files)

	return &result, nil
}

// Size return cache size in bytes.
func (c *Cache) Size() (int64, error) {
	l, err := os.Stat(c.fileschecksum)
	if err != nil {
		return -1, err
	}

	s, err := os.Stat(c.scan)
	if err != nil {
		return -1, err
	}

	return l.Size() + s.Size(), nil
}

// Clean remove cached files.
func (c *Cache) Clean() error {
	return os.RemoveAll(c.dir)
}

// Dir return the cache directory.
func (c *Cache) Dir() string {
	return c.dir
}
