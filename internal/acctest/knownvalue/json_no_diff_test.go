// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package knownvalue_test

import (
	"testing"

	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
)

func TestJSONNoDiff_CheckValue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		value    string // the expected JSON passed to JSONNoDiff
		other    any    // the actual value passed to CheckValue
		wantErr  bool
	}{
		{
			testName: "identical JSON",
			value:    `{"A": "test1", "B": 42, "C": {"A": true}, "D": ["test2", "test3"]}`,
			other:    `{"A": "test1", "B": 42, "C": {"A": true}, "D": ["test2", "test3"]}`,
		},
		{
			testName: "same JSON, different key order and whitespace",
			value:    `{"A": "test1", "D": ["test2", "test3"], "C": {"A": true}, "B": 42}`,
			other:    `{"A":"test1", "B":42, "C":{"A":true}, "D": ["test2", "test3"]}`,
		},
		{
			testName: "empty objects",
			value:    `{}`,
			other:    `{}`,
		},
		{
			testName: "differing scalar value",
			value:    `{"A": "test1", "B": 42, "C": {"A": true}, "D": ["test2", "test3"]}`,
			other:    `{"A": "test1", "B": 41, "C": {"A": true}, "D": ["test2", "test3"]}`,
			wantErr:  true,
		},
		{
			testName: "differing nested array",
			value:    `{"A": "test1", "B": 42, "C": {"A": true}, "D": ["test2", "test3"]}`,
			other:    `{"A": "test1", "B": 42, "C": {"A": true}, "D": ["test3"]}`,
			wantErr:  true,
		},
		{
			testName: "missing key",
			value:    `{"A": "test1", "B": 42}`,
			other:    `{"A": "test1"}`,
			wantErr:  true,
		},
		{
			testName: "extra key",
			value:    `{"A": "test1"}`,
			other:    `{"A": "test1", "B": 42}`,
			wantErr:  true,
		},
		{
			testName: "non-string input (int)",
			value:    `{"A": "test1"}`,
			other:    42,
			wantErr:  true,
		},
		{
			testName: "non-string input (nil)",
			value:    `{"A": "test1"}`,
			other:    nil,
			wantErr:  true,
		},
		{
			testName: "invalid actual JSON",
			value:    `{"A": "test1"}`,
			other:    `not json`,
			wantErr:  true,
		},
		{
			testName: "invalid expected JSON",
			value:    `not json`,
			other:    `{"A": "test1"}`,
			wantErr:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			err := tfknownvalue.JSONNoDiff(testCase.value).CheckValue(testCase.other)

			if got, want := err != nil, testCase.wantErr; got != want {
				t.Errorf("CheckValue() error = %v, wantErr %t", err, want)
			}
		})
	}
}
