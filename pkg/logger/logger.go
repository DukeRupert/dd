package logger

import (
	"flag"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

func InitLogger() *zerolog.Logger {
	// Define the log level flag
	logLevelFlag := flag.String("log-level", "info", "sets log level: panic, fatal, error, warn, info, debug, trace")
	flag.Parse()

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set log level based on parameter
	switch strings.ToLower(*logLevelFlag) {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		// Default to info level if invalid level provided
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Create the main logger
	var logger zerolog.Logger

	// Check if debug level or lower is set for pretty console logging
	if zerolog.GlobalLevel() <= zerolog.DebugLevel {
		// Pretty console logging in debug mode
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
		logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	} else {
		// JSON logging for production
		logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	}

	return &logger
}
