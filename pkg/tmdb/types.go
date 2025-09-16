package tmdb

import (
	"math"

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
	VoteCount        int64
	VoteAverage      float32
	Popularity       float32
}

func (r response) GetPopularity() int {
	if r.Popularity > 0 {
		return int(r.Popularity)
	}
	if r.VoteCount == 0 {
		return 0
	}
	return int(math.Round(float64(r.VoteAverage) * math.Log(float64(r.VoteCount))))
}
