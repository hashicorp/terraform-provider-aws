// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import "testing"

func TestUserEmailFilter(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input string
		want  string
	}{
		"plain":               {input: "user@example.com", want: `email = "user@example.com"`},
		"quotation mark":      {input: `user"tag@example.com`, want: `email = "user\"tag@example.com"`},
		"backslash":           {input: `user\tag@example.com`, want: `email = "user\\tag@example.com"`},
		"backslash and quote": {input: `user\"tag@example.com`, want: `email = "user\\\"tag@example.com"`},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := userEmailFilter(testCase.input); got != testCase.want {
				t.Fatalf("userEmailFilter(%q) = %q, want %q", testCase.input, got, testCase.want)
			}
		})
	}
}
