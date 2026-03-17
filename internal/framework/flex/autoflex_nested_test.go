// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

// Tests AutoFlex's Expand/Flatten of nested blocks and complex types.

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type tfSingleStringField struct {
	Field1 types.String `tfsdk:"field1"`
}

type tfListOfNestedObjectLegacy struct {
	Field1 fwtypes.ListNestedObjectValueOf[tfSingleStringField] `tfsdk:"field1" autoflex:",legacy"`
}

type tfListOfNestedObject struct {
	Field1 fwtypes.ListNestedObjectValueOf[tfSingleStringField] `tfsdk:"field1"`
}

type tfSetOfNestedObject struct {
	Field1 fwtypes.SetNestedObjectValueOf[tfSingleStringField] `tfsdk:"field1"`
}

type tfSetOfNestedObjectLegacy struct {
	Field1 fwtypes.SetNestedObjectValueOf[tfSingleStringField] `tfsdk:"field1" autoflex:",legacy"`
}

type awsNestedObjectPointer struct {
	Field1 *awsSingleStringValue
}

type awsSliceOfNestedObjectPointers struct {
	Field1 []*awsSingleStringValue
}

type awsSliceOfNestedObjectValues struct {
	Field1 []awsSingleStringValue
}

func TestExpandSimpleSingleNestedBlock(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.String `tfsdk:"field1"`
		Field2 types.Int64  `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 *string
		Field2 int64
	}

	type tf02 struct {
		Field1 fwtypes.ObjectValueOf[tf01] `tfsdk:"field1"`
	}
	type aws02 struct {
		Field1 *aws01
	}
	type aws03 struct {
		Field1 aws01
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"single nested block pointer": {
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfMust(ctx, &tf01{Field1: types.StringValue("a"), Field2: types.Int64Value(1)})},
			Target:     &aws02{},
			WantTarget: &aws02{Field1: &aws01{Field1: aws.String("a"), Field2: 1}},
		},
		"single nested block nil": {
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfNull[tf01](ctx)},
			Target:     &aws02{},
			WantTarget: &aws02{},
		},
		"single nested block value": {
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfMust(ctx, &tf01{Field1: types.StringValue("a"), Field2: types.Int64Value(1)})},
			Target:     &aws03{},
			WantTarget: &aws03{Field1: aws01{Field1: aws.String("a"), Field2: 1}},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestExpandNestedComplex(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"complex Source and complex Target": {
			Source: &tfComplexValue{
				Field1: types.StringValue("m"),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("n"),
					}),
				}),
				Field3: types.MapValueMust(types.StringType, map[string]attr.Value{
					"X": types.StringValue("x"),
					"Y": types.StringValue("y"),
				}),
				Field4: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleInt64Field{
					{Field1: types.Int64Value(100)},
					{Field1: types.Int64Value(2000)},
					{Field1: types.Int64Value(30000)},
				}),
			},
			Target: &awsComplexValue{},
			WantTarget: &awsComplexValue{
				Field1: "m",
				Field2: &awsNestedObjectPointer{
					Field1: &awsSingleStringValue{
						Field1: "n",
					},
				},
				Field3: aws.StringMap(map[string]string{
					"X": "x",
					"Y": "y",
				}),
				Field4: []awsSingleInt64Value{
					{Field1: 100},
					{Field1: 2000},
					{Field1: 30000},
				},
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestExpandComplexSingleNestedBlock(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Bool                        `tfsdk:"field1"`
		Field2 fwtypes.ListValueOf[types.String] `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 bool
		Field2 []string
	}

	type tf02 struct {
		Field1 fwtypes.ObjectValueOf[tf01] `tfsdk:"field1"`
	}
	type aws02 struct {
		Field1 *aws01
	}

	type tf03 struct {
		Field1 fwtypes.ObjectValueOf[tf02] `tfsdk:"field1"`
	}
	type aws03 struct {
		Field1 *aws02
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"single nested block pointer": {
			Source: &tf03{
				Field1: fwtypes.NewObjectValueOfMust(
					ctx,
					&tf02{
						Field1: fwtypes.NewObjectValueOfMust(
							ctx,
							&tf01{
								Field1: types.BoolValue(true),
								Field2: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("a"), types.StringValue("b")}),
							},
						),
					},
				),
			},
			Target: &aws03{},
			WantTarget: &aws03{
				Field1: &aws02{
					Field1: &aws01{
						Field1: true,
						Field2: []string{"a", "b"},
					},
				},
			},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestExpandTopLevelListOfNestedObject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid value to []struct": {
			Source: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			Target: &[]awsSingleStringValue{},
			WantTarget: &[]awsSingleStringValue{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
		},
		"empty value to []struct": {
			Source:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			Target:     &[]awsSingleStringValue{},
			WantTarget: &[]awsSingleStringValue{},
		},
		"null value to []struct": {
			Source:     fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
			Target:     &[]awsSingleStringValue{},
			WantTarget: &[]awsSingleStringValue{},
		},

		"valid value to []*struct": {
			Source: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			Target: &[]*awsSingleStringValue{},
			WantTarget: &[]*awsSingleStringValue{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
		},
		"empty value to []*struct": {
			Source:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			Target:     &[]*awsSingleStringValue{},
			WantTarget: &[]*awsSingleStringValue{},
		},
		"null value to []*struct": {
			Source:     fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
			Target:     &[]*awsSingleStringValue{},
			WantTarget: &[]*awsSingleStringValue{},
		},

		"single list value to single struct": {
			Source: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
			}),
			Target: &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{
				Field1: "value1",
			},
		},
		"empty list value to single struct": {
			Source:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
		},
		"null value to single struct": {
			Source:     fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestExpandSetOfNestedObject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid value to []struct": {
			Source: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			Target: &[]awsSingleStringValue{},
			WantTarget: &[]awsSingleStringValue{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
		},
		"empty value to []struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			Target:     &[]awsSingleStringValue{},
			WantTarget: &[]awsSingleStringValue{},
		},
		"null value to []struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfNull[tfSingleStringField](ctx),
			Target:     &[]awsSingleStringValue{},
			WantTarget: &[]awsSingleStringValue{},
		},

		"valid value to []*struct": {
			Source: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			Target: &[]*awsSingleStringValue{},
			WantTarget: &[]*awsSingleStringValue{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
		},
		"empty value to []*struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			Target:     &[]*awsSingleStringValue{},
			WantTarget: &[]*awsSingleStringValue{},
		},
		"null value to []*struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfNull[tfSingleStringField](ctx),
			Target:     &[]*awsSingleStringValue{},
			WantTarget: &[]*awsSingleStringValue{},
		},

		"single set value to single struct": {
			Source: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
			}),
			Target: &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{
				Field1: "value1",
			},
		},
		"empty set value to single struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
		},
		"null value to single struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfNull[tfSingleStringField](ctx),
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestExpandSimpleNestedBlockWithStringEnum(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Int64                  `tfsdk:"field1"`
		Field2 fwtypes.StringEnum[testEnum] `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 int64
		Field2 testEnum
	}

	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &tf01{
				Field1: types.Int64Value(1),
				Field2: fwtypes.StringEnumValue(testEnumList),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: 1,
				Field2: testEnumList,
			},
		},
		"single nested null value": {
			Source: &tf01{
				Field1: types.Int64Value(1),
				Field2: fwtypes.StringEnumNull[testEnum](),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: 1,
				Field2: "",
			},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestExpandComplexNestedBlockWithStringEnum(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field2 fwtypes.StringEnum[testEnum] `tfsdk:"field2"`
	}
	type tf02 struct {
		Field1 types.Int64                           `tfsdk:"field1"`
		Field2 fwtypes.ListNestedObjectValueOf[tf01] `tfsdk:"field2"`
	}
	type aws02 struct {
		Field2 testEnum
	}
	type aws01 struct {
		Field1 int64
		Field2 *aws02
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: fwtypes.StringEnumValue(testEnumList),
				}),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field2: testEnumList,
				},
			},
		},
		"single nested null value": {
			Source: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: fwtypes.StringEnumNull[testEnum](),
				}),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field2: "",
				},
			},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestExpandListOfNestedObjectField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"ListNestedObject to *struct": {
			"value": {
				Source: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("a"),
					}),
				},
				Target: &awsNestedObjectPointer{},
				WantTarget: &awsNestedObjectPointer{
					Field1: &awsSingleStringValue{
						Field1: "a",
					},
				},
			},
		},

		"ListNestedObject to []struct": {
			"empty": {
				Source: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
				Target: &awsSliceOfNestedObjectValues{},
				WantTarget: &awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{},
				},
			},
			"values": {
				Source: &tfListOfNestedObject{Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
					{Field1: types.StringValue("a")},
					{Field1: types.StringValue("b")},
				})},
				Target: &awsSliceOfNestedObjectValues{},
				WantTarget: &awsSliceOfNestedObjectValues{Field1: []awsSingleStringValue{
					{Field1: "a"},
					{Field1: "b"},
				}},
			},
		},

		"ListNestedObject to []*struct": {
			"empty": {
				Source: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
				Target: &awsSliceOfNestedObjectPointers{},
				WantTarget: &awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{},
				},
			},
			"values": {
				Source: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
				Target: &awsSliceOfNestedObjectPointers{},
				WantTarget: &awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

func TestExpandSetOfNestedObjectField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"SetNestedObject to *struct": {
			"value": {
				Source: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("a"),
					}),
				},
				Target: &awsNestedObjectPointer{},
				WantTarget: &awsNestedObjectPointer{
					Field1: &awsSingleStringValue{
						Field1: "a",
					},
				},
			},
		},

		"SetNestedObject to []*struct": {
			"empty": {
				Source:     &tfSetOfNestedObject{Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{})},
				Target:     &awsSliceOfNestedObjectPointers{},
				WantTarget: &awsSliceOfNestedObjectPointers{Field1: []*awsSingleStringValue{}},
			},
			"values": {
				Source: &tfSetOfNestedObject{Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{
					{Field1: types.StringValue("a")},
					{Field1: types.StringValue("b")},
				})},
				Target: &awsSliceOfNestedObjectPointers{},
				WantTarget: &awsSliceOfNestedObjectPointers{Field1: []*awsSingleStringValue{
					{Field1: "a"},
					{Field1: "b"},
				}},
			},
		},

		"SetNestedObject to []struct": {
			"values": {
				Source: &tfSetOfNestedObject{Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
					{Field1: types.StringValue("a")},
					{Field1: types.StringValue("b")},
				})},
				Target: &awsSliceOfNestedObjectValues{},
				WantTarget: &awsSliceOfNestedObjectValues{Field1: []awsSingleStringValue{
					{Field1: "a"},
					{Field1: "b"},
				}},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

func TestFlattenNestedComplex(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"complex Source and complex Target": {
			Source: &awsComplexValue{
				Field1: "m",
				Field2: &awsNestedObjectPointer{Field1: &awsSingleStringValue{Field1: "n"}},
				Field3: aws.StringMap(map[string]string{"X": "x", "Y": "y"}),
				Field4: []awsSingleInt64Value{{Field1: 100}, {Field1: 2000}, {Field1: 30000}},
			},
			Target: &tfComplexValue{},
			WantTarget: &tfComplexValue{
				Field1: types.StringValue("m"),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("n"),
					}),
				}),
				Field3: types.MapValueMust(types.StringType, map[string]attr.Value{
					"X": types.StringValue("x"),
					"Y": types.StringValue("y"),
				}),
				Field4: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleInt64Field{
					{Field1: types.Int64Value(100)},
					{Field1: types.Int64Value(2000)},
					{Field1: types.Int64Value(30000)},
				}),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestFlattenSimpleNestedBlockWithStringEnum(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Int64                  `tfsdk:"field1"`
		Field2 fwtypes.StringEnum[testEnum] `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 int64
		Field2 testEnum
	}

	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &aws01{
				Field1: 1,
				Field2: testEnumList,
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.Int64Value(1),
				Field2: fwtypes.StringEnumValue(testEnumList),
			},
		},
		"single nested empty value": {
			Source: &aws01{
				Field1: 1,
				Field2: "",
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.Int64Value(1),
				Field2: fwtypes.StringEnumNull[testEnum](),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenComplexNestedBlockWithStringEnum(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field2 fwtypes.StringEnum[testEnum] `tfsdk:"field2"`
	}
	type tf02 struct {
		Field1 types.Int64                           `tfsdk:"field1"`
		Field2 fwtypes.ListNestedObjectValueOf[tf01] `tfsdk:"field2"`
	}
	type aws02 struct {
		Field2 testEnum
	}
	type aws01 struct {
		Field1 int64
		Field2 *aws02
	}

	ctx := context.Background()
	var zero fwtypes.StringEnum[testEnum]
	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field2: testEnumList,
				},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: fwtypes.StringEnumValue(testEnumList),
				}),
			},
		},
		"single nested empty value": {
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{Field2: ""},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: fwtypes.StringEnumNull[testEnum](),
				}),
			},
		},
		"single nested zero value": {
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field2: ""},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: zero,
				}),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenSimpleSingleNestedBlock(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.String `tfsdk:"field1"`
		Field2 types.Int64  `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 *string
		Field2 int64
	}

	type tf02 struct {
		Field1 fwtypes.ObjectValueOf[tf01] `tfsdk:"field1"`
	}
	type aws02 struct {
		Field1 *aws01
	}
	type aws03 struct {
		Field1 aws01
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"single nested block pointer": {
			Source: &aws02{
				Field1: &aws01{
					Field1: aws.String("a"),
					Field2: 1,
				},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tf01{
					Field1: types.StringValue("a"),
					Field2: types.Int64Value(1),
				}),
			},
		},
		"single nested block nil": {
			Source: &aws02{},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: fwtypes.NewObjectValueOfNull[tf01](ctx),
			},
		},
		"single nested block value": {
			Source: &aws03{
				Field1: aws01{
					Field1: aws.String("a"),
					Field2: 1},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tf01{
					Field1: types.StringValue("a"),
					Field2: types.Int64Value(1),
				}),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenComplexSingleNestedBlock(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Bool                        `tfsdk:"field1"`
		Field2 fwtypes.ListValueOf[types.String] `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 bool
		Field2 []string
	}

	type tf02 struct {
		Field1 fwtypes.ObjectValueOf[tf01] `tfsdk:"field1"`
	}
	type aws02 struct {
		Field1 *aws01
	}

	type tf03 struct {
		Field1 fwtypes.ObjectValueOf[tf02] `tfsdk:"field1"`
	}
	type aws03 struct {
		Field1 *aws02
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"single nested block pointer": {
			Source: &aws03{
				Field1: &aws02{
					Field1: &aws01{
						Field1: true,
						Field2: []string{"a", "b"},
					},
				},
			},
			Target: &tf03{},
			WantTarget: &tf03{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tf02{
					Field1: fwtypes.NewObjectValueOfMust(ctx, &tf01{
						Field1: types.BoolValue(true),
						Field2: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
							types.StringValue("a"),
							types.StringValue("b"),
						}),
					}),
				}),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenSimpleNestedBlockWithFloat32(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Int64   `tfsdk:"field1"`
		Field2 types.Float64 `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 int64
		Field2 *float32
	}

	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source:     &aws01{Field1: 1, Field2: aws.Float32(0.01)},
			Target:     &tf01{},
			WantTarget: &tf01{Field1: types.Int64Value(1), Field2: types.Float64Value(0.01)},
		},
	}
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenComplexNestedBlockWithFloat32(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Float64 `tfsdk:"field1"`
		Field2 types.Float64 `tfsdk:"field2"`
	}
	type tf02 struct {
		Field1 types.Int64                           `tfsdk:"field1"`
		Field2 fwtypes.ListNestedObjectValueOf[tf01] `tfsdk:"field2"`
	}
	type aws02 struct {
		Field1 float32
		Field2 *float32
	}
	type aws01 struct {
		Field1 int64
		Field2 *aws02
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field1: 1.11,
					Field2: aws.Float32(-2.22),
				},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field1: types.Float64Value(1.11),
					Field2: types.Float64Value(-2.22),
				}),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenSimpleNestedBlockWithFloat64(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Int64   `tfsdk:"field1"`
		Field2 types.Float64 `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 int64
		Field2 *float64
	}

	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &aws01{
				Field1: 1,
				Field2: aws.Float64(0.01),
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.Int64Value(1),
				Field2: types.Float64Value(0.01),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenComplexNestedBlockWithFloat64(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Float64 `tfsdk:"field1"`
		Field2 types.Float64 `tfsdk:"field2"`
	}
	type tf02 struct {
		Field1 types.Int64                           `tfsdk:"field1"`
		Field2 fwtypes.ListNestedObjectValueOf[tf01] `tfsdk:"field2"`
	}
	type aws02 struct {
		Field1 float64
		Field2 *float64
	}
	type aws01 struct {
		Field1 int64
		Field2 *aws02
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field1: 1.11,
					Field2: aws.Float64(-2.22),
				},
			},
			Target:     &tf02{},
			WantTarget: &tf02{Field1: types.Int64Value(1), Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{Field1: types.Float64Value(1.11), Field2: types.Float64Value(-2.22)})},
		},
	}
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenObjectValueField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"*struct to ObjectValue": {
			"nil": {
				Source: awsNestedObjectPointer{},
				Target: &tfObjectValue[tfSingleStringField]{},
				WantTarget: &tfObjectValue[tfSingleStringField]{
					Field1: fwtypes.NewObjectValueOfNull[tfSingleStringField](ctx),
				},
			},
			"value": {
				Source: awsNestedObjectPointer{
					Field1: &awsSingleStringValue{
						Field1: "a",
					},
				},
				Target: &tfObjectValue[tfSingleStringField]{},
				WantTarget: &tfObjectValue[tfSingleStringField]{
					Field1: fwtypes.NewObjectValueOfMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("a"),
					}),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

func TestFlattenListOfNestedObjectField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"*struct to ListNestedObject": {
			"nil": {
				Source: awsNestedObjectPointer{},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
				},
			},
			"value": {
				Source: awsNestedObjectPointer{
					Field1: &awsSingleStringValue{
						Field1: "a",
					},
				},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("a"),
					}),
				},
			},
		},

		"legacy *struct to ListNestedObject": {
			"nil": {
				Source: awsNestedObjectPointer{},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{}),
				},
			},
			"value": {
				Source: awsNestedObjectPointer{
					Field1: &awsSingleStringValue{
						Field1: "a",
					},
				},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("a"),
					}),
				},
			},
		},

		"[]struct to ListNestedObject": {
			"nil": {
				Source: awsSliceOfNestedObjectValues{},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
				},
			},
			"empty": {
				Source: awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{},
				},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
			},
			"values": {
				Source: awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
			},
		},

		"legacy []struct to ListNestedObject": {
			"nil": {
				Source: awsSliceOfNestedObjectValues{},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
			},
			"empty": {
				Source: awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{},
				},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
			},
			"values": {
				Source: awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
			},
		},

		"[]*struct to ListNestedObject": {
			"nil": {
				Source: awsSliceOfNestedObjectPointers{},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
				},
			},
			"empty": {
				Source: awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{},
				},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
			},
			"values": {
				Source: awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
			},
		},

		"legacy []*struct to ListNestedObject": {
			"nil": {
				Source: awsSliceOfNestedObjectPointers{},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
			},
			"empty": {
				Source: awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{},
				},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
			},
			"values": {
				Source: awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

func TestFlattenTopLevelListOfNestedObject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]toplevelTestCase[[]awsSingleStringValue, fwtypes.ListNestedObjectValueOf[tfSingleStringField]]{
		"values": {
			source: []awsSingleStringValue{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
			expectedValue: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
		},

		"empty": {
			source:        []awsSingleStringValue{},
			expectedValue: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
		},

		"null": {
			source:        nil,
			expectedValue: fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
		},
	}

	runTopLevelTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenSetOfNestedObjectField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"*struct to SetNestedObject": {
			"nil": {
				Source: awsNestedObjectPointer{},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfNull[tfSingleStringField](ctx),
				},
			},
			"value": {
				Source: awsNestedObjectPointer{
					Field1: &awsSingleStringValue{Field1: "a"},
				},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("a"),
					}),
				},
			},
		},

		"[]struct to SetNestedObject": {
			"nil": {
				Source: &awsSliceOfNestedObjectValues{},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfNull[tfSingleStringField](ctx),
				},
			},
			"empty": {
				Source: &awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{},
				},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
			},
			"values": {
				Source: &awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
			},
		},

		"legacy []struct to SetNestedObject": {
			"nil": {
				Source: &awsSliceOfNestedObjectValues{},
				Target: &tfSetOfNestedObjectLegacy{},
				WantTarget: &tfSetOfNestedObjectLegacy{
					Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
			},
			"empty": {
				Source: &awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{},
				},
				Target: &tfSetOfNestedObjectLegacy{},
				WantTarget: &tfSetOfNestedObjectLegacy{
					Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
			},
			"values": {
				Source: &awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfSetOfNestedObjectLegacy{},
				WantTarget: &tfSetOfNestedObjectLegacy{
					Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
			},
		},

		"[]*struct to SetNestedObject": {
			"nil": {
				Source: &awsSliceOfNestedObjectPointers{},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfNull[tfSingleStringField](ctx),
				},
			},
			"empty": {
				Source: &awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{},
				},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
			},
			"values": {
				Source: &awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
			},
		},

		"legacy []*struct to SetNestedObject": {
			"nil": {
				Source: &awsSliceOfNestedObjectPointers{},
				Target: &tfSetOfNestedObjectLegacy{},
				WantTarget: &tfSetOfNestedObjectLegacy{
					Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
			},
			"empty": {
				Source: &awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{},
				},
				Target: &tfSetOfNestedObjectLegacy{},
				WantTarget: &tfSetOfNestedObjectLegacy{
					Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
			},
			"values": {
				Source: &awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfSetOfNestedObjectLegacy{},
				WantTarget: &tfSetOfNestedObjectLegacy{
					Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

type rootStringModel struct {
	Field1 types.String `tfsdk:"field1"`
}

type rootListNestedObjectModel struct {
	Field1 fwtypes.ListNestedObjectValueOf[nestedModel] `tfsdk:"field1"`
}

type rootSetNestedObjectModel struct {
	Field1 fwtypes.SetNestedObjectValueOf[nestedModel] `tfsdk:"field1"`
}

type nestedModel struct {
	Field1 types.String `tfsdk:"field1"`
}

func TestFlattenPrePopulate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testCases := map[string]struct {
		target   any
		expected any
	}{
		"string": {
			target: &rootStringModel{},
			expected: &rootStringModel{
				Field1: types.StringNull(),
			},
		},

		"nested list": {
			target: &rootListNestedObjectModel{},
			expected: &rootListNestedObjectModel{
				Field1: fwtypes.NewListNestedObjectValueOfNull[nestedModel](ctx),
			},
		},

		"nested set": {
			target: &rootSetNestedObjectModel{},
			expected: &rootSetNestedObjectModel{
				Field1: fwtypes.NewSetNestedObjectValueOfNull[nestedModel](ctx),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			valTo := reflect.ValueOf(testCase.target)

			diags := flattenPrePopulate(ctx, valTo)

			if l := len(diags); l > 0 {
				t.Fatalf("expected 0 diags, got %s", fwdiag.DiagnosticsString(diags))
			}

			if diff := cmp.Diff(testCase.target, testCase.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
