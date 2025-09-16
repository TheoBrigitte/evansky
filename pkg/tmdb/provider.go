package tmdb

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

// Name of the provider
const name = "tmdb"

// Flag variables
var (
	apiKey   string
	cacheTTL time.Duration

	apiKeyFlag = "tmdb-api-key"
)

// Provider returns the tmdb provider with its flags
func Provider() provider.Provider {
	flags := pflag.NewFlagSet(name, pflag.ExitOnError)
	flags.StringVar(&apiKey, apiKeyFlag, "", "tmdb api key")
	flags.DurationVar(&cacheTTL, "tmdb-client-cache-ttl", 60*time.Second, "tmdb http client cache ttl, 0 to disable")

	return provider.Provider{
		Name:  name,
		New:   New,
		Flags: flags,
	}
}
