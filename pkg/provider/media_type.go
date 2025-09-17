package provider

// MediaType represents the type of media (movie, tv, person, etc).
type MediaType int

const (
	MediaTypeUnknown MediaType = iota
	MediaTypeMovie
	MediaTypeTV
	MediaTypeTVSeason
	MediaTypeTVEpisode
	MediaTypeCollection
)

func (m MediaType) String() string {
	switch m {
	case MediaTypeMovie:
		return "movie"
	case MediaTypeTV:
		return "tv"
	case MediaTypeTVSeason:
		return "tv_season"
	case MediaTypeTVEpisode:
		return "tv_episode"
	case MediaTypeCollection:
		return "collection"
	default:
		return "unknown"
	}
}
