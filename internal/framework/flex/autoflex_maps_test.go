// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

// Tests AutoFlex's Expand/Flatten of maps.

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type tfComplexValue struct {
	Field1 types.String                                          `tfsdk:"field1"`
	Field2 fwtypes.ListNestedObjectValueOf[tfListOfNestedObject] `tfsdk:"field2"`
	Field3 types.Map                                             `tfsdk:"field3"`
	Field4 fwtypes.SetNestedObjectValueOf[tfSingleInt64Field]    `tfsdk:"field4"`
}

type awsComplexValue struct {
	Field1 string
	Field2 *awsNestedObjectPointer
	Field3 map[string]*string
	Field4 []awsSingleInt64Value
}

type tfMapOfString struct {
	FieldInner fwtypes.MapValueOf[basetypes.StringValue] `tfsdk:"field_inner"`
}

type tfMapOf[T attr.Value] struct {
	Field1 fwtypes.MapValueOf[T] `tfsdk:"field1"`
}

type tfNestedMapOfString struct {
	FieldOuter fwtypes.ListNestedObjectValueOf[tfMapOfString] `tfsdk:"field_outer"`
}

type awsNestedMapOfString struct {
	FieldOuter awsMapOfString
}

type tfMapOfMapOfString struct {
	Field1 fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]] `tfsdk:"field1"`
}

type tfMapOfMapOfInt32 struct {
	Field1 fwtypes.MapValueOf[fwtypes.MapValueOf[types.Int32]] `tfsdk:"field1"`
}

type awsMapOfString struct {
	FieldInner map[string]string
}

type awsMapOf[T any] struct {
	Field1 map[string]T
}

type awsMapOfMapOfString struct {
	Field1 map[string]map[string]string
}

type awsMapOfMapOfInt32 struct {
	Field1 map[string]map[string]int32
}

type awsMapOfMapOfStringPointer struct {
	Field1 map[string]map[string]*string
}

type awsMapOfMapOfInt32Pointer struct {
	Field1 map[string]map[string]*int32
}

type awsMapOfStringPointer struct {
	FieldInner map[string]*string
}

type tfMapOfListOfString struct {
	Field1 fwtypes.MapValueOf[fwtypes.ListValueOf[types.String]] `tfsdk:"field1"`
}

type awsMapOfListOfString struct {
	Field1 map[string][]string
}

type awsMapOfListOfInt32 struct {
	Field1 map[string][]int32
}

type awsMapOfInt struct {
	Field1 map[string]int
}

type tfSingleMapField struct {
	Field1 types.Map `tfsdk:"field1"`
}

type awsMapOfStringPointerField struct {
	Field1 map[string]*string
}

type tfMapStringPointerBlockElement struct {
	MapBlockKey types.String `tfsdk:"map_block_key"`
}

type tfMapStringPointerBlock struct {
	Field1 fwtypes.ListNestedObjectValueOf[tfMapStringPointerBlockElement] `tfsdk:"field1"`
}

type tfMapBlockList struct {
	MapBlock fwtypes.ListNestedObjectValueOf[tfMapBlockElement] `tfsdk:"map_block"`
}

type tfMapBlockSet struct {
	MapBlock fwtypes.SetNestedObjectValueOf[tfMapBlockElement] `tfsdk:"map_block"`
}

type awsMapBlockValues struct {
	MapBlock map[string]awsMapBlockElement
}

type awsMapBlockPointers struct {
	MapBlock map[string]*awsMapBlockElement
}

type tfMapBlockElement struct {
	MapBlockKey types.String `tfsdk:"map_block_key"`
	Attr1       types.String `tfsdk:"attr1"`
	Attr2       types.String `tfsdk:"attr2"`
}

type awsMapBlockElement struct {
	Attr1 string
	Attr2 string
}

type tfMapBlockListEnumKey struct {
	MapBlock fwtypes.ListNestedObjectValueOf[tfMapBlockElementEnumKey] `tfsdk:"map_block"`
}

