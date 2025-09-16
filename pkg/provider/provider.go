package provider

import (
	"github.com/spf13/pflag"
)

type Interface interface {
	Name() string
	Search(Request, MediaType) ([]Response, error)
}

type NewFunc func(*pflag.FlagSet) (Interface, error)

type ProviderFunc func() Provider

type Provider struct {
	Name  string
	New   NewFunc
	Flags *pflag.FlagSet
}

type Request struct {
	Query   string
	Year    int
	Season  int
	Episode int
}

type Response interface {
	GetPopularity() int
}

// MediaType represents the type of media (movie, tv, person, etc).
type MediaType int

const (
	MediaTypeUnknown MediaType = iota
	MediaTypeMovie
	MediaTypeTV
)
