package provider

// MediaType represents the type of media (movie, tv, person, etc).
type MediaType int

const (
	MediaTypeUnknown MediaType = iota
	MediaTypeMovie
	MediaTypeTV
	MediaTypeSeason
	MediaTypeEpisode
	MediaTypeCollection
)