type tfMapBlockElementEnumKey struct {
	MapBlockKey fwtypes.StringEnum[testEnum] `tfsdk:"map_block_key"`
	Attr1       types.String                 `tfsdk:"attr1"`
	Attr2       types.String                 `tfsdk:"attr2"`
}

type tfMapBlockListNoKey struct {
	MapBlock fwtypes.ListNestedObjectValueOf[tfMapBlockElementNoKey] `tfsdk:"map_block"`
}

type tfMapBlockElementNoKey struct {
	Attr1 types.String `tfsdk:"attr1"`
	Attr2 types.String `tfsdk:"attr2"`
}

func TestExpandMaps(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"map of string": {
			Source: &tfMapOfString{
				FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			Target: &awsMapOfString{},
			WantTarget: &awsMapOfString{
				FieldInner: map[string]string{
					"x": "y",
				},
			},
		},
		"map of string pointer": {
			Source: &tfMapOfString{
				FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			Target: &awsMapOfStringPointer{},
			WantTarget: &awsMapOfStringPointer{
				FieldInner: map[string]*string{
					"x": aws.String("y"),
				},
			},
		},
		"map of list of string": {
			Source: &tfMapOfListOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.ListValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
						types.StringValue("y"),
						types.StringValue("z"),
					}),
				}),
			},
			Target: &awsMapOfListOfString{},
			WantTarget: &awsMapOfListOfString{
				Field1: map[string][]string{
					"x": {"y", "z"},
				},
			},
		},
		"map of map of string": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			Target: &awsMapOfMapOfString{},
			WantTarget: &awsMapOfMapOfString{
				Field1: map[string]map[string]string{
					"x": {
						"y": "z",
					},
				},
			},
		},
		"map of map of string pointer": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			Target: &awsMapOfMapOfStringPointer{},
			WantTarget: &awsMapOfMapOfStringPointer{
				Field1: map[string]map[string]*string{
					"x": {
						"y": aws.String("z"),
					},
				},
			},
		},
		"nested string map": {
			Source: &tfNestedMapOfString{
				FieldOuter: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfMapOfString{
					FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"x": types.StringValue("y"),
					}),
				}),
			},
			Target: &awsNestedMapOfString{},
			WantTarget: &awsNestedMapOfString{
				FieldOuter: awsMapOfString{
					FieldInner: map[string]string{
						"x": "y",
					},
				},
			},
		},
		"map of string to incompatible target": {
			Source: &tfSingleMapField{
				Field1: types.MapValueMust(types.StringType, map[string]attr.Value{"a": types.StringValue("b")}),
			},
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
		},
		"incompatible map value type": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			Target:     &awsMapOfInt{},
			WantTarget: &awsMapOfInt{},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{})
}

func TestExpandMapOfString(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			Target: &awsMapOf[string]{},
			WantTarget: &awsMapOf[string]{
				Field1: map[string]string{
					"x": "y",
				},
			},
		},
		"multiple values": {
			Source: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
					"y": types.StringValue("z"),
				}),
			},
			Target: &awsMapOf[string]{},
			WantTarget: &awsMapOf[string]{
				Field1: map[string]string{
					"x": "y",
					"y": "z",
				},
			},
		},
		"empty value": {
			Source: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{}),
			},
			Target: &awsMapOf[string]{},
			WantTarget: &awsMapOf[string]{
				Field1: map[string]string{},
			},
		},
		"null value": {
			Source: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfNull[types.String](ctx),
			},
			Target: &awsMapOf[string]{},
			WantTarget: &awsMapOf[string]{
				Field1: nil,
			},
		},
		"invalid conversion": {
			Source: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			Target:     &awsMapOf[int32]{},
			WantTarget: &awsMapOf[int32]{},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{})
}

