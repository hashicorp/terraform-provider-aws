// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cty_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tfcty "github.com/hashicorp/terraform-provider-aws/internal/cty"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestGetPrimitives(t *testing.T) {
	t.Parallel()

	type A struct {
		Bool       types.Bool                                 `tfsdk:"bool"`
		Float32    types.Float32                              `tfsdk:"float32"`
		Float64    types.Float64                              `tfsdk:"float64"`
		Int32      types.Int32                                `tfsdk:"int32"`
		Int64      types.Int64                                `tfsdk:"int64"`
		String1    types.String                               `tfsdk:"string1"`
		String2    types.String                               `tfsdk:"string2"`
		StringEnum fwtypes.StringEnum[awstypes.AclPermission] `tfsdk:"string_enum"`
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
				"string1": cty.StringVal("Alice"),
			}),
			wantErr: true,
		},
		"source object, string value target": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string1": cty.StringVal("Alice"),
			}),
			target:  "not a pointer",
			wantErr: true,
		},
		"source object, string pointer target": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string1": cty.StringVal("Alice"),
			}),
			target:  new(string),
			wantErr: true,
		},
		"source object, nil struct pointer target": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string1": cty.StringVal("Alice"),
			}),
			target:  (*A)(nil),
			wantErr: true,
		},
		"source object, struct pointer target, one string field": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string1": cty.StringVal("Alice"),
			}),
			target: &A{},
			wantTarget: &A{
				String1: types.StringValue("Alice"),
			},
		},
		"source object, struct pointer target, two string fields": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string1": cty.StringVal("Alice"),
				"string2": cty.StringVal("Beagle"),
			}),
			target: &A{},
			wantTarget: &A{
				String1: types.StringValue("Alice"),
				String2: types.StringValue("Beagle"),
			},
		},
		"source object, struct pointer target, all fields": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string1":     cty.StringVal("Alice"),
				"string2":     cty.NullVal(cty.String),
				"bool":        cty.BoolVal(true),
				"float32":     cty.UnknownVal(cty.Number),
				"float64":     cty.NumberFloatVal(-64.64),
				"int32":       cty.NumberIntVal(32),
				"int64":       cty.NumberIntVal(-64),
				"string_enum": cty.StringVal(string(awstypes.AclPermissionRead)),
			}),
			target: &A{},
			wantTarget: &A{
				String1:    types.StringValue("Alice"),
				String2:    types.StringNull(),
				Bool:       types.BoolValue(true),
				Float32:    types.Float32Unknown(),
				Float64:    types.Float64Value(-64.64),
				Int32:      types.Int32Value(32),
				Int64:      types.Int64Value(-64),
				StringEnum: fwtypes.StringEnumValue(awstypes.AclPermissionRead),
			},
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
