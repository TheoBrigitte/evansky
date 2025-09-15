package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/evansky/cmd/common"
	"github.com/TheoBrigitte/evansky/cmd/completion"
	"github.com/TheoBrigitte/evansky/cmd/rename"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:               "evansky",
	Short:             "media renamer",
	Long:              `Rename media files in order to be detected by media server like Jellyfin.`,
	PersistentPreRunE: common.LogLevel,
	SilenceUsage:      true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(completion.Cmd)
	rootCmd.AddCommand(rename.Cmd)
}
