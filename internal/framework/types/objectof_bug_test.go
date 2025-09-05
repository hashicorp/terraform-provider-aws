// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// Test enum for reproducing CustomType equality issues
type testEnum string

const (
	testEnumValue1 testEnum = "value1"
	testEnumValue2 testEnum = "value2"
)

// Values implements the enum.Valueser interface
func (testEnum) Values() []testEnum {
	return []testEnum{
		testEnumValue1,
		testEnumValue2,
	}
}

// Test models that reproduce the deep nesting CustomType equality bug

// Innermost model with fwtypes.StringEnum (causes CustomType equality issue)
type testInnerModel struct {
	EnumField fwtypes.StringEnum[testEnum] `tfsdk:"enum_field"`
}

// Middle model containing nested custom types
type testMiddleModelA struct {
	Behavior    fwtypes.StringEnum[testEnum]                    `tfsdk:"behavior"`
	StringField types.String                                    `tfsdk:"string_field"`
	IntField    types.Int64                                     `tfsdk:"int_field"`
	InnerNested fwtypes.ListNestedObjectValueOf[testInnerModel] `tfsdk:"inner_nested"`
}

// Innermost model with types.List (also causes CustomType equality issue)
type testInnerModelB struct {
	Region    types.String `tfsdk:"region"`
	ListField types.List   `tfsdk:"list_field"`
}

// Middle model containing types.List nested structure
type testMiddleModelB struct {
	StringField types.String                                     `tfsdk:"string_field"`
	IntField    types.Int64                                      `tfsdk:"int_field"`
	InnerNested fwtypes.ListNestedObjectValueOf[testInnerModelB] `tfsdk:"inner_nested"`
}

// Outer model containing both problematic middle models
type testOuterModel struct {
	Name        types.String                                      `tfsdk:"name"`
	Description types.String                                      `tfsdk:"description"`
	NestedA     fwtypes.ListNestedObjectValueOf[testMiddleModelA] `tfsdk:"nested_a"`
	NestedB     fwtypes.ListNestedObjectValueOf[testMiddleModelB] `tfsdk:"nested_b"`
}

// Test that reproduces the CustomType equality bug with deep nesting
func TestObjectTypeOf_DeepNestedCustomTypeEquality(t *testing.T) {
	ctx := context.Background()

	// Two calls should create equal types
	type1 := fwtypes.NewListNestedObjectTypeOf[testOuterModel](ctx)
	type2 := fwtypes.NewListNestedObjectTypeOf[testOuterModel](ctx)

	if !type1.Equal(type2) {
		t.Error("Structurally identical deeply nested types should be equal")
	}
}

// Test that reproduces the StringEnum nesting failure
func TestObjectTypeOf_NestedStringEnum(t *testing.T) {
	ctx := context.Background()

	type1 := fwtypes.NewListNestedObjectTypeOf[testMiddleModelA](ctx)
	type2 := fwtypes.NewListNestedObjectTypeOf[testMiddleModelA](ctx)

	if !type1.Equal(type2) {
		t.Error("Nested StringEnum types should be equal")
	}
}

// Test that reproduces the types.List nesting failure
func TestObjectTypeOf_NestedTypesList(t *testing.T) {
	ctx := context.Background()

	type1 := fwtypes.NewListNestedObjectTypeOf[testMiddleModelB](ctx)
	type2 := fwtypes.NewListNestedObjectTypeOf[testMiddleModelB](ctx)

	if !type1.Equal(type2) {
		t.Error("Nested types.List should be equal")
	}
}

// Control test: simple nested types should work fine
func TestObjectTypeOf_SimpleNested(t *testing.T) {
	ctx := context.Background()

	type simpleInnerModel struct {
		Field types.String `tfsdk:"field"`
	}
	type simpleOuterModel struct {
		Nested fwtypes.ListNestedObjectValueOf[simpleInnerModel] `tfsdk:"nested"`
	}

	type1 := fwtypes.NewListNestedObjectTypeOf[simpleOuterModel](ctx)
	type2 := fwtypes.NewListNestedObjectTypeOf[simpleOuterModel](ctx)

	if !type1.Equal(type2) {
		t.Error("Simple nested types should be equal")
	}
}

// Control test: direct custom types should work
func TestObjectTypeOf_DirectCustomType(t *testing.T) {
	ctx := context.Background()

	type directCustomModel struct {
		EnumField fwtypes.StringEnum[testEnum] `tfsdk:"enum_field"`
	}

	type1 := fwtypes.NewListNestedObjectTypeOf[directCustomModel](ctx)
	type2 := fwtypes.NewListNestedObjectTypeOf[directCustomModel](ctx)

	if !type1.Equal(type2) {
		t.Error("Direct custom types should be equal")
	}
}
