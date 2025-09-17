package tmdb

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"time"

	gotmdb "github.com/cyruzin/golang-tmdb"
	"github.com/gohugoio/hugo/cache/filecache"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/tmdb/httpcache"
)

// New return a new tmdb client.
func New(flags *pflag.FlagSet) (provider.Interface, error) {
	// Validate api key early to catch error before Init.
	if apiKey == "" {
		return nil, fmt.Errorf("--%s is required", apiKeyFlag)
	}

	tmdbClient, err := gotmdb.Init(apiKey)
	if err != nil {
		return nil, err
	}
	tmdbClient.SetClientConfig(newClient(&clientOptions{
		ttl: cacheTTL,
	}))

	c := &Client{
		client: tmdbClient,
	}

	return c, nil
}

func (c *Client) Name() string {
	return name
}

type clientOptions struct {
	ttl time.Duration
}

// newClient returns a new http.Client with caching capabilities if ttl > 0.
func newClient(o *clientOptions) http.Client {
	var transport http.RoundTripper

	if o != nil && o.ttl > 0 {
		// use osFs to store the base files on disk
		osFs := afero.NewOsFs()

		// use baseFs to restrict access to a specific directory
		baseFs := afero.NewBasePathFs(osFs, "/home/theo/projects/evansky/cache")

		// use cacheFs to read the cached files from memory
		cacheFs := afero.NewMemMapFs()

		// Create the caching layer using disk as write and memory as read
		// use 0, since cache ttl is handled by httpcache
		cachedFs := afero.NewCacheOnReadFs(baseFs, cacheFs, 0)

		// create http cache
		c := filecache.NewCache(cachedFs, o.ttl, "")

		transport = &httpcache.Transport{
			Cache:    c.AsHTTPCache(),
			CacheKey: cacheKey,
		}
	} else {
		transport = http.DefaultTransport
	}

	return http.Client{
		Transport: transport,
	}
}

// cacheKey generates a cache key for the given request
// using a SHA1 hash of the method and URL.
func cacheKey(req *http.Request) string {
	key := fmt.Sprintf("%s%s", req.Method, req.URL.String())
	h := sha1.New()
	h.Write([]byte(key))
	key = fmt.Sprintf("%x", h.Sum(nil))

	return key
}
