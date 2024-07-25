// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package jsoncmp_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
)

func TestDiff(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		x, y     string
		wantDiff bool
	}{
		{
			testName: "no diff",
			x:        `{"A": "test1", "D": ["test2", "test3"], "C": {"A": true}, "B": 42}`,
			y:        `{"A":"test1", "B":42, "C":{"A":true}, "D": ["test2", "test3"]}`,
		},
		{
			testName: "has diff",
			x:        `{"A": "test1", "D": ["test2", "test3"], "C": {"A": true}, "B": 42}`,
			y:        `{"A":"test1", "B":41, "C":{"A":true}, "D": ["test3"]}`,
			wantDiff: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			output := jsoncmp.Diff(testCase.x, testCase.y)
			if got, want := output != "", testCase.wantDiff; got != want {
				t.Errorf("Diff(%s, %s) = %t, want %t", testCase.x, testCase.y, got, want)
			}
		})
	}
}
