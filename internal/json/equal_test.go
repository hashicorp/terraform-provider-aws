// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestEqualStrings(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName  string
		x, y      string
		wantEqual bool
	}{
		{
			testName: "invalid JSON first",
			x:        `test`,
			y:        `{}`,
		},
		{
			testName: "invalid JSON second",
			x:        `{}`,
			y:        `test`,
		},
		{
			testName:  "empty JSON",
			x:         `{}`,
			y:         ` { } `,
			wantEqual: true,
		},
		{
			testName: "empty JSON first, not empty second",
			x:        `{}`,
			y:        `{"A": "test"}`,
		},
		{
			testName: "not empty, equal",
			x:        `{"A": "test1", "D": ["test2", "test3"], "C": {"A": true}, "B": 42}`,
			y: `{
				"C": {"A": true},
				"D": ["test2", "test3"],
				"A":"test1","B":  42
			}`,
			wantEqual: true,
		},
		{
			testName: "not empty, not equal",
			x:        `{"A": "test1", "D": ["test2", "test3"], "C": {"A": true}, "B": 42}`,
			y: `{
				"C": {"A": true},
				"D": ["test2", "test3"],
				"A":"test4","B":  42
			}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			equal := json.EqualStrings(testCase.x, testCase.y)
			if got, want := equal, testCase.wantEqual; !cmp.Equal(got, want) {
				t.Errorf("EqualStrings(%s, %s) = %t, want %t", testCase.x, testCase.y, got, want)
			}
		})
	}
}
