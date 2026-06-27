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
func Init(level string, logDir string, maxSizeMB int, maxBackups int, maxAgeDays int) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var writers []io.Writer
	// Always log to stdout
	writers = append(writers, os.Stdout)

	if logDir != "" {
		// Ensure log directory exists
		if err := os.MkdirAll(logDir, 0755); err == nil {
			logFile := filepath.Join(logDir, "tendr.log")
			if maxSizeMB <= 0 {
				maxSizeMB = 50
			}
			if maxBackups <= 0 {
				maxBackups = 5
			}
			if maxAgeDays <= 0 {
				maxAgeDays = 28
			}
			writers = append(writers, &lumberjack.Logger{
				Filename:   logFile,
				MaxSize:    maxSizeMB,
				MaxBackups: maxBackups,
				MaxAge:     maxAgeDays,
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
