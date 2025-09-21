package rename

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/TheoBrigitte/evansky/pkg/provider/register"
	"github.com/TheoBrigitte/evansky/pkg/renamer"
	"github.com/TheoBrigitte/evansky/pkg/renamer/format"

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

	dryRun     bool
	force      bool
	language   string
	output     string
	renameMode string
	write      bool
)

func init() {
	Cmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", true, "show what would be done only and do not rename anything")
	Cmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "skip confirmation")
	Cmd.PersistentFlags().StringVar(&language, "language", "en", "language used for destination names (ISO 639-1 code)")
	Cmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output directory (default: same as source)")
	Cmd.PersistentFlags().StringVar(&renameMode, "mode", "symlink", "rename mode: symlink, hardlink, copy, move")
	Cmd.PersistentFlags().BoolVar(&write, "write", false, "actually perform the rename operation (default: false)")

	register.Initialize(Cmd)
}

func runner(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("directory missing")
	}

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
			return fmt.Errorf("output %q is not a directory", output)
		}
		output = filepath.Clean(output)
	}

	renameOptions := renamer.Options{
		Formatter:  formatter,
		RenameMode: renameMode,
		Output:     output,
	}
	if !dryRun || write {
		renameOptions.Write = true
	}
	r, err := renamer.New(args, providers, renameOptions)
	if err != nil {
		return err
	}

	slog.Info("start", "sources", len(args), "provider", len(providers))

	return r.Run()
}
