package tmdb

import (
	"math"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type multiResponse struct {
	ID               int64
	Title            string
	Name             string
	OriginalLanguage string
	OriginalName     string
	OriginalTitle    string
	ReleaseDate      string
	FirstAirDate     string
	VoteCount        int64
	VoteAverage      float32
	Popularity       float32

	MediaType provider.MediaType
}

// TODO: replace this with an implementation to generate a proper path, the return path structure should be made configurable in order to be compatble with various mediaserver (Plex, Jellyfin, etc). This should live in a separate package.
func (r multiResponse) GetPath() string {
	if r.MediaType == provider.MediaTypeMovie && r.Title != "" && r.ReleaseDate != "" {
		year := ""
		if len(r.ReleaseDate) >= 4 {
			year = r.ReleaseDate[:4]
		}
		return r.Title + " (" + year + ")"
	}
	if r.MediaType == provider.MediaTypeTV && r.Name != "" && r.FirstAirDate != "" {
		year := ""
		if len(r.FirstAirDate) >= 4 {
			year = r.FirstAirDate[:4]
		}
		return r.Name + " (" + year + ")"
	}
	if r.Title != "" {
		return r.Title
	}
	if r.Name != "" {
		return r.Name
	}
	return ""
}

func (r multiResponse) GetPopularity() int {
	return computePopularity(r.Popularity, r.VoteAverage, r.VoteCount)
}

func computePopularity(popularity, voteAverage float32, voteCount int64) int {
	if popularity > 0 {
		return int(popularity)
	}
	if voteCount == 0 {
		return 0
	}
	return int(math.Round(float64(voteAverage) * math.Log(float64(voteCount))))
}

func (r multiResponse) GetSeasonNumber() int {
	return -1
}

func (r multiResponse) GetEpisodeNumber() int {
	return -1
}
