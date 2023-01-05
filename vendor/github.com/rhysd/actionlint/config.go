package actionlint

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config is configuration of actionlint. This struct instance is parsed from "actionlint.yaml"
// file usually put in ".github" directory.
type Config struct {
	// SelfHostedRunner is configuration for self-hosted runner.
	SelfHostedRunner struct {
		// Labels is label names for self-hosted runner.
		Labels []string `yaml:"labels"`
	} `yaml:"self-hosted-runner"`
}

func parseConfig(b []byte, path string) (*Config, error) {
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		msg := strings.ReplaceAll(err.Error(), "\n", " ")
		return nil, fmt.Errorf("could not parse config file %q: %s", path, msg)
	}
	return &c, nil
}

func readConfigFile(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file %q: %w", path, err)
	}
	return parseConfig(b, path)
}

func writeDefaultConfigFile(path string) error {
	b := []byte(`self-hosted-runner:
  # Labels of self-hosted runner in array of string
  labels: []
`)
	if err := os.WriteFile(path, b, 0644); err != nil {
		return fmt.Errorf("could not write default configuration file at %q: %w", path, err)
	}
	return nil
}
