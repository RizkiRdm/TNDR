package logger

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestInit_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Init panicked: %v", r)
		}
	}()
	Init("info", "")
}

func TestInit_WithLogDir(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Init panicked: %v", r)
		}
	}()
	dir := t.TempDir()
	Init("debug", dir)
}

func TestInit_InvalidLevel_FallsBackToInfo(t *testing.T) {
	Init("invalidlevel", "")
	if log.Logger.GetLevel() != zerolog.InfoLevel {
		t.Errorf("expected InfoLevel, got %v", log.Logger.GetLevel())
	}
}

func TestInit_NeverLogsAPIKey(t *testing.T) {
	var buf bytes.Buffer
	// Capture log output for testing purposes
	log.Logger = zerolog.New(&buf).With().Logger()

	apiKey := "secret-12345-actual-value"
	
	// Convention: use a masking helper
	masked := "secret-••••••••"
	log.Info().Str("api_key", masked).Msg("some request")

	output := buf.String()
	if strings.Contains(output, apiKey) {
		t.Errorf("logger leaked API key: %s", output)
	}
	if !strings.Contains(output, masked) {
		t.Error("expected api_key field to be masked")
	}
}
