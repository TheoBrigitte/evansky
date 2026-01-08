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
	Scan(string, []provider.Interface, Options) []Node
}

type NodeType int

const (
	NodeTypeUnknown NodeType = iota
	NodeTypeMedia
	NodeTypeSubtitle
)

// Node represents a single file or directory rename operation.
// It contains the original path, the file info and
// metadata retrieved from providers.
type Node struct {
	// Entry is the file system entry information.
	Entry fs.DirEntry
	// Error indicates if there was an error processing this node.
	Error error
	// Info contains the parsed information about the file.
	Info parser.Info
	// Type indicates the type of node (media, subtitle, etc.).
	Type NodeType
	// Path is the original file or directory path.
	Path string
	// Responses holds metadata responses from provider.
	Response provider.Response
}

// Options configures the behavior of source scanning operations.
type Options struct {
	ExcludeGlob  string // A glob pattern to exclude files or directories
	ExcludeRegex string // A regex pattern to exclude files or directories
	IncludeRegex string // A regex pattern to include files or directories
	MediaExts    []string
	SubtitleExts []string
	// TODO: add setting to prefer file name preference over parent directories when finding a match
	Recursive     bool   // Whether to scan directories recursively
	Query         string // Query override for metadata retrieval
	QueryLanguage string // Language code for metadata retrieval
	Language      string // Language code for destination names
	MinDepth      int    // Minimum directory depth to process
	MaxDepth      int    // Maximum directory depth to process
	// TODO: might be an options just for renaming and not sourcing
	SkipDirectories bool // Whether to skip processing directories themselves
	StripComponents int  // Number of leading path components to strip from source paths
}

// Scan is a convenience function that creates a generic source scanner and
// performs a scan operation with the given parameters.
// It returns a list of nodes representing potential rename operations.
func Scan(path string, providers []provider.Interface, o Options) []Node {
	s := &generic{
		path:      path,
		providers: providers,
		options:   o,
	}

	nodes, err := s.scan()
	if err != nil {
		n := Node{
			Path:  path,
			Error: err,
		}
		return []Node{n}
	}

	return nodes
}
