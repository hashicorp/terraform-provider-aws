// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"testing"
)

func TestReverseDNS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "amazonaws.com",
			input:    "amazonaws.com",
			expected: "com.amazonaws",
		},
		{
			name:     "amazonaws.com.cn",
			input:    "amazonaws.com.cn",
			expected: "cn.com.amazonaws",
		},
		{
			name:     "sc2s.sgov.gov",
			input:    "sc2s.sgov.gov",
			expected: "gov.sgov.sc2s",
		},
		{
			name:     "c2s.ic.gov",
			input:    "c2s.ic.gov",
			expected: "gov.ic.c2s",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got, want := ReverseDNS(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %s, expected: %s", got, want)
			}
		})
	}
}
