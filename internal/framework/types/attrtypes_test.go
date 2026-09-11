// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type attributeTypesTestStruct1 struct{}

type attributeTypesTestStruct2 struct {
	ARN             types.String `tfsdk:"arn"`
	ID              types.Int64  `tfsdk:"id"`
	IncludeProperty types.Bool   `tfsdk:"include_property"`
}

type attributeTypesTestStruct3 struct {
	F1 types.String `tfsdk:"f1"`
	attributeTypesTestStruct2
	F2 types.Int32 `tfsdk:"f2"`
}

type attributeTypesTestStructExcludedField struct {
	Name     types.String `tfsdk:"name"`
	Excluded types.String `tfsdk:"-"`
}

type attributeTypesTestStructMissingTag struct {
	Name types.String
}

type attributeTypesTestStructNonAttrValue struct {
	Name    types.String `tfsdk:"name"`
	NotAttr string       `tfsdk:"not_attr"`
}

type attributeTypesTestInnerModel struct {
	Value types.String `tfsdk:"value"`
}

type attributeTypesTestNestedModel struct {
	Name   types.String                                                  `tfsdk:"name"`
	Nested fwtypes.ObjectValueOf[attributeTypesTestInnerModel]           `tfsdk:"nested"`
	List   fwtypes.ListNestedObjectValueOf[attributeTypesTestInnerModel] `tfsdk:"list"`
}

func TestAttributeTypes(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		attributeTypes func(context.Context) (map[string]attr.Type, diag.Diagnostics)
		expectErr      bool
		expected       map[string]attr.Type
	}{
		"empty struct": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestStruct1],
			expected:       map[string]attr.Type{},
		},
		"non-struct type": {
			attributeTypes: fwtypes.AttributeTypes[int],
			expectErr:      true,
		},
		"missing tfsdk tag": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestStructMissingTag],
			expectErr:      true,
		},
		"flat struct": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestStruct2],
			expected: map[string]attr.Type{
				"arn":              types.StringType,
				"id":               types.Int64Type,
				"include_property": types.BoolType,
			},
		},
		"pointer to struct": {
			attributeTypes: fwtypes.AttributeTypes[*attributeTypesTestStruct2],
			expected: map[string]attr.Type{
				"arn":              types.StringType,
				"id":               types.Int64Type,
				"include_property": types.BoolType,
			},
		},
		"embedded struct": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestStruct3],
			expected: map[string]attr.Type{
				"f1":               types.StringType,
				"arn":              types.StringType,
				"id":               types.Int64Type,
				"include_property": types.BoolType,
				"f2":               types.Int32Type,
			},
		},
		"excluded field": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestStructExcludedField],
			expected: map[string]attr.Type{
				"name": types.StringType,
			},
		},
		"non-attr.Value field skipped": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestStructNonAttrValue],
			expected: map[string]attr.Type{
				"name": types.StringType,
			},
		},
		"nested framework value fields": {
			attributeTypes: fwtypes.AttributeTypes[attributeTypesTestNestedModel],
			expected: map[string]attr.Type{
				"name":   types.StringType,
				"nested": fwtypes.NewObjectTypeOf[attributeTypesTestInnerModel](context.Background()),
				"list":   fwtypes.NewListNestedObjectTypeOf[attributeTypesTestInnerModel](context.Background()),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			got, diags := testCase.attributeTypes(ctx)

			if got, want := diags.HasError(), testCase.expectErr; got != want {
				t.Fatalf("expectErr = %t, got diagnostics: %v", want, diags)
			}
			if testCase.expectErr {
				return
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected diff (+expected, -got): %s", diff)
			}
		})
	}
}

func BenchmarkAttributeTypes(b *testing.B) {
	ctx := b.Context()

	b.Run("flat", func(b *testing.B) {
		for b.Loop() {
			fwtypes.AttributeTypesMust[attributeTypesTestStruct2](ctx)
		}
	})

	b.Run("nested", func(b *testing.B) {
		for b.Loop() {
			fwtypes.AttributeTypesMust[attributeTypesTestNestedModel](ctx)
		}
	})
}
