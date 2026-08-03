// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-version"
)

func TestVersionDecrementMinor(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		input         *version.Version
		expected      *version.Version
		expectedError error
	}{
		"valid": {
			input:    version.Must(version.NewVersion("6.10.3")),
			expected: version.Must(version.NewVersion("6.9.0")),
		},
		"minor is zero": {
			input:         version.Must(version.NewVersion("6.0.2")),
			expectedError: fmt.Errorf("minor version is zero, cannot decrement: %s", "6.0.2"),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, err := VersionDecrementMinor(tc.input)
			if tc.expectedError == nil {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error but got none")
				} else if err.Error() != tc.expectedError.Error() {
					t.Fatalf("unexpected error: got %s, want %s", err, tc.expectedError)
				}
			}

			if !result.Equal(tc.expected) {
				t.Errorf("unexpected result: got %v, want %v", result, tc.expected)
			}
		})
	}
}
