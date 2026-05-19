// Package tmdb provides a client for The Movie Database (TMDB) API.
package tmdb

import (
	"context" //nolint:gosec
	"fmt"
	"os"
	"path/filepath"

	"github.com/golusoris/goenvoy/metadata"
	"github.com/golusoris/goenvoy/metadata/video/tmdb"
	"github.com/spf13/pflag"

	"github.com/TheoBrigitte/evansky/pkg/httpcache"
	"github.com/TheoBrigitte/evansky/pkg/provider"
)

// New return a new tmdb client.
func New(flags *pflag.FlagSet) (provider.Interface, error) {
	// Validate api key early to catch error before Init.
	if apiKey == "" {
		apiKey = os.Getenv(apiKeyEnvVar)
		if apiKey == "" {
			return nil, fmt.Errorf("TMDB Api Key is required, set it either via --%s flag or %s environment variable", apiKeyFlag, apiKeyEnvVar)
		}
	}

	cd := cacheDir
	if cd == "" {
		dir, err := os.UserCacheDir()
		if err != nil {
			return nil, err
		}
		cd = filepath.Join(dir, defaultCacheDir)
	}

	tmdbClient := tmdb.New(apiKey, metadata.WithHTTPClient(httpcache.New(httpcache.Options{
		CacheDir: cd,
		TTL:      cacheTTL,
	})))

	c := &Client{
		client: tmdbClient,
		ctx:    context.TODO(),
	}

	return c, nil
}

func (c *Client) Name() string {
	return name
}
