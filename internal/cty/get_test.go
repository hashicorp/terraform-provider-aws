// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cty_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tfcty "github.com/hashicorp/terraform-provider-aws/internal/cty"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestGetFrameworkPrimitives(t *testing.T) {
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
		"source null object, struct pointer target, one string field": {
			source: cty.NullVal(cty.Object(map[string]cty.Type{
				"string1": cty.String,
			})),
			target:  &A{},
			wantErr: true,
		},
		"source unknown object, struct pointer target, one string field": {
			source: cty.UnknownVal(cty.Object(map[string]cty.Type{
				"string1": cty.String,
			})),
			target:  &A{},
			wantErr: true,
		},
		"source object, struct pointer target, one null string field": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string1": cty.NullVal(cty.String),
			}),
			target: &A{},
			wantTarget: &A{
				String1: types.StringNull(),
			},
		},
		"source object, struct pointer target, one unknown string field": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string1": cty.UnknownVal(cty.String),
			}),
			target: &A{},
			wantTarget: &A{
				String1: types.StringUnknown(),
			},
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

			err := tfcty.GetFramework(ctx, testCase.source, testCase.target)
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

func TestGetFrameworkAggregatePrimitives(t *testing.T) {
	t.Parallel()

	type A struct {
		Int64List      fwtypes.ListOfInt64      `tfsdk:"int64_list"`
		StringList     fwtypes.ListOfString     `tfsdk:"string_list"`
		StringSet      fwtypes.SetOfString      `tfsdk:"string_set"`
		StringMap      fwtypes.MapOfString      `tfsdk:"string_map"`
		StringMapOfMap fwtypes.MapOfMapOfString `tfsdk:"string_map_of_map"`
	}

	ctx := t.Context()
	testCases := map[string]struct {
		source     cty.Value
		target     any
		wantErr    bool
		wantTarget any
	}{
		"one field populated": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string_list": cty.ListVal([]cty.Value{
					cty.StringVal("apple"),
					cty.StringVal("cherry"),
					cty.StringVal("kangaroo"),
				}),
				"int64_list":        cty.NullVal(cty.List(cty.Number)),
				"string_set":        cty.UnknownVal(cty.Set(cty.String)),
				"string_map":        cty.NullVal(cty.Map(cty.String)),
				"string_map_of_map": cty.NullVal(cty.Map(cty.Map(cty.String))),
			}),
			target: &A{},
			wantTarget: &A{
				StringList:     fwflex.FlattenFrameworkStringValueListOfString(ctx, []string{"apple", "cherry", "kangaroo"}),
				Int64List:      fwtypes.NewListValueOfNull[types.Int64](ctx),
				StringSet:      fwtypes.NewSetValueOfUnknown[types.String](ctx),
				StringMap:      fwtypes.NewMapValueOfNull[types.String](ctx),
				StringMapOfMap: fwtypes.NewMapValueOfNull[fwtypes.MapOfString](ctx),
			},
		},
		"all fields populated": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string_list": cty.ListVal([]cty.Value{
					cty.StringVal("apple"),
					cty.StringVal("cherry"),
					cty.StringVal("kangaroo"),
				}),
				"int64_list": cty.ListVal([]cty.Value{
					cty.NumberIntVal(-1),
					cty.NumberIntVal(1),
				}),
				"string_set": cty.SetVal([]cty.Value{
					cty.StringVal("ball"),
					cty.StringVal("rope"),
				}),
				"string_map": cty.MapVal(map[string]cty.Value{
					"foo": cty.StringVal("bar"),
					"baz": cty.StringVal("qux"),
				}),
				"string_map_of_map": cty.MapVal(map[string]cty.Value{
					"key1": cty.MapVal(map[string]cty.Value{
						"key2": cty.StringVal("val"),
					}),
				}),
			}),
			target: &A{},
			wantTarget: &A{
				StringList: fwflex.FlattenFrameworkStringValueListOfString(ctx, []string{"apple", "cherry", "kangaroo"}),
				Int64List: fwtypes.NewListValueOfMust[types.Int64](ctx, []attr.Value{
					types.Int64Value(-1),
					types.Int64Value(1),
				}),
				StringSet: fwflex.FlattenFrameworkStringValueSetOfString(ctx, []string{"ball", "rope"}),
				StringMap: fwflex.FlattenFrameworkStringValueMapOfString(ctx, map[string]string{
					"foo": "bar",
					"baz": "qux",
				}),
				StringMapOfMap: fwflex.FlattenFrameworkStringValueMapOfMapOfString(ctx, map[string]map[string]string{
					"key1": {
						"key2": "val",
					},
				}),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tfcty.GetFramework(ctx, testCase.source, testCase.target)
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

func TestGetFrameworkSimpleNestedObject(t *testing.T) {
	t.Parallel()

	type B struct {
		Bool types.Bool `tfsdk:"bool"`
	}
	type A struct {
		String types.String                       `tfsdk:"string"`
		B      fwtypes.ListNestedObjectValueOf[B] `tfsdk:"b"`
	}

	ctx := t.Context()
	testCases := map[string]struct {
		source     cty.Value
		target     any
		wantErr    bool
		wantTarget any
	}{
		"null nested object": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("Alice"),
				"b":      cty.NullVal(cty.List(cty.Object(map[string]cty.Type{"bool": cty.Bool}))),
			}),
			target: &A{},
			wantTarget: &A{
				String: types.StringValue("Alice"),
				B:      fwtypes.NewListNestedObjectValueOfNull[B](ctx),
			},
		},
		"unknown nested object": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("Alice"),
				"b":      cty.UnknownVal(cty.List(cty.Object(map[string]cty.Type{"bool": cty.Bool}))),
			}),
			target: &A{},
			wantTarget: &A{
				String: types.StringValue("Alice"),
				B:      fwtypes.NewListNestedObjectValueOfUnknown[B](ctx),
			},
		},
		"known nested object": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("Alice"),
				"b": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
					"bool": cty.BoolVal(true),
				})}),
			}),
			target: &A{},
			wantTarget: &A{
				String: types.StringValue("Alice"),
				B:      fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &B{Bool: types.BoolValue(true)}),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tfcty.GetFramework(ctx, testCase.source, testCase.target)
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

