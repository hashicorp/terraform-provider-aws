// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

// Tests AutoFlex's Expand/Flatten of maps.

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

type tfNestedMapOfString struct {
	FieldOuter fwtypes.ListNestedObjectValueOf[tfMapOfString] `tfsdk:"field_outer"`
}

type awsNestedMapOfString struct {
	FieldOuter awsMapOfString
}

type tfMapOfMapOfString struct {
	Field1 fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]] `tfsdk:"field1"`
}

type awsMapOfString struct {
	FieldInner map[string]string
}

type awsMapOfMapOfString struct {
	Field1 map[string]map[string]string
}

type awsMapOfMapOfStringPointer struct {
	Field1 map[string]map[string]*string
}

type awsMapOfStringPointer struct {
	FieldInner map[string]*string
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
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
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
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
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
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
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
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}
