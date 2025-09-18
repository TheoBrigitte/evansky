package provider

import "time"

// Response represents a search response from a provider.
type Response interface {
	// Commont methods for all media types.
	GetID() int
	GetName() string
	GetDate() time.Time
	// TODO: improve this by returning a pointer and nil when not applicable, or implement sub interface for movie/tv/season/episode
	GetSeasonNumber() int
	GetEpisodeNumber() int

	GetPopularity() int
	GetMediaType() MediaType
}
