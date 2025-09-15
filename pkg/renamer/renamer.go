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
	sources   []source.Interface
	providers []provider.Interface
}

func New(sources []source.Interface, providers []provider.Interface) (Renamer, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("at least one source is required")
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("at least one provider is required")
	}

	r := &renamer{
		sources:   sources,
		providers: providers,
	}

	return r, nil
}

func (r *renamer) Run(dryRun, force bool) error {
	fmt.Println("rename is not implemented yet")
	return nil
}
