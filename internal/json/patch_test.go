// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package json_test

import (
	"cmp"
	"slices"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	mattbairdjsonpatch "github.com/mattbaird/jsonpatch"
)

func TestCreatePatchFromStrings(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName  string
		a, b      string
		wantPatch []mattbairdjsonpatch.JsonPatchOperation
		wantErr   bool
	}{
		{
			testName: "invalid JSON",
			a:        `test`,
			b:        `{}`,
			wantErr:  true,
		},
		{
			testName:  "empty patch, empty JSON",
			a:         `{}`,
			b:         `{}`,
			wantPatch: []mattbairdjsonpatch.JsonPatchOperation{},
		},
		{
			testName:  "empty patch, non-empty JSON",
			a:         `{"A": "test1", "B": 42}`,
			b:         `{"B": 42, "A": "test1"}`,
			wantPatch: []mattbairdjsonpatch.JsonPatchOperation{},
		},
		{
			testName: "from empty JSON",
			a:        `{}`,
			b:        `{"A": "test1", "B": 42}`,
			wantPatch: []mattbairdjsonpatch.JsonPatchOperation{
				{Operation: "add", Path: "/A", Value: "test1"},
				{Operation: "add", Path: "/B", Value: float64(42)},
			},
		},
		{
			testName: "to empty JSON",
			a:        `{"A": "test1", "B": 42}`,
			b:        `{}`,
			wantPatch: []mattbairdjsonpatch.JsonPatchOperation{
				{Operation: "remove", Path: "/A"},
				{Operation: "remove", Path: "/B"},
			},
		},
		{
			testName: "change values",
			a:        `{"A": "test1", "B": 42}`,
			b:        `{"A": ["test2"], "B": false}`,
			wantPatch: []mattbairdjsonpatch.JsonPatchOperation{
				{Operation: "replace", Path: "/A", Value: []any{"test2"}},
				{Operation: "replace", Path: "/B", Value: false},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			got, err := tfjson.CreatePatchFromStrings(testCase.a, testCase.b)
			if got, want := err != nil, testCase.wantErr; !gocmp.Equal(got, want) {
				t.Errorf("CreatePatchFromStrings(%s, %s) err %t, want %t", testCase.a, testCase.b, got, want)
			}
			if err == nil {
				sortTransformer := gocmp.Transformer("SortPatchOps", func(ops []mattbairdjsonpatch.JsonPatchOperation) []mattbairdjsonpatch.JsonPatchOperation {
					sorted := make([]mattbairdjsonpatch.JsonPatchOperation, len(ops))
					copy(sorted, ops)
					slices.SortFunc(sorted, func(a, b mattbairdjsonpatch.JsonPatchOperation) int {
						return cmp.Or(cmp.Compare(a.Operation, b.Operation), cmp.Compare(a.Path, b.Path))
					})
					return sorted
				})

				if diff := gocmp.Diff(got, testCase.wantPatch, sortTransformer); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}
