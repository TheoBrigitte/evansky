package list

import (
	"fmt"

	"github.com/TheoBrigitte/evansky/pkg/cache"

	"github.com/spf13/cobra"
)

// Cmd represents the list command
var Cmd = &cobra.Command{
	Use:   "list",
	Short: "list directory contents",
	Long:  `List content of given directory, return entries sorted by filename`,
	RunE:  runner,
}

func runner(cmd *cobra.Command, args []string) error {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}

	caches, err := cache.NewMultiple()
	if err != nil {
		return err
	}

	if len(caches) > 0 {
		for _, c := range caches {
			list, err := c.GetList()
			if err != nil {
				return err
			}

			scan, err := c.GetScan()
			if err != nil {
				return err
			}

			fmt.Printf("%s results=%d/%d", list.Path, scan.Found, scan.Total)
			if verbose {
				fmt.Printf(" filesChecksum=%s pathChecksum=%s", list.FilesChecksum, list.PathChecksum)
			}

			fmt.Println("")
		}
	} else {
		fmt.Println("no results")
	}

	return nil
}
