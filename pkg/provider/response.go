package provider

import "time"

// Response represents a search response from a provider.
type Response interface {
	GetID() int
	GetName() string
	GetDate() time.Time
	GetPopularity() int
	GetMediaType() MediaType
}
