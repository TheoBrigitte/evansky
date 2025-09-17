package renamer

import (
	"fmt"

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
		s, err := source.NewSource(path, r.providers, o)
		if err != nil {
			return fmt.Errorf("failed to initialize path %s: %w", path, err)
		}

		err = s.Process()
		if err != nil {
			return fmt.Errorf("failed to get items for path %s: %w", path, err)
		}
	}

	// TODO: handle non destructive renaming, keeping other files (subtitles, etc)

	return nil
}
