package tmdb

import (
	gotmdb "github.com/cyruzin/golang-tmdb"
)

// Client to communicate with tmdb api.
type Client struct {
	client *gotmdb.Client
}

type response struct {
	ID               int64
	Title            string
	Name             string
	MediaType        string
	OriginalLanguage string
	OriginalName     string
	OriginalTitle    string
	ReleaseDate      string
	FirstAirDate     string
}
