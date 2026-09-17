// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidIdentifier(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"a",
		"hello-world",
		"hello-world-0123456789",
		strings.Repeat("w", 63),
	}
	for _, v := range validNames {
		_, errors := validIdentifier(v, names.AttrName)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid DocumentDB Identifier: %q", v, errors)
		}
	}

	invalidNames := []string{
		"",
		"special@character",
		"slash/in-the-middle",
		"dot.in-the-middle",
		"two-hyphen--in-the-middle",
		"0-first-numeric",
		"-first-hyphen",
		"end-hyphen-",
		strings.Repeat("W", 64),
	}
	for _, v := range invalidNames {
		_, errors := validIdentifier(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid DocumentDB Identifier", v)
		}
	}
}

func TestValidParamGroupNamePrefix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		value    string
		errCount int
	}{
		{
			value: "valid-name",
		},
		{
			value:    "testing123!",
			errCount: 1,
		},
		{
			value:    "testing_123",
			errCount: 1,
		},
		{
			value:    "1testing123",
			errCount: 1,
		},
		{
			value:    "testing--123",
			errCount: 1,
		},
		{
			value:    strings.Repeat("w", 230),
			errCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validParamGroupNamePrefix(tc.value, names.AttrNamePrefix)

		if len(errors) != tc.errCount {
			t.Fatalf("Unexpected error count for value %q. got: %d, want: %d", tc.value, len(errors), tc.errCount)
		}
	}
}
