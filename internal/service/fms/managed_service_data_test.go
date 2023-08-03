// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms_test

import (
	"testing"

	tffms "github.com/hashicorp/terraform-provider-aws/internal/service/fms"
)

func TestRemoveEmptyFieldsFromJSON(t *testing.T) {
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
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			if got, want := tffms.RemoveEmptyFieldsFromJSON(testCase.input), testCase.want; got != want {
				t.Errorf("RemoveEmptyFieldsFromJSON(%q) = %q, want %q", testCase.input, got, want)
			}
		})
	}
}
