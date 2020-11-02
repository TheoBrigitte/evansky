package tmdb

import (
	gotmdb "github.com/cyruzin/golang-tmdb"
)

// Client to communicate with tmdb api.
type Client struct {
	client *gotmdb.Client
}

// Config contains configuration for tmdb Client.
type Config struct {
	APIKey string
}
