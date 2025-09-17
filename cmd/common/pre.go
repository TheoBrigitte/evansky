package common

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

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

	logWithSource, err := cmd.Flags().GetBool("log-with-source")

	// Set the default logger for the application.
	loggerOptions := &slog.HandlerOptions{
		AddSource: logWithSource,
		Level:     level,
	}
	logger := slog.NewTextHandler(os.Stderr, loggerOptions)
	slog.SetDefault(slog.New(logger))

	return nil
}

var (
	logLevels = map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
)

func allLogLevels() string {
	levels := make([]string, 0, len(logLevels))
	for level := range logLevels {
		levels = append(levels, level)
	}
	return strings.Join(levels, ", ")
}

func parseLogLevel(logLevelStr string) (slog.Level, error) {
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
