// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"testing"
)

func TestRemoveReadOnlyFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		input    string
		want     string
	}{
		{
			testName: "empty JSON",
			input:    "{}",
			want:     "{}",
		},
		{
			testName: "single field",
			input:    `{ "key": 42 }`,
			want:     `{"key":42}`,
		},
		{
			testName: "with read-only field",
			input:    "{\"unifiedAlerting\": {\"enabled\": true}, \"plugins\": {\"pluginAdminEnabled\" :false}}",
			want:     "{\"unifiedAlerting\":{\"enabled\":true}}",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			if got, want := RemoveReadOnlyFields(testCase.input, `"plugins"`), testCase.want; got != want {
				t.Errorf("RemoveReadOnlyFields(%q) = %q, want %q", testCase.input, got, want)
			}
		})
	}
}
