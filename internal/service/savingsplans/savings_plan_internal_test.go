// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package savingsplans

import "testing"

func TestNormalizeCommitmentValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "preserves concise decimal",
			input:    "1.158",
			expected: "1.158",
		},
		{
			name:     "trims trailing zeros",
			input:    "1.15800000",
			expected: "1.158",
		},
		{
			name:     "removes redundant decimal part",
			input:    "1.00000000",
			expected: "1",
		},
		{
			name:     "handles leading zeros",
			input:    "0001.2300",
			expected: "1.23",
		},
		{
			name:     "handles zero values",
			input:    "0.00000000",
			expected: "0",
		},
		{
			name:     "leaves non-decimal input unchanged",
			input:    "abc",
			expected: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := normalizeCommitmentValue(tt.input); got != tt.expected {
				t.Errorf("normalizeCommitmentValue(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
