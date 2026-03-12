// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package json_test

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestTopLevelKeys(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		input    string
		want     []string
	}{
		{
			testName: "empty JSON",
			input:    "{}",
		},
		{
			testName: "single field",
			input:    `{ "key": 42 }`,
			want:     []string{`"key"`},
		},
		{
			testName: "multiple fields",
			input:    `{"key1": {"value1": 42}, "key2": false, "key3": ["value3"], "key4": {"value4": {"key44": 44"}}}`,
			want:     []string{`"key1"`, `"key2"`, `"key3"`, `"key4"`},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			got := slices.Collect(tfjson.TopLevelKeys(testCase.input))
			if diff := cmp.Diff(got, testCase.want); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
