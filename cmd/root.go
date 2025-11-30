package cmd

import (
	"os"

	"github.com/prometheus/common/version"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/TheoBrigitte/evansky/cmd/common"
	"github.com/TheoBrigitte/evansky/cmd/completion"
	"github.com/TheoBrigitte/evansky/cmd/rename"
)

// rootCmd represents the base command when called without any subcommands
var (
	name = "evansky"

	rootCmd = &cobra.Command{
		Use:           name,
		Short:         "media renamer",
		Long:          `Rename media files in order to be detected by media server like Jellyfin.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.Print(name),
	}
)

func init() {
	rootCmd.SetVersionTemplate(`{{.Version}}`)
	rootCmd.AddCommand(completion.Cmd)
	rootCmd.AddCommand(rename.Cmd)
	common.Register(rootCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Err(err).Send()
		os.Exit(1)
	}
}
