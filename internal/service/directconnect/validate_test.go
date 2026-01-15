// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"testing"
)

func TestValidLongASN(t *testing.T) {
	t.Parallel()

	validCases := []string{
		"1",
		"65000",
		"4200000000",
		"4294967294",
		"64086.59904",
		"1.1",
	}

	for _, v := range validCases {
		_, errors := validLongASN()(v, "bgp_asn_long")
		if len(errors) != 0 {
			t.Fatalf("%q should be valid: %v", v, errors)
		}
	}

	invalidCases := []string{
		"0",
		"4294967295",
		"invalid",
		"64086.59904.1",
		"-1",
	}

	for _, v := range invalidCases {
		_, errors := validLongASN()(v, "bgp_asn_long")
		if len(errors) == 0 {
			t.Fatalf("%q should be invalid", v)
		}
	}
}

func TestParseASN(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    string
		expected int64
	}{
		{"4200000000", 4200000000},
		{"65000", 65000},
		{"1", 1},
		{"64086.59904", 4200000000},
		{"1.1", 65537},
	}

	for _, tc := range cases {
		result, err := parseASN(tc.input)
		if err != nil {
			t.Fatalf("parseASN(%q) returned error: %v", tc.input, err)
		}
		if result != tc.expected {
			t.Fatalf("parseASN(%q) = %d, expected %d", tc.input, result, tc.expected)
		}
	}
}
