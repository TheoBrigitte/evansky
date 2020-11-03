package list

import (
	"fmt"

	"github.com/docker/go-units"
	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/evansky/pkg/cache"
)

var (
	Cmd = &cobra.Command{
		Use:   "list",
		Short: "list cache entries",
		RunE:  runner,
	}
)

func runner(cmd *cobra.Command, args []string) error {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}

	caches, err := cache.NewMultiple()
	if err != nil {
		return err
	}

	var totalSize int64
	fmt.Printf("%d cache entries found\n", len(caches))
	for _, c := range caches {
		size, err := c.Size()
		if err != nil {
			return err
		}
		totalSize += size

		fmt.Printf("%s size=%s", c.Dir(), units.HumanSize(float64(size)))

		if verbose {
			list, err := c.GetList()
			if err != nil {
				return err
			}

			scan, err := c.GetScan()
			if err != nil {
				return err
			}

			fmt.Printf(" results=%d/%d path=%s", scan.Found, scan.Total, list.Path)
		}

		fmt.Println("")
	}

	fmt.Printf("\ntotal size %s\n", units.HumanSize(float64(totalSize)))

	return nil
}