func TestExpandMapOfStringPointer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			Target: &awsMapOf[*string]{},
			WantTarget: &awsMapOf[*string]{
				Field1: map[string]*string{
					"x": aws.String("y"),
				},
			},
		},
		"multiple values": {
			Source: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
					"y": types.StringValue("z"),
				}),
			},
			Target: &awsMapOf[*string]{},
			WantTarget: &awsMapOf[*string]{
				Field1: map[string]*string{
					"x": aws.String("y"),
					"y": aws.String("z"),
				},
			},
		},
		"empty value": {
			Source: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{}),
			},
			Target: &awsMapOf[*string]{},
			WantTarget: &awsMapOf[*string]{
				Field1: map[string]*string{},
			},
		},
		"null value": {
			Source: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfNull[types.String](ctx),
			},
			Target: &awsMapOf[*string]{},
			WantTarget: &awsMapOf[*string]{
				Field1: nil,
			},
		},
		"invalid conversion": {
			Source: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			Target:     &awsMapOf[int32]{},
			WantTarget: &awsMapOf[int32]{},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{})
}

func TestExpandMapOfInt32(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfMust[types.Int32](ctx, map[string]attr.Value{
					"x": types.Int32Value(42),
				}),
			},
			Target: &awsMapOf[int32]{},
			WantTarget: &awsMapOf[int32]{
				Field1: map[string]int32{
					"x": 42,
				},
			},
		},
		"multiple values": {
			Source: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfMust[types.Int32](ctx, map[string]attr.Value{
					"x": types.Int32Value(42),
					"y": types.Int32Value(43),
				}),
			},
			Target: &awsMapOf[int32]{},
			WantTarget: &awsMapOf[int32]{
				Field1: map[string]int32{
					"x": 42,
					"y": 43,
				},
			},
		},
		"empty value": {
			Source: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfMust[types.Int32](ctx, map[string]attr.Value{}),
			},
			Target: &awsMapOf[int32]{},
			WantTarget: &awsMapOf[int32]{
				Field1: map[string]int32{},
			},
		},
		"null value": {
			Source: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfNull[types.Int32](ctx),
			},
			Target: &awsMapOf[int32]{},
			WantTarget: &awsMapOf[int32]{
				Field1: nil,
			},
		},
		"invalid conversion": {
			Source: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfMust[types.Int32](ctx, map[string]attr.Value{
					"x": types.Int32Value(42),
				}),
			},
			Target:     &awsMapOf[string]{},
			WantTarget: &awsMapOf[string]{},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{})
}

func TestExpandMapOfInt32Pointer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfMust[types.Int32](ctx, map[string]attr.Value{
					"x": types.Int32Value(42),
				}),
			},
			Target: &awsMapOf[*int32]{},
			WantTarget: &awsMapOf[*int32]{
				Field1: map[string]*int32{
					"x": aws.Int32(42),
				},
			},
		},
		"multiple values": {
			Source: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfMust[types.Int32](ctx, map[string]attr.Value{
					"x": types.Int32Value(42),
					"y": types.Int32Value(43),
				}),
			},
			Target: &awsMapOf[*int32]{},
			WantTarget: &awsMapOf[*int32]{
				Field1: map[string]*int32{
					"x": aws.Int32(42),
					"y": aws.Int32(43),
				},
			},
		},
		"empty value": {
			Source: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfMust[types.Int32](ctx, map[string]attr.Value{}),
			},
			Target: &awsMapOf[*int32]{},
			WantTarget: &awsMapOf[*int32]{
				Field1: map[string]*int32{},
			},
		},
		"null value": {
			Source: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfNull[types.Int32](ctx),
			},
			Target: &awsMapOf[*int32]{},
			WantTarget: &awsMapOf[*int32]{
				Field1: nil,
			},
		},
		"invalid conversion": {
			Source: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfMust[types.Int32](ctx, map[string]attr.Value{
					"x": types.Int32Value(42),
				}),
			},
			Target:     &awsMapOf[string]{},
			WantTarget: &awsMapOf[string]{},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{})
}

