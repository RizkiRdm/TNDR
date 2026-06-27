package config

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestSetupWizard_Run(t *testing.T) {
	// Simulated inputs:
	// Port: 8080
	// Log level: debug
	// OpenAI key: "sk-openai" (valid)
	// Anthropic key: "sk-anthropic" (first fails validation, then retries with invalid key and says yes to override)
	// Gemini key: (empty, skips)
	// Groq key: (empty, skips)
	inputData := "8080\ndebug\nsk-openai\nsk-bad-anthropic\ny\n\n\n"
	stdin := bytes.NewBufferString(inputData)
	var stdout bytes.Buffer

	envMap := map[string]string{
		"OPENAI_API_KEY": "sk-openai",
	}

	validateCalls := make(map[string][]string)

	sw := &SetupWizard{
		Stdin:  stdin,
		Stdout: &stdout,
		EnvGet: func(key string) string {
			return envMap[key]
		},
		Validate: func(ctx context.Context, provider, key string) error {
			validateCalls[provider] = append(validateCalls[provider], key)
			if provider == "anthropic" && key == "sk-bad-anthropic" {
				return errors.New("invalid key")
			}
			return nil
		},
		WriteFile: func(filename string, data []byte, perm os.FileMode) error {
			if filename != "test-config.yaml" {
				t.Errorf("expected filename 'test-config.yaml', got %s", filename)
			}
			if perm != 0600 {
				t.Errorf("expected permissions 0600, got %o", perm)
			}
			content := string(data)
			if !strings.Contains(content, `port: 8080`) {
				t.Errorf("config missing port: %s", content)
			}
			if !strings.Contains(content, `log_level: debug`) {
				t.Errorf("config missing log_level: %s", content)
			}
			if !strings.Contains(content, `api_key: "sk-openai"`) {
				t.Errorf("config missing openai key: %s", content)
			}
			if !strings.Contains(content, `api_key: "sk-bad-anthropic"`) {
				t.Errorf("config missing anthropic key: %s", content)
			}
			return nil
		},
	}

	err := sw.Run(context.Background(), "test-config.yaml")
	if err != nil {
		t.Fatalf("wizard run failed: %v", err)
	}

	// Verify validation calls
	if len(validateCalls["openai"]) != 1 || validateCalls["openai"][0] != "sk-openai" {
		t.Errorf("unexpected openai validation calls: %v", validateCalls["openai"])
	}
	if len(validateCalls["anthropic"]) != 1 || validateCalls["anthropic"][0] != "sk-bad-anthropic" {
		t.Errorf("unexpected anthropic validation calls: %v", validateCalls["anthropic"])
	}
}
