package rename

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/TheoBrigitte/evansky/pkg/provider/register"
	"github.com/TheoBrigitte/evansky/pkg/renamer"
	"github.com/TheoBrigitte/evansky/pkg/renamer/format"
	"github.com/TheoBrigitte/evansky/pkg/source"

	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:   "rename [flags] <file | directory>...",
		Short: "rename directory content",
		Long:  `Rename and organize directory content`,
		RunE:  runner,
		Args:  cobra.MinimumNArgs(1),
	}

	exclude    string
	force      bool
	language   string
	output     string
	query      string
	renameMode string
	write      bool
)

func init() {
	Cmd.PersistentFlags().StringVar(&exclude, "exclude", "", "exclude files matching the given glob pattern")
	Cmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "skip confirmation")
	Cmd.PersistentFlags().StringVar(&language, "language", "en", "language used for destination names (ISO 639-1 code)")
	Cmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output directory (default: same as source)")
	Cmd.PersistentFlags().StringVar(&query, "query", "", "search query override")
	Cmd.PersistentFlags().StringVar(&renameMode, "mode", "symlink", "rename mode: symlink, hardlink, copy, move")
	Cmd.PersistentFlags().BoolVar(&write, "write", false, "actually perform the rename operation (default: false)")

	register.Initialize(Cmd)
}

func runner(cmd *cobra.Command, args []string) error {
	providers, err := register.GetProviders()
	if err != nil {
		return err
	}

	formatter := format.NewJellyfinFormatter()

	if output != "" {
		info, err := os.Lstat(output)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil && !info.IsDir() {
			return fmt.Errorf("output is not a directory: %s", output)
		}
		output = filepath.Clean(output)
	}

	renameOptions := renamer.Options{
		Formatter:  formatter,
		Output:     output,
		RenameMode: renameMode,
	}
	if write {
		renameOptions.Write = true
	}

	r, err := renamer.New(args, providers, renameOptions)
	if err != nil {
		return err
	}

	sourceOptions := source.Options{
		Query:       query,
		ExcludeGlob: exclude,
	}

	return r.Run(sourceOptions)
}
