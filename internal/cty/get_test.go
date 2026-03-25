// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cty_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tfcty "github.com/hashicorp/terraform-provider-aws/internal/cty"
)

func TestGetPrimitives(t *testing.T) {
	t.Parallel()

	type A struct {
		Bool   types.Bool   `tfsdk:"bool"`
		Int32  types.Int32  `tfsdk:"int32"`
		Int64  types.Int64  `tfsdk:"int64"`
		String types.String `tfsdk:"string"`
	}

	ctx := t.Context()
	testCases := map[string]struct {
		source     cty.Value
		target     any
		wantErr    bool
		wantTarget any
	}{
		"source string, nil target": {
			source:  cty.StringVal("test"),
			wantErr: true,
		},
		"source list, nil target": {
			source: cty.ListVal([]cty.Value{
				cty.StringVal("apple"),
				cty.StringVal("cherry"),
				cty.StringVal("kangaroo"),
			}),
			wantErr: true,
		},
		"source object, nil target": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("Alice"),
			}),
			wantErr: true,
		},
		"source object, string value target": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("Alice"),
			}),
			target:  "not a pointer",
			wantErr: true,
		},
		"source object, string pointer target": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("Alice"),
			}),
			target:  new(string),
			wantErr: true,
		},
		"source object, nil struct pointer target": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("Alice"),
			}),
			target:     new(A),
			wantTarget: new(A),
		},
		"source object, struct pointer target": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("Alice"),
			}),
			target:     &A{},
			wantTarget: &A{String: types.StringValue("Alice")},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tfcty.Get(ctx, testCase.source, testCase.target)
			gotErr := err != nil

			if gotErr != testCase.wantErr {
				t.Errorf("gotErr = %v, wantErr = %v", gotErr, testCase.wantErr)
			}

			if gotErr {
				if !testCase.wantErr {
					t.Errorf("err = %q", err)
				}
			} else if diff := cmp.Diff(testCase.target, testCase.wantTarget); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