func TestExpandMapOfListOfString(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &tfMapOfListOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.ListValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
						types.StringValue("y"),
						types.StringValue("z"),
					}),
				}),
			},
			Target: &awsMapOfListOfString{},
			WantTarget: &awsMapOfListOfString{
				Field1: map[string][]string{
					"x": {"y", "z"},
				},
			},
		},
		"multiple values": {
			Source: &tfMapOfListOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.ListValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
						types.StringValue("y"),
						types.StringValue("z"),
					}),
					"a": fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
						types.StringValue("b"),
						types.StringValue("c"),
					}),
				}),
			},
			Target: &awsMapOfListOfString{},
			WantTarget: &awsMapOfListOfString{
				Field1: map[string][]string{
					"x": {"y", "z"},
					"a": {"b", "c"},
				},
			},
		},
		"empty value": {
			Source: &tfMapOfListOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.ListValueOf[types.String]](ctx, map[string]attr.Value{}),
			},
			Target: &awsMapOfListOfString{},
			WantTarget: &awsMapOfListOfString{
				Field1: map[string][]string{},
			},
		},
		"null value": {
			Source: &tfMapOfListOfString{
				Field1: fwtypes.NewMapValueOfNull[fwtypes.ListValueOf[types.String]](ctx),
			},
			Target: &awsMapOfListOfString{},
			WantTarget: &awsMapOfListOfString{
				Field1: nil,
			},
		},
		"invalid conversion": {
			Source: &tfMapOfListOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.ListValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
						types.StringValue("y"),
					}),
				}),
			},
			Target:     &awsMapOfListOfInt32{},
			WantTarget: &awsMapOfListOfInt32{},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{})
}

func TestExpandMapOfMapOfString(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			Target: &awsMapOfMapOfString{},
			WantTarget: &awsMapOfMapOfString{
				Field1: map[string]map[string]string{
					"x": {
						"y": "z",
					},
				},
			},
		},
		"multiple values": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
					"a": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"b": types.StringValue("c"),
					}),
				}),
			},
			Target: &awsMapOfMapOfString{},
			WantTarget: &awsMapOfMapOfString{
				Field1: map[string]map[string]string{
					"x": {
						"y": "z",
					},
					"a": {
						"b": "c",
					},
				},
			},
		},
		"empty value": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{}),
			},
			Target: &awsMapOfMapOfString{},
			WantTarget: &awsMapOfMapOfString{
				Field1: map[string]map[string]string{},
			},
		},
		"null value": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfNull[fwtypes.MapValueOf[types.String]](ctx),
			},
			Target: &awsMapOfMapOfString{},
			WantTarget: &awsMapOfMapOfString{
				Field1: nil,
			},
		},
		"invalid conversion": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("y"),
					}),
				}),
			},
			Target:     &awsMapOfMapOfInt32{},
			WantTarget: &awsMapOfMapOfInt32{},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{})
}

func TestExpandMapOfMapOfStringPointer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			Target: &awsMapOfMapOfStringPointer{},
			WantTarget: &awsMapOfMapOfStringPointer{
				Field1: map[string]map[string]*string{
					"x": {
						"y": aws.String("z"),
					},
				},
			},
		},
		"multiple values": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
					"a": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"b": types.StringValue("c"),
					}),
				}),
			},
			Target: &awsMapOfMapOfStringPointer{},
			WantTarget: &awsMapOfMapOfStringPointer{
				Field1: map[string]map[string]*string{
					"x": {
						"y": aws.String("z"),
					},
					"a": {
						"b": aws.String("c"),
					},
				},
			},
		},
		"empty value": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{}),
			},
			Target: &awsMapOfMapOfStringPointer{},
			WantTarget: &awsMapOfMapOfStringPointer{
				Field1: map[string]map[string]*string{},
			},
		},
		"null value": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfNull[fwtypes.MapValueOf[types.String]](ctx),
			},
			Target: &awsMapOfMapOfStringPointer{},
			WantTarget: &awsMapOfMapOfStringPointer{
				Field1: nil,
			},
		},
		"invalid conversion": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("y"),
					}),
				}),
			},
			Target:     &awsMapOfMapOfInt32Pointer{},
			WantTarget: &awsMapOfMapOfInt32Pointer{},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{})
}

