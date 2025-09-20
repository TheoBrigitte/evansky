// Package source provides interfaces and types for media file scanning and processing.
// It handles the discovery and analysis of media files (movies and TV shows) within
// directory structures, preparing them for renaming operations.
package source

import (
	"io/fs"

	"github.com/TheoBrigitte/evansky/pkg/parser"
	"github.com/TheoBrigitte/evansky/pkg/provider"
)

// Source defines the interface for media source scanners.
// Implementations of this interface can scan directory structures to identify
// media files and generate rename operations based on metadata providers.
type Source interface {
	// Scan processes the given path using the provided metadata providers and options,
	// returning a list of nodes that represent potential rename operations.
	Scan(string, []provider.Interface, Options) ([]Node, error)
}

// Node represents a single file or directory rename operation.
// It contains the original path, the file info and
// metadata retrieved from providers.
type Node struct {
	Entry    fs.DirEntry
	Info     parser.Info
	Path     string
	Response provider.Response
}

// Options configures the behavior of source scanning operations.
type Options struct {
	// TODO: add setting to prefer file name preference over parent directories when finding a match
	Recursive bool // Whether to scan directories recursively
	MinDepth  int  // Minimum directory depth to process
	MaxDepth  int  // Maximum directory depth to process
	// TODO: might be an options just for renaming and not sourcing
	SkipDirectories bool // Whether to skip processing directories themselves
}

// Scan is a convenience function that creates a generic source scanner and
// performs a scan operation with the given parameters.
// It returns a list of nodes representing potential rename operations.
func Scan(path string, providers []provider.Interface, o Options) ([]Node, error) {
	s := &generic{
		path:      path,
		providers: providers,
		options:   o,
	}

	return s.scan()
}
