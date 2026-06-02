package logger

import (
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Init initializes the global logger with the specified log level and log directory.
func Init(level string, logDir string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var writers []io.Writer
	// Always log to stdout
	writers = append(writers, os.Stdout)

	if logDir != "" {
		// Ensure log directory exists
		if err := os.MkdirAll(logDir, 0755); err == nil {
			logFile := filepath.Join(logDir, "tendr.log")
			writers = append(writers, &lumberjack.Logger{
				Filename:   logFile,
				MaxSize:    10, // megabytes
				MaxBackups: 3,
				MaxAge:     28, // days
				Compress:   true,
			})
		}
	}

	multi := io.MultiWriter(writers...)
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	log.Logger = zerolog.New(multi).With().Timestamp().Logger().Level(lvl)
}