func TestExpandMapBlock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"nil map block key": {
			Source: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfNull[tfMapBlockElement](ctx),
			},
			Target: &awsMapBlockValues{},
			WantTarget: &awsMapBlockValues{
				MapBlock: nil,
			},
		},
		"map block key list": {
			Source: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
					{
						MapBlockKey: types.StringValue("y"),
						Attr1:       types.StringValue("c"),
						Attr2:       types.StringValue("d"),
					},
				}),
			},
			Target: &awsMapBlockValues{},
			WantTarget: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
					"y": {
						Attr1: "c",
						Attr2: "d",
					},
				},
			},
		},
		"map block key set": {
			Source: &tfMapBlockSet{
				MapBlock: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
					{
						MapBlockKey: types.StringValue("y"),
						Attr1:       types.StringValue("c"),
						Attr2:       types.StringValue("d"),
					},
				}),
			},
			Target: &awsMapBlockValues{},
			WantTarget: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
					"y": {
						Attr1: "c",
						Attr2: "d",
					},
				},
			},
		},
		"map block key ptr source": {
			Source: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
					{
						MapBlockKey: types.StringValue("y"),
						Attr1:       types.StringValue("c"),
						Attr2:       types.StringValue("d"),
					},
				}),
			},
			Target: &awsMapBlockValues{},
			WantTarget: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
					"y": {
						Attr1: "c",
						Attr2: "d",
					},
				},
			},
		},
		"map block key ptr both": {
			Source: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
					{
						MapBlockKey: types.StringValue("y"),
						Attr1:       types.StringValue("c"),
						Attr2:       types.StringValue("d"),
					},
				}),
			},
			Target: &awsMapBlockPointers{},
			WantTarget: &awsMapBlockPointers{
				MapBlock: map[string]*awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
					"y": {
						Attr1: "c",
						Attr2: "d",
					},
				},
			},
		},
		"map block enum key": {
			Source: &tfMapBlockListEnumKey{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfMapBlockElementEnumKey{
					{
						MapBlockKey: fwtypes.StringEnumValue(testEnumList),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
					{
						MapBlockKey: fwtypes.StringEnumValue(testEnumScalar),
						Attr1:       types.StringValue("c"),
						Attr2:       types.StringValue("d"),
					},
				}),
			},
			Target: &awsMapBlockValues{},
			WantTarget: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					string(testEnumList): {
						Attr1: "a",
						Attr2: "b",
					},
					string(testEnumScalar): {
						Attr1: "c",
						Attr2: "d",
					},
				},
			},
		},

		"map block list no key": {
			Source: &tfMapBlockListNoKey{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfMapBlockElementNoKey{
					{
						Attr1: types.StringValue("a"),
						Attr2: types.StringValue("b"),
					},
					{
						Attr1: types.StringValue("c"),
						Attr2: types.StringValue("d"),
					},
				}),
			},
			Target:        &awsMapBlockValues{},
			ExpectedDiags: diagAF[tfMapBlockElementNoKey](diagExpandingNoMapBlockKey),
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{})
}

