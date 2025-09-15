package rename

import (
	"errors"
	"fmt"

	"github.com/TheoBrigitte/evansky/pkg/provider/register"
	"github.com/TheoBrigitte/evansky/pkg/renamer"
	"github.com/TheoBrigitte/evansky/pkg/source"

	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:   "rename",
		Short: "rename directory content",
		Long:  `Rename and organize directory content`,
		RunE:  runner,
		Args:  cobra.MinimumNArgs(1),
	}

	dryRun bool
	force  bool
)

func init() {
	Cmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "do not change anything")
	Cmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "skip confirmation")
	register.Initialize(Cmd)
}

func runner(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("directory missing")
	}

	var sources []source.Interface
	var errs []error
	for _, path := range args {
		source, err := source.New(path)
		if err != nil {
			// collect errors but try all paths
			// in order to give the user a complete feedback
			errs = append(errs, err)
			continue
		}
		sources = append(sources, source)
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to read path(s)\n%v", errors.Join(errs...))
	}

	providers, err := register.GetProviders()
	if err != nil {
		return err
	}

	r, err := renamer.New(sources, providers)
	if err != nil {
		return err
	}

	fmt.Printf("renaming %d source(s)\n", len(sources))
	fmt.Printf("using %d provider(s)\n", len(providers))
	return r.Run(dryRun, force)
}