func TestGetFrameworkComplexNestedObject(t *testing.T) {
	t.Parallel()

	type C struct {
		Bool      types.Bool          `tfsdk:"bool"`
		StringMap fwtypes.MapOfString `tfsdk:"string_map"`
	}
	type B struct {
		C fwtypes.SetNestedObjectValueOf[C] `tfsdk:"c"`
	}
	type A struct {
		String types.String                       `tfsdk:"string"`
		B      fwtypes.ListNestedObjectValueOf[B] `tfsdk:"b"`
	}

	ctx := t.Context()
	testCases := map[string]struct {
		source     cty.Value
		target     any
		wantErr    bool
		wantTarget any
	}{
		"null in nested object": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("Alice"),
				"b": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
					"c": cty.NullVal(cty.Set(cty.Object(map[string]cty.Type{
						"bool":       cty.Bool,
						"string_map": cty.Map(cty.String),
					}))),
				})}),
			}),
			target: &A{},
			wantTarget: &A{
				String: types.StringValue("Alice"),
				B:      fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &B{C: fwtypes.NewSetNestedObjectValueOfNull[C](ctx)}),
			},
		},
		"fully known nested object": {
			source: cty.ObjectVal(map[string]cty.Value{
				"string": cty.StringVal("Alice"),
				"b": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
					"c": cty.SetVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
						"bool":       cty.BoolVal(true),
						"string_map": cty.MapVal(map[string]cty.Value{"key": cty.StringVal("val")}),
					})}),
				})}),
			}),
			target: &A{},
			wantTarget: &A{
				String: types.StringValue("Alice"),
				B: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &B{C: fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &C{
					Bool:      types.BoolValue(true),
					StringMap: fwflex.FlattenFrameworkStringValueMapOfString(ctx, map[string]string{"key": "val"}),
				})}),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := tfcty.GetFramework(ctx, testCase.source, testCase.target)
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
