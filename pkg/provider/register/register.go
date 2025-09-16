package register

import (
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
	// get sorted list of registered provider names.
	names := slices.Collect(maps.Keys(registeredProviders))
	slices.Sort(names)

	// add --provider flag
	cmd.PersistentFlags().StringSliceVar(&chosenProviders, "provider", defaultProviders, "list of commad separated data providers, available: "+strings.Join(names, ","))

	// add providers flags
	cmd.PersistentFlags().AddFlagSet(flags)
}

// mustRegister register a new provider into the registeredProviders map.
// It panics if a provider with the same name is already registered.
func mustRegister(f provider.ProviderFunc) {
	p := f()

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

	for _, name := range chosenProviders {
		newFunc, exists := registeredProviders[name]
		if !exists {
			return nil, fmt.Errorf("provider not registered: %s", name)
		}

		provider, err := newFunc(flags)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize %s provider: %w", name, err)
		}

		providers = append(providers, provider)
	}

	return providers, nil
}
