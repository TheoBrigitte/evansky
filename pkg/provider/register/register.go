package register

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/TheoBrigitte/evansky/pkg/provider"
	"github.com/TheoBrigitte/evansky/pkg/tmdb"
)

var (
	defaultProviders = []string{"tmdb"}
	chosenProviders  []string

	// global FlagSet containing all flags from all registeredProviders.
	flags = pflag.NewFlagSet("provider", pflag.ExitOnError)
	// registered providers.
	registeredProviders = map[string]provider.NewFunc{}
)

func init() {
	mustRegister(tmdb.Provider)
}

func Initialize(cmd *cobra.Command) {
	names := slices.Collect(maps.Keys(registeredProviders))
	cmd.PersistentFlags().StringSliceVar(&chosenProviders, "provider", defaultProviders, "list of commad separated data providers, available: "+strings.Join(names, ","))
	cmd.PersistentFlags().AddFlagSet(flags)
}

// mustRegister register a new provider into the registeredProviders map.
// It panics if a provider with the same name is already registered.
func mustRegister(p provider.Provider) {
	if _, exists := registeredProviders[p.Name]; exists {
		panic("provider already registered: " + p.Name)
	}

	if p.Flags != nil {
		// register all flags from newSet into the global flags FlagSet.
		// panics if a flag is already registered.
		p.Flags.VisitAll(func(flag *pflag.Flag) {
			// enforce namespacing of flags by prefixing them with the provider name.
			if !strings.HasPrefix(flag.Name, p.Name) {
				panic(fmt.Sprintf("flag %s must be prefixed with provider name %s", flag.Name, p.Name))
			}

			if flags.Lookup(flag.Name) != nil {
				panic(fmt.Sprintf("flag %s already registered", flag.Name))
			}

			flags.AddFlag(flag)
		})
	}

	registeredProviders[p.Name] = p.New
}

func GetProviders() ([]provider.Interface, error) {
	var providers []provider.Interface
	var errs []error

	for _, name := range chosenProviders {
		newFunc, exists := registeredProviders[name]
		if !exists {
			errs = append(errs, fmt.Errorf("unknown provider: %s", name))
			continue
		}

		provider, err := newFunc(flags)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to initialize provider %s: %w", name, err))
			continue
		}

		providers = append(providers, provider)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to get providers:\n%v", errors.Join(errs...))
	}

	return providers, nil
}