func TestFlattenMaps(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"map of string": {
			Source: &awsMapOfString{
				FieldInner: map[string]string{
					"x": "y",
				},
			},
			Target: &tfMapOfString{},
			WantTarget: &tfMapOfString{
				FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
		},
		"map of string pointer": {
			Source: &awsMapOfStringPointer{
				FieldInner: map[string]*string{
					"x": aws.String("y"),
				},
			},
			Target: &tfMapOfString{},
			WantTarget: &tfMapOfString{
				FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
		},
		"nested string map": {
			Source: &awsNestedMapOfString{
				FieldOuter: awsMapOfString{
					FieldInner: map[string]string{
						"x": "y",
					},
				},
			},
			Target: &tfNestedMapOfString{},
			WantTarget: &tfNestedMapOfString{
				FieldOuter: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfMapOfString{
					FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"x": types.StringValue("y"),
					}),
				}),
			},
		},
		"map of map of string": {
			Source: &awsMapOfMapOfString{
				Field1: map[string]map[string]string{
					"x": {
						"y": "z",
					},
				},
			},
			Target: &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
		},
		"map of map of string pointer": {
			Source: &awsMapOfMapOfStringPointer{
				Field1: map[string]map[string]*string{
					"x": {
						"y": aws.String("z"),
					},
				},
			},
			Target: &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
		},
		"nil map of map of string": {
			Source:     &awsMapOfMapOfString{},
			Target:     &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{Field1: fwtypes.NewMapValueOfNull[fwtypes.MapValueOf[types.String]](ctx)},
		},
		"nil map of map of string pointer": {
			Source:     &awsMapOfMapOfStringPointer{},
			Target:     &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{Field1: fwtypes.NewMapValueOfNull[fwtypes.MapValueOf[types.String]](ctx)},
		},
		"incompatible map value type": {
			Source:     &awsMapOfInt{Field1: map[string]int{"a": 1}},
			Target:     &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{},
			WantDiff:   true,
		},
		// Previously caused a panic when structMapToObjectList attempted
		// to call flattenStruct on a dereferenced *string value.
		"map of string pointer to list nested object": {
			Source: &awsMapOfStringPointerField{
				Field1: map[string]*string{"key1": aws.String("val1")},
			},
			Target:        &tfMapStringPointerBlock{},
			WantTarget:    &tfMapStringPointerBlock{},
			ExpectedDiags: diag.Diagnostics{DiagFlatteningIncompatibleTypes(reflect.TypeFor[map[string]*string](), reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapStringPointerBlockElement]]())},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{})
}

func TestFlattenMapOfString(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &awsMapOf[string]{
				Field1: map[string]string{
					"x": "y",
				},
			},
			Target: &tfMapOf[types.String]{},
			WantTarget: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
		},
		"multiple values": {
			Source: &awsMapOf[string]{
				Field1: map[string]string{
					"x": "y",
					"y": "z",
				},
			},
			Target: &tfMapOf[types.String]{},
			WantTarget: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
					"y": types.StringValue("z"),
				}),
			},
		},
		"empty value": {
			Source: &awsMapOf[string]{
				Field1: map[string]string{},
			},
			Target: &tfMapOf[types.String]{},
			WantTarget: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfEmpty[types.String](ctx),
			},
		},
		"null value": {
			Source: &awsMapOf[string]{
				Field1: nil,
			},
			Target: &tfMapOf[types.String]{},
			WantTarget: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfNull[types.String](ctx),
			},
		},
		"invalid conversion": {
			Source: &awsMapOf[string]{
				Field1: map[string]string{
					"x": "y",
				},
			},
			Target:        &tfMapOf[types.Int32]{},
			ExpectedDiags: diagInvalidMapElementType(types.StringType, "x", types.Int32Type),
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{})
}

func TestFlattenMapOfStringPointer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &awsMapOf[*string]{
				Field1: map[string]*string{
					"x": aws.String("y"),
				},
			},
			Target: &tfMapOf[types.String]{},
			WantTarget: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
		},
		"multiple values": {
			Source: &awsMapOf[*string]{
				Field1: map[string]*string{
					"x": aws.String("y"),
					"y": aws.String("z"),
				},
			},
			Target: &tfMapOf[types.String]{},
			WantTarget: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
					"y": types.StringValue("z"),
				}),
			},
		},
		"empty value": {
			Source: &awsMapOf[*string]{
				Field1: map[string]*string{},
			},
			Target: &tfMapOf[types.String]{},
			WantTarget: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfEmpty[types.String](ctx),
			},
		},
		"null value": {
			Source: &awsMapOf[*string]{
				Field1: nil,
			},
			Target: &tfMapOf[types.String]{},
			WantTarget: &tfMapOf[types.String]{
				Field1: fwtypes.NewMapValueOfNull[types.String](ctx),
			},
		},
		"invalid conversion": {
			Source: &awsMapOf[*string]{
				Field1: map[string]*string{
					"x": aws.String("y"),
				},
			},
			Target:        &tfMapOf[types.Int32]{},
			ExpectedDiags: diagInvalidMapElementType(types.StringType, "x", types.Int32Type),
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{})
}

