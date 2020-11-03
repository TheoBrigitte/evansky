package status

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/evansky/pkg/cache"
	"github.com/TheoBrigitte/evansky/pkg/list"
)

var (
	Cmd = &cobra.Command{
		Use:   "status",
		Short: "show cache status",
		RunE:  runner,
	}
)

func runner(cmd *cobra.Command, args []string) error {
	caches, err := cache.NewMultiple()
	if err != nil {
		return err
	}

	fmt.Printf("%d cache entries found\n", len(caches))
	for _, c := range caches {
		f, err := c.GetList()
		if err != nil {
			return err
		}

		lister, err := list.New(f.Path)
		if err != nil {
			return err
		}

		result, err := lister.List()
		if err != nil {
			return err
		}

		s, err := c.Status(result.FilesChecksum)
		if err != nil {
			return err
		}
		fmt.Printf("%s status=%s\n", c.Dir(), s)
	}

	return nil
}
