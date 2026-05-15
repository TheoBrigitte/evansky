package httpcache

import (
	"crypto/sha1" //nolint:gosec
	"fmt"
	"net/http"
	"time"

	"github.com/gohugoio/hugo/cache/filecache"
	"github.com/spf13/afero"
)

type Options struct {
	CacheDir string
	TTL      time.Duration
}

// New returns a new http.Client with caching capabilities if ttl > 0.
func New(o Options) *http.Client {
	var transport http.RoundTripper

	if o.TTL > 0 {
		// use osFs to store the base files on disk
		osFs := afero.NewOsFs()

		// use baseFs to restrict access to a specific directory
		baseFs := afero.NewBasePathFs(osFs, o.CacheDir)

		// use cacheFs to read the cached files from memory
		cacheFs := afero.NewMemMapFs()

		// Create the caching layer using disk as write and memory as read
		// use 0, since cache ttl is handled by httpcache
		cachedFs := afero.NewCacheOnReadFs(baseFs, cacheFs, 0)

		// create http cache
		// TODO: cleanup cache directory
		c := filecache.NewCache(cachedFs, filecache.FileCacheConfig{
			MaxAge: o.TTL,
			Dir:    o.CacheDir,
		})

		transport = &Transport{
			Cache:    c.AsHTTPCache(),
			CacheKey: cacheKey,
		}
	} else {
		transport = http.DefaultTransport
	}

	return &http.Client{
		Transport: transport,
	}
}

// cacheKey generates a cache key for the given request
// using a SHA1 hash of the method and URL.
func cacheKey(req *http.Request) string {
	key := fmt.Sprintf("%s%s", req.Method, req.URL.String())
	h := sha1.New() //nolint:gosec
	h.Write([]byte(key))
	key = fmt.Sprintf("%x", h.Sum(nil))

	return key
}
