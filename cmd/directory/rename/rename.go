package rename

import (
	"fmt"
	"os"
	"path"

	"github.com/TheoBrigitte/evansky/pkg/cache"
	"github.com/TheoBrigitte/evansky/pkg/input"
	"github.com/TheoBrigitte/evansky/pkg/list"

	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:   "rename",
		Short: "rename directory content",
		Long:  `Rename and organize directory content`,
		RunE:  runner,
		Args:  cobra.ExactValidArgs(1),
	}

	dryRun bool
	force  bool
)

func init() {
	Cmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "do not change anything")
	Cmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "skip confirmation")
}

func runner(cmd *cobra.Command, args []string) error {
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return err
	}

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

	if !force {
		fmt.Printf("> is scan complete ? ")
		if !results.IsComplete() {
			fmt.Println("no")
			fmt.Printf("> continue ? [y/N] ")
			ok, err := input.IsResponseYes(input.CurrentMode)
			if err != nil {
				return err
			}

			if !ok {
				os.Exit(0)
			}
		} else {
			fmt.Println("yes")
		}

		fmt.Printf("> review scan ? [y/N] ")
		ok, err := input.IsResponseYes(input.CurrentMode)
		if err != nil {
			return err
		}
		if ok {
			results.Print(os.Stdout, false)
		}

		fmt.Printf("> about to rename %d file(s) in %s. proceed ? [y/N] ", results.Found, args[0])
		ok, err = input.IsResponseYes(input.CurrentMode)
		if err != nil {
			return err
		}

		if !ok {
			fmt.Println("> stoping")
			os.Exit(0)
		}
	}

	fmt.Printf("> renaming")
	if dryRun {
		fmt.Printf(" (dry run)")
	}
	fmt.Println("")

	for name, r := range results.Results {
		source := path.Join(args[0], name)
		destination := path.Join(args[0], r.Path)

		if !r.IsDir {
			if !quiet {
				fmt.Printf("create directory %#q\n", path.Dir(destination))
			}
			if !dryRun {
				err = os.MkdirAll(path.Dir(destination), 0755)
				if err != nil {
					return err
				}
			}
		}

		_, err = os.Stat(destination)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil {
			return fmt.Errorf("path already exists: %#q\n", destination)
		}

		if !quiet {
			fmt.Printf("rename %#q -> %#q\n", source, destination)
		}
		if !dryRun {
			err = os.Rename(source, destination)
			if err != nil {
				return err
			}
		}
	}

	fmt.Printf("> renamed %d file(s)", len(results.Results))
	if dryRun {
		fmt.Printf(" (dry run)")
	}
	fmt.Println("")

	if !dryRun {
		err = c.Clean()
		if err != nil {
			return err
		}
	}
	fmt.Printf("> cleaned cache %s", c.Dir())
	if dryRun {
		fmt.Printf(" (dry run)")
	}
	fmt.Println("")

	return nil
}
