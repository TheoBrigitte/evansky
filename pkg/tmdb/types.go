package tmdb

import (
	gotmdb "github.com/cyruzin/golang-tmdb"
)

// Client to communicate with tmdb api.
type Client struct {
	client *gotmdb.Client
}
