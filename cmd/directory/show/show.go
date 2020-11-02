package show

import (
	"fmt"
	"os"
	"text/tabwriter"

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

	if len(results.Results) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 7, ' ', tabwriter.AlignRight)

		fmt.Fprintf(w, "original\tnew\t")
		if verbose {
			fmt.Fprintf(w, "id\tisDir\t")
		}
		fmt.Fprintln(w, "")

		fmt.Fprintf(w, "--------\t---\t")
		if verbose {
			fmt.Fprintf(w, "--\t-----\t")
		}
		fmt.Fprintln(w, "")

		for name, r := range results.Results {
			fmt.Fprintf(w, "%s\t%s\t", name, r.Path)
			if verbose {
				fmt.Fprintf(w, "%d\t%t\t", r.ID, r.IsDir)
			}

			fmt.Fprintln(w, "")
		}
		w.Flush()
		fmt.Println("")
		fmt.Printf("%d/%d result(s)  %s complete\n", results.Found, results.Total, results.CompletePercentage())
	} else {
		fmt.Println("no results")
	}

	return nil
}
