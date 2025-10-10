// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dns

import (
	"testing"
)

func Test_normalizeCasingAndEscapeCodes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input, output string
	}{
		// Preserve escape code
		{"a\\000c.example.com", "a\\000c.example.com"},
		{"a\\056c.example.com", "a\\056c.example.com"}, // with escaped "."

		// Preserve "-" / "_" as-is
		{"a-b.example.com", "a-b.example.com"},
		{"_abc.example.com", "_abc.example.com"},

		// no conversion
		{"www.example.com", "www.example.com"},

		// converted into lower-case
		{"AbC.example.com", "abc.example.com"},

		// convert into escape code
		{"*.example.com", "\\052.example.com"},
		{"!.example.com", "\\041.example.com"},
		{"a/b.example.com", "a\\057b.example.com"},
		{"/.example.com", "\\057.example.com"},
		{"~.example.com", "\\176.example.com"},
		{"a\\2B.example.com", "a\\1342b.example.com"},
	}

	for _, tc := range cases {
		output := normalizeCasingAndEscapeCodes(tc.input)

		if got, want := output, tc.output; got != want {
			t.Errorf("normalizeCasingAndEscapeCodes(%q) = %v, want %v", tc.input, got, want)
		}
	}
}

func TestNormalize(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input  string
		output string
	}{
		{"example.com", "example.com"},
		{"example.com.", "example.com"},
		{"www.example.com.", "www.example.com"},
		{"www.ExAmPlE.COM.", "www.example.com"},
		{"", ""},
		{".", "."},
	}

	for _, tc := range cases {
		output := Normalize(tc.input)

		if got, want := output, tc.output; got != want {
			t.Errorf("Normalize(%q) = %v, want %v", tc.input, got, want)
		}
	}
}
