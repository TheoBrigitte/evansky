package common

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// LogLevel set the level of the logger.
func LogLevel(cmd *cobra.Command, args []string) error {
	level, err := log.ParseLevel(cmd.Flag("log-level").Value.String())
	if err != nil {
		return err
	}
	log.SetLevel(level)

	return nil
}

func MultiRuns(cmds ...func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, c := range cmds {
			if err := c(cmd, args); err != nil {
				return err
			}
		}

		return nil
	}
}
