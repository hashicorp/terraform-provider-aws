// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Unit tests for observability helper functions. These tests do not require AWS
// credentials and can be run with `go test ./internal/service/bedrockagentcore/ -run Unit`.

package bedrockagentcore

import (
	"testing"
)

func TestUnitIsOtelEnvVar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  string
		want bool
	}{
		{"AGENT_OBSERVABILITY_ENABLED", true},
		{"OTEL_EXPORTER_OTLP_LOGS_ENDPOINT", true},
		{"OTEL_EXPORTER_OTLP_LOGS_HEADERS", true},
		{"OTEL_EXPORTER_OTLP_PROTOCOL", true},
		{"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", true},
		{"OTEL_RESOURCE_ATTRIBUTES", true},
		{"OTEL_TRACES_EXPORTER", true},
		{"OTEL_PYTHON_CONFIGURATOR", true},
		{"OTEL_PYTHON_DISTRO", true},
		{"OTEL_PYTHON_LOGGING_AUTO_INSTRUMENTATION_ENABLED", true},
		// Non-OTEL keys must not be removed.
		{"MY_APP_SECRET", false},
		{"LOG_LEVEL", false},
		{"OTEL_CUSTOM_VAR", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			t.Parallel()
			if got := isOtelEnvVar(tt.key); got != tt.want {
				t.Errorf("isOtelEnvVar(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestUnitXraySamplingRuleName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		runtimeID string
		want      string
	}{
		{"abc123", "bedrock-agentcore-abc123"},
		// "bedrock-agentcore-runtime-xyz-456" = 33 chars → truncated to 32
		{"runtime-xyz-456", "bedrock-agentcore-runtime-xyz-45"},
		{"my_runtime_01", "bedrock-agentcore-my_runtime_01"},
		// Long IDs are truncated to 32 chars total
		{"thisIsAVeryLongRuntimeIDThatExceedsLimit", "bedrock-agentcore-thisIsAVeryLon"},
	}

	for _, tt := range tests {
		t.Run(tt.runtimeID, func(t *testing.T) {
			t.Parallel()
			if got := xraySamplingRuleName(tt.runtimeID); got != tt.want {
				t.Errorf("xraySamplingRuleName(%q) = %q, want %q", tt.runtimeID, got, tt.want)
			}
		})
	}
}
