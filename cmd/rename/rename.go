package rename

import (
	"fmt"
	"log/slog"

	"github.com/TheoBrigitte/evansky/pkg/provider/register"
	"github.com/TheoBrigitte/evansky/pkg/renamer"

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

	providers, err := register.GetProviders()
	if err != nil {
		return err
	}

	r, err := renamer.New(args, providers)
	if err != nil {
		return err
	}

	slog.Info("start", "sources", len(args), "provider", len(providers))

	return r.Run(dryRun, force)
}
