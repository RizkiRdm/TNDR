package config

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// SetupWizard runs the interactive setup process for TENDR.
type SetupWizard struct {
	Stdin     io.Reader
	Stdout    io.Writer
	EnvGet    func(string) string
	Validate  func(ctx context.Context, provider, key string) error
	WriteFile func(filename string, data []byte, perm os.FileMode) error
}

// Run executes the setup wizard workflow.
func (sw *SetupWizard) Run(ctx context.Context, defaultPath string) error {
	scanner := bufio.NewScanner(sw.Stdin)
	fmt.Fprintln(sw.Stdout, "=== TENDR Interactive Setup Wizard ===")

	// 1. Port
	fmt.Fprint(sw.Stdout, "Server Port [4821]: ")
	var port int
	if scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			port = 4821
		} else {
			p, err := strconv.Atoi(text)
			if err != nil || p <= 0 || p > 65535 {
				fmt.Fprintln(sw.Stdout, "Invalid port. Using default 4821.")
				port = 4821
			} else {
				port = p
			}
		}
	}

	// 2. Log Level
	fmt.Fprint(sw.Stdout, "Log Level (debug, info, warn, error) [info]: ")
	logLevel := "info"
	if scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			textLower := strings.ToLower(text)
			if textLower == "debug" || textLower == "info" || textLower == "warn" || textLower == "error" {
				logLevel = textLower
			} else {
				fmt.Fprintln(sw.Stdout, "Invalid log level. Using default 'info'.")
			}
		}
	}

	// Helper for prompting keys
	promptKey := func(providerName, envVar string) string {
		envVal := ""
		if sw.EnvGet != nil {
			envVal = sw.EnvGet(envVar)
		}

		var promptStr string
		if envVal != "" {
			masked := envVal
			if len(envVal) > 8 {
				masked = envVal[:4] + "••••" + envVal[len(envVal)-4:]
			} else {
				masked = "••••••••"
			}
			promptStr = fmt.Sprintf("%s API Key (detected: %s, press Enter to keep): ", strings.Title(providerName), masked)
		} else {
			promptStr = fmt.Sprintf("%s API Key [optional]: ", strings.Title(providerName))
		}

		for {
			fmt.Fprint(sw.Stdout, promptStr)
			if !scanner.Scan() {
				return ""
			}
			input := strings.TrimSpace(scanner.Text())
			key := input
			if input == "" && envVal != "" {
				key = envVal
			}

			if key == "" {
				return ""
			}

			// Validate
			if sw.Validate != nil {
				fmt.Fprintf(sw.Stdout, "Validating %s API Key...\n", providerName)
				err := sw.Validate(ctx, providerName, key)
				if err != nil {
					fmt.Fprintf(sw.Stdout, "⚠️  Validation failed for %s: %v\n", providerName, err)
					fmt.Fprint(sw.Stdout, "Do you want to use this key anyway? (y/n) [n]: ")
					if scanner.Scan() {
						ans := strings.ToLower(strings.TrimSpace(scanner.Text()))
						if ans == "y" || ans == "yes" {
							return key
						}
					}
					// Retry prompt
					continue
				} else {
					fmt.Fprintf(sw.Stdout, "✅  %s API Key is valid!\n", providerName)
				}
			}
			return key
		}
	}

	openaiKey := promptKey("openai", "OPENAI_API_KEY")
	anthropicKey := promptKey("anthropic", "ANTHROPIC_API_KEY")
	geminiKey := promptKey("gemini", "GEMINI_API_KEY")
	groqKey := promptKey("groq", "GROQ_API_KEY")

	// Construct YAML config content
	yamlContent := fmt.Sprintf(`server:
  port: %d
  log_level: %s

providers:
  openai:
    api_key: "%s"
  anthropic:
    api_key: "%s"
  gemini:
    api_key: "%s"
  groq:
    api_key: "%s"

models:
  - alias: coding
    fallback_mode: smart
    providers:
      - openai
      - anthropic
`, port, logLevel, openaiKey, anthropicKey, geminiKey, groqKey)

	// Write config file with 0600 permissions
	err := sw.WriteFile(defaultPath, []byte(yamlContent), 0600)
	if err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	fmt.Fprintf(sw.Stdout, "\nConfiguration successfully written to %s (permissions: 0600)\n", defaultPath)
	return nil
}
