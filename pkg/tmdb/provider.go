package tmdb

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

const (
	// Name of the provider
	name            = "tmdb"
	defaultCacheDir = "evansky/tmdb"
)

// Flag variables
var (
	apiKey   string
	cacheTTL time.Duration

	apiKeyFlag = "tmdb-api-key"
	cacheDir   string
)

// Provider returns the tmdb provider with its flags
func Provider() provider.Provider {
	flags := pflag.NewFlagSet(name, pflag.ExitOnError)
	flags.StringVar(&apiKey, apiKeyFlag, "", "tmdb api key")
	flags.StringVar(&cacheDir, "tmdb-cache-dir", "", "cache directory (default: $XDG_CACHE_HOME/evansky/tmdb or $HOME/.cache/evansky/tmdb)")
	flags.DurationVar(&cacheTTL, "tmdb-client-cache-ttl", 60*time.Second, "tmdb http client cache ttl, 0 to disable")

	return provider.Provider{
		Name:  name,
		New:   New,
		Flags: flags,
	}
}