func TestFlattenMapOfInt32(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &awsMapOf[int32]{
				Field1: map[string]int32{
					"x": 42,
				},
			},
			Target: &tfMapOf[types.Int32]{},
			WantTarget: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfMust[types.Int32](ctx, map[string]attr.Value{
					"x": types.Int32Value(42),
				}),
			},
		},
		"multiple values": {
			Source: &awsMapOf[int32]{
				Field1: map[string]int32{
					"x": 42,
					"y": 43,
				},
			},
			Target: &tfMapOf[types.Int32]{},
			WantTarget: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfMust[types.Int32](ctx, map[string]attr.Value{
					"x": types.Int32Value(42),
					"y": types.Int32Value(43),
				}),
			},
		},
		"empty value": {
			Source: &awsMapOf[int32]{
				Field1: map[string]int32{},
			},
			Target: &tfMapOf[types.Int32]{},
			WantTarget: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfEmpty[types.Int32](ctx),
			},
		},
		"null value": {
			Source: &awsMapOf[int32]{
				Field1: nil,
			},
			Target: &tfMapOf[types.Int32]{},
			WantTarget: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfNull[types.Int32](ctx),
			},
		},
		"invalid conversion": {
			Source: &awsMapOf[int32]{
				Field1: map[string]int32{
					"x": 42,
				},
			},
			Target:        &tfMapOf[types.String]{},
			ExpectedDiags: diagInvalidMapElementType(types.Int32Type, "x", types.StringType),
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{})
}

func TestFlattenMapOfInt32Pointer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &awsMapOf[*int32]{
				Field1: map[string]*int32{
					"x": aws.Int32(42),
				},
			},
			Target: &tfMapOf[types.Int32]{},
			WantTarget: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfMust[types.Int32](ctx, map[string]attr.Value{
					"x": types.Int32Value(42),
				}),
			},
		},
		"multiple values": {
			Source: &awsMapOf[*int32]{
				Field1: map[string]*int32{
					"x": aws.Int32(42),
					"y": aws.Int32(43),
				},
			},
			Target: &tfMapOf[types.Int32]{},
			WantTarget: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfMust[types.Int32](ctx, map[string]attr.Value{
					"x": types.Int32Value(42),
					"y": types.Int32Value(43),
				}),
			},
		},
		"empty value": {
			Source: &awsMapOf[*int32]{
				Field1: map[string]*int32{},
			},
			Target: &tfMapOf[types.Int32]{},
			WantTarget: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfEmpty[types.Int32](ctx),
			},
		},
		"null value": {
			Source: &awsMapOf[*int32]{
				Field1: nil,
			},
			Target: &tfMapOf[types.Int32]{},
			WantTarget: &tfMapOf[types.Int32]{
				Field1: fwtypes.NewMapValueOfNull[types.Int32](ctx),
			},
		},
		"invalid conversion": {
			Source: &awsMapOf[*int32]{
				Field1: map[string]*int32{
					"x": aws.Int32(42),
				},
			},
			Target:        &tfMapOf[types.String]{},
			ExpectedDiags: diagInvalidMapElementType(types.Int32Type, "x", types.StringType),
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{})
}

func TestFlattenMapOfMapOfString(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &awsMapOfMapOfString{
				Field1: map[string]map[string]string{
					"x": {
						"y": "z",
					},
				},
			},
			Target: &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
		},
		"empty value": {
			Source: &awsMapOfMapOfString{
				Field1: map[string]map[string]string{},
			},
			Target: &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfEmpty[fwtypes.MapValueOf[types.String]](ctx),
			},
		},
		"null value": {
			Source: &awsMapOfMapOfString{
				Field1: nil,
			},
			Target: &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfNull[fwtypes.MapValueOf[types.String]](ctx),
			},
		},
		"invalid conversion": {
			Source: &awsMapOfMapOfString{
				Field1: map[string]map[string]string{
					"x": {
						"y": "z",
					},
				},
			},
			Target:        &tfMapOfMapOfInt32{},
			ExpectedDiags: diagInvalidMapElementType(fwtypes.NewMapTypeOf[types.String](ctx), "x", fwtypes.NewMapTypeOf[types.Int32](ctx)),
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{})
}

