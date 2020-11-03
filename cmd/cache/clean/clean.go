package clean

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/evansky/pkg/cache"
	"github.com/TheoBrigitte/evansky/pkg/input"
)

var (
	Cmd = &cobra.Command{
		Use:   "clean",
		Short: "clean cache",
		RunE:  runner,
	}

	force bool
)

func init() {
	Cmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "skip confirmation")
}

func runner(cmd *cobra.Command, args []string) error {
	caches, err := cache.NewMultiple()
	if err != nil {
		return err
	}

	fmt.Printf("%d cache entries found\n", len(caches))

	if len(caches) > 0 {
		if !force {
			fmt.Printf("> clean ? [y/N] ")
			ok, err := input.IsResponseYes(input.CurrentMode)
			if err != nil {
				return err
			}
			if !ok {
				os.Exit(0)
			}
		}

		for _, c := range caches {
			err := c.Clean()
			if err != nil {
				return err
			}

			fmt.Printf("%s removed\n", c.Dir())
		}
	}

	return nil
}
