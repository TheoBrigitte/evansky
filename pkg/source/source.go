package source

import (
	"github.com/TheoBrigitte/evansky/pkg/provider"
)

type Source interface {
	Scan(string, []provider.Interface, Options) ([]Node, error)
}

type Node struct {
	PathOld string
	PathNew string
}

type Options struct {
	// TODO: add setting to prefer file name preference over parent directories when finding a match
	Recursive bool
	MinDepth  int
	MaxDepth  int
	// TODO: might be an options just for renaming and not sourcing
	SkipDirectories bool
}

func Scan(path string, providers []provider.Interface, o Options) ([]Node, error) {
	s := &generic{
		path:      path,
		providers: providers,
		options:   o,
	}

	return s.scan()
}
