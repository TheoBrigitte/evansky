package renamer

import (
	"fmt"
	"log/slog"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/source"
)

type Renamer interface {
	Run(dryRun, force bool) error
}

type renamer struct {
	paths     []string
	providers []provider.Interface
}

func New(paths []string, providers []provider.Interface) (Renamer, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("at least one source is required")
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("at least one provider is required")
	}

	r := &renamer{
		paths:     paths,
		providers: providers,
	}

	return r, nil
}

func (r *renamer) Run(dryRun, force bool) error {
	o := source.Options{}

	for _, path := range r.paths {
		_, err := source.Scan(path, r.providers, o)
		if err != nil {
			slog.Error("scan failed", "path", path, "error", err)
			continue
		}
	}

	// TODO: handle non destructive renaming, keeping other files (subtitles, etc)
	// TODO: include directories in the renaming process

	return nil
}
