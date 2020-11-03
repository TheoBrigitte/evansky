package show

import (
	"fmt"
	"os"

	"github.com/TheoBrigitte/evansky/pkg/cache"
	"github.com/TheoBrigitte/evansky/pkg/list"

	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:     "show",
		Aliases: []string{"view"},
		Short:   "show scan result",
		RunE:    runner,
		Args:    cobra.ExactValidArgs(1),
	}

	verbose bool
)

func init() {
	Cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show more informations")
}

func runner(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("directory missing")
	}

	lister, err := list.New(args[0])
	if err != nil {
		return err
	}

	c, err := cache.New(lister.PathChecksum())
	if err != nil {
		return err
	}

	results, err := c.GetScan()
	if err != nil {
		return err
	}

	results.Print(os.Stdout, verbose)

	return nil
}