func TestFlattenMapOfMapOfStringPointer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single value": {
			Source: &awsMapOfMapOfStringPointer{
				Field1: map[string]map[string]*string{
					"x": {
						"y": aws.String("z"),
					},
				},
			},
			Target: &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
		},
		"empty value": {
			Source: &awsMapOfMapOfStringPointer{
				Field1: map[string]map[string]*string{},
			},
			Target: &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfEmpty[fwtypes.MapValueOf[types.String]](ctx),
			},
		},
		"null value": {
			Source: &awsMapOfMapOfStringPointer{
				Field1: nil,
			},
			Target: &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfNull[fwtypes.MapValueOf[types.String]](ctx),
			},
		},
		"invalid conversion": {
			Source: &awsMapOfMapOfStringPointer{
				Field1: map[string]map[string]*string{
					"x": {
						"y": aws.String("z"),
					},
				},
			},
			Target:        &tfMapOfMapOfInt32{},
			ExpectedDiags: diagInvalidMapElementType(fwtypes.NewMapTypeOf[types.String](ctx), "x", fwtypes.NewMapTypeOf[types.Int32](ctx)),
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{})
}

func TestFlattenMapBlock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"nil map block key": {
			Source: &awsMapBlockValues{
				MapBlock: nil,
			},
			Target: &tfMapBlockList{},
			WantTarget: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfNull[tfMapBlockElement](ctx),
			},
		},
		"map block key list": {
			Source: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &tfMapBlockList{},
			WantTarget: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
		},
		"map block key set": {
			Source: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &tfMapBlockSet{},
			WantTarget: &tfMapBlockSet{
				MapBlock: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
		},
		"nil map block key ptr": {
			Source: &awsMapBlockPointers{
				MapBlock: nil,
			},
			Target: &tfMapBlockList{},
			WantTarget: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfNull[tfMapBlockElement](ctx),
			},
		},
		"map block key ptr source": {
			Source: &awsMapBlockPointers{
				MapBlock: map[string]*awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &tfMapBlockList{},
			WantTarget: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
		},
		"map block key ptr both": {
			Source: &awsMapBlockPointers{
				MapBlock: map[string]*awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &tfMapBlockList{},
			WantTarget: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
		},
		"map block enum key": {
			Source: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					string(testEnumList): {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &tfMapBlockListEnumKey{},
			WantTarget: &tfMapBlockListEnumKey{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfMapBlockElementEnumKey{
					{
						MapBlockKey: fwtypes.StringEnumValue(testEnumList),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
		},

		"map block list no key": {
			Source: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target:        &tfMapBlockListNoKey{},
			ExpectedDiags: diagAF[tfMapBlockElementNoKey](diagFlatteningNoMapBlockKey),
		},
	}
	runAutoFlattenTestCases(t, testCases, runChecks{})
}

// This error message is copied from `basetypes.NewMapValue`
// https://github.com/hashicorp/terraform-plugin-framework/blob/00badb1b0584b22f3799c474b9f1cea4b49e5163/types/basetypes/map_value.go#L70
// If the error message changes, this will need to be updated.
func diagInvalidMapElementType(sourceElementType attr.Type, key string, targetElementType attr.Type) diag.Diagnostics {
	return diag.Diagnostics{
		diag.NewErrorDiagnostic(
			"Invalid Map Element Type",
			"While creating a Map value, an invalid element was detected. "+
				"A Map must use the single, given element type. "+
				"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
				fmt.Sprintf("Map Element Type: %s\n", targetElementType.String())+
				fmt.Sprintf("Map Key (%s) Element Type: %s", key, sourceElementType.String()),
		),
	}
}
