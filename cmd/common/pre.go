package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	logLevel      string
	logWithSource bool
)

func Register(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", fmt.Sprintf("set log level (%s)", allLogLevels()))
	cmd.PersistentFlags().BoolVar(&logWithSource, "log-with-source", false, "include source file and line number in log messages")
	cmd.PersistentPreRunE = LogLevel
}

// LogLevel set the level of the logger.
func LogLevel(cmd *cobra.Command, args []string) error {
	level, err := parseLogLevel(logLevel)
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(level)

	// Set log to be printed on stderr with human-friendly format
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	return nil
}

var logLevels = map[string]zerolog.Level{
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"warn":  zerolog.WarnLevel,
	"error": zerolog.ErrorLevel,
}

func allLogLevels() string {
	levels := make([]string, 0, len(logLevels))
	for level := range logLevels {
		levels = append(levels, level)
	}
	return strings.Join(levels, ", ")
}

func parseLogLevel(logLevelStr string) (zerolog.Level, error) {
	level, ok := logLevels[logLevelStr]
	if !ok {
		return -1, fmt.Errorf("invalid log level: %s", logLevelStr)
	}
	return level, nil
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
