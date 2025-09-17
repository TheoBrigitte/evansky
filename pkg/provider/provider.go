package provider

import (
	"time"

	"github.com/spf13/pflag"
)

// Interface is the interface that all providers must implement.
type Interface interface {
	Name() string
	Search(Request) ([]Response, error)
}

// NewFunc is a function that creates a new provider instance.
type NewFunc func(*pflag.FlagSet) (Interface, error)

// ProviderFunc is a function that returns a Provider.
type ProviderFunc func() Provider

// Provider represents a data provider with its name, constructor function, and flags.
type Provider struct {
	Name  string
	New   NewFunc
	Flags *pflag.FlagSet
}

// Response represents a search response from a provider.
type Response interface {
	GetID() int
	GetName() string
	GetDate() time.Time
	GetPopularity() int
	GetMediaType() MediaType
}
