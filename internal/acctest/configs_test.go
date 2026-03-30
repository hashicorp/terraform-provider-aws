// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acctest_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestConfigRandomPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		overrides []string
		expected  string
	}{
		{
			name:      "Default configuration",
			overrides: nil,
			expected: `
ephemeral "aws_secretsmanager_random_password" "test" {
  password_length     = 20
  exclude_punctuation = true
}
`,
		},
		{
			name:      "Override password_length",
			overrides: []string{"password_length = 30"},
			expected: `
ephemeral "aws_secretsmanager_random_password" "test" {
  password_length     = 30
  exclude_punctuation = true
}
`,
		},
		{
			name:      "Override exclude_punctuation",
			overrides: []string{"exclude_punctuation = false"},
			expected: `
ephemeral "aws_secretsmanager_random_password" "test" {
  password_length     = 20
  exclude_punctuation = false
}
`,
		},
		{
			name:      "Multiple overrides",
			overrides: []string{"password_length = 25", "exclude_punctuation = false"},
			expected: `
ephemeral "aws_secretsmanager_random_password" "test" {
  password_length     = 25
  exclude_punctuation = false
}
`,
		},
		{
			name:      "Optional keys",
			overrides: []string{"exclude_characters = abcdef", "include_space = true"},
			expected: `
ephemeral "aws_secretsmanager_random_password" "test" {
  password_length     = 20
  exclude_punctuation = true
  exclude_characters = "abcdef"
  include_space = true
}
`,
		},
		{
			name:      "Exclude characters with quotes",
			overrides: []string{"exclude_characters = \"abcdef\"", "include_space = true"},
			expected: `
ephemeral "aws_secretsmanager_random_password" "test" {
  password_length     = 20
  exclude_punctuation = true
  exclude_characters = "abcdef"
  include_space = true
}
`,
		},
		{
			name:      "Exclude characters including a quote",
			overrides: []string{"exclude_characters = abc\"def", "include_space = true"},
			expected: `
ephemeral "aws_secretsmanager_random_password" "test" {
  password_length     = 20
  exclude_punctuation = true
  exclude_characters = "abc\"def"
  include_space = true
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := acctest.ConfigRandomPassword(tt.overrides...)
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tt.expected)

			if result != expected {
				t.Errorf("unexpected result for %s:\nGot:\n%s\n\nExpected:\n%s", tt.name, result, expected)
			}
		})
	}
}
