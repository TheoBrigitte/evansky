package source

import (
	"os"

	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type Source interface {
	Process() error
}

type Options struct {
	MediaType provider.MediaType
	Recursive bool
	MinDepth  int
	MaxDepth  int
	// TODO: might be an options just for renaming and not sourcing
	SkipDirectories bool
}

type source struct {
	info os.FileInfo

	items []Source
}

// NewSource creates a new source from the given path.
func NewSource(path string, providers []provider.Interface, o Options) (Source, error) {
	return newGeneric(path, providers)
}

type Node struct {
	PathOld string
	PathNew string
}
