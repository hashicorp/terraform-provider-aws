// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

// Tests AutoFlex's Expand/Flatten for list, set, and map conversions—verifying value correctness and
// diagnostics, not internal logging or trace output. For logging validation, see autoflex_dispatch_test.go.
// Specific map tests are in autoflex_maps_test.go.

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// List/Set/Map of primitive types.
type tfCollectionsOfPrimitiveElements struct {
	Field1 types.List `tfsdk:"field1"`
	Field2 types.List `tfsdk:"field2"`
	Field3 types.Set  `tfsdk:"field3"`
	Field4 types.Set  `tfsdk:"field4"`
	Field5 types.Map  `tfsdk:"field5"`
	Field6 types.Map  `tfsdk:"field6"`
}

type awsCollectionsOfPrimitiveElements struct {
	Field1 []string
	Field2 []*string
	Field3 []string
	Field4 []*string
	Field5 map[string]string
	Field6 map[string]*string
}

func TestExpandCollections(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"Collection of primitive types Source and slice or map of primtive types Target": {
			Source: &tfCollectionsOfPrimitiveElements{
				Field1: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field2: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field3: types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field4: types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field5: types.MapValueMust(types.StringType, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
				Field6: types.MapValueMust(types.StringType, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
			},
			Target: &awsCollectionsOfPrimitiveElements{},
			WantTarget: &awsCollectionsOfPrimitiveElements{
				Field1: []string{"a", "b"},
				Field2: aws.StringSlice([]string{"a", "b"}),
				Field3: []string{"a", "b"},
				Field4: aws.StringSlice([]string{"a", "b"}),
				Field5: map[string]string{"A": "a", "B": "b"},
				Field6: aws.StringMap(map[string]string{"A": "a", "B": "b"}),
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
}

func TestExpandListOfInt64(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"valid value []int64": {
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int64{},
			WantTarget: &[]int64{1, -1},
		},
		"empty value []int64": {
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
		},
		"null value []int64": {
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
		},
		"valid value []*int64": {
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{aws.Int64(1), aws.Int64(-1)},
		},
		"empty value []*int64": {
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
		},
		"null value []*int64": {
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
		},
		"valid value []int32": {
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int32{},
			WantTarget: &[]int32{1, -1},
		},
		"empty value []int32": {
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
		},
		"null value []int32": {
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
		},
		"valid value []*int32": {
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{aws.Int32(1), aws.Int32(-1)},
		},
		"empty value []*int32": {
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
		},
		"null value []*int32": {
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
}

func TestExpandSetOfInt64(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"valid value []int64": {
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int64{},
			WantTarget: &[]int64{1, -1},
		},
		"empty value []int64": {
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
		},
		"null value []int64": {
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
		},
		"valid value []*int64": {
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{aws.Int64(1), aws.Int64(-1)},
		},
		"empty value []*int64": {
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
		},
		"null value []*int64": {
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
		},
		"valid value []int32": {
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int32{},
			WantTarget: &[]int32{1, -1},
		},
		"empty value []int32": {
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
		},
		"null value []int32": {
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
		},
		"valid value []*int32": {
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{aws.Int32(1), aws.Int32(-1)},
		},
		"empty value []*int32": {
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
		},
		"null value []*int32": {
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
}

func TestExpandListOfStringEnum(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"valid value": {
			Source: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue(string(testEnumScalar)),
				types.StringValue(string(testEnumList)),
			}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{testEnumScalar, testEnumList},
		},
		"empty value": {
			Source:     types.ListValueMust(types.StringType, []attr.Value{}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
		},
		"null value": {
			Source:     types.ListNull(types.StringType),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
}

func TestExpandSetOfStringEnum(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"valid value": {
			Source: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue(string(testEnumScalar)),
				types.StringValue(string(testEnumList)),
			}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{testEnumScalar, testEnumList},
		},
		"empty value": {
			Source:     types.SetValueMust(types.StringType, []attr.Value{}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
		},
		"null value": {
			Source:     types.SetNull(types.StringType),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
}

type tfListOfStringEnum struct {
	Field1 fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]] `tfsdk:"field1"`
}

type awsSliceOfStringEnum struct {
	Field1 []testEnum
}

func TestExpandStructListOfStringEnum(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid value": {
			Source: &tfListOfStringEnum{
				Field1: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
			Target: &awsSliceOfStringEnum{},
			WantTarget: &awsSliceOfStringEnum{
				Field1: []testEnum{testEnumScalar, testEnumList},
			},
		},
		"empty value": {
			Source: &tfListOfStringEnum{
				Field1: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
			},
			Target: &awsSliceOfStringEnum{},
			WantTarget: &awsSliceOfStringEnum{
				Field1: []testEnum{},
			},
		},
		"null value": {
			Source: &tfListOfStringEnum{
				Field1: fwtypes.NewListValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
			},
			Target:     &awsSliceOfStringEnum{},
			WantTarget: &awsSliceOfStringEnum{},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
}

type tfSetOfStringEnum struct {
	Field1 fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]] `tfsdk:"field1"`
}

func TestExpandStructSetOfStringEnum(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid value": {
			Source: &tfSetOfStringEnum{
				Field1: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
			Target: &awsSliceOfStringEnum{},
			WantTarget: &awsSliceOfStringEnum{
				Field1: []testEnum{testEnumScalar, testEnumList},
			},
		},
		"empty value": {
			Source: &tfSetOfStringEnum{
				Field1: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
			},
			Target: &awsSliceOfStringEnum{},
			WantTarget: &awsSliceOfStringEnum{
				Field1: []testEnum{},
			},
		},
		"null value": {
			Source: &tfSetOfStringEnum{
				Field1: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
			},
			Target:     &awsSliceOfStringEnum{},
			WantTarget: &awsSliceOfStringEnum{},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
}

type tfSingleStringFieldIgnore struct {
	Field1 types.String `tfsdk:"field1" autoflex:"-"`
}

func TestExpandIgnoreStructTag(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"to value": {
			Source: tfSingleStringFieldIgnore{
				Field1: types.StringValue("value1"),
			},
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
		},
		"to pointer": {
			Source: tfSingleStringFieldIgnore{
				Field1: types.StringValue("value1"),
			},
			Target:     &awsSingleStringPointer{},
			WantTarget: &awsSingleStringPointer{},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
}

type TFExportedStruct struct {
	Field1 types.String `tfsdk:"field1"`
}

type tfExportedEmbeddedStruct struct {
	TFExportedStruct
	Field2 types.String `tfsdk:"field2"`
}

type tfUnexportedEmbeddedStruct struct {
	tfSingleStringField
	Field2 types.String `tfsdk:"field2"`
}

type awsEmbeddedStruct struct {
	Field1 string
	Field2 string
}

func TestExpandEmbeddedStruct(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"exported": {
			Source: &tfExportedEmbeddedStruct{
				TFExportedStruct: TFExportedStruct{
					Field1: types.StringValue("a"),
				},
				Field2: types.StringValue("b"),
			},
			Target: &awsEmbeddedStruct{},
			WantTarget: &awsEmbeddedStruct{
				Field1: "a",
				Field2: "b",
			},
		},
		"unexported": {
			Source: &tfUnexportedEmbeddedStruct{
				tfSingleStringField: tfSingleStringField{
					Field1: types.StringValue("a"),
				},
				Field2: types.StringValue("b"),
			},
			Target: &awsEmbeddedStruct{},
			WantTarget: &awsEmbeddedStruct{
				Field1: "a",
				Field2: "b",
			},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
}

// List/Set/Map of string types.
type tfTypedCollectionsOfPrimitiveElements struct {
	Field1 fwtypes.ListValueOf[types.String] `tfsdk:"field1"`
	Field2 fwtypes.ListValueOf[types.String] `tfsdk:"field2"`
	Field3 fwtypes.SetValueOf[types.String]  `tfsdk:"field3"`
	Field4 fwtypes.SetValueOf[types.String]  `tfsdk:"field4"`
	Field5 fwtypes.MapValueOf[types.String]  `tfsdk:"field5"`
	Field6 fwtypes.MapValueOf[types.String]  `tfsdk:"field6"`
}

func TestFlattenCollections(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"zero value slice or map of primitive types Source and Collection of primtive types Target": {
			Source: &awsCollectionsOfPrimitiveElements{},
			Target: &tfCollectionsOfPrimitiveElements{},
			WantTarget: &tfCollectionsOfPrimitiveElements{
				Field1: types.ListNull(types.StringType),
				Field2: types.ListNull(types.StringType),
				Field3: types.SetNull(types.StringType),
				Field4: types.SetNull(types.StringType),
				Field5: types.MapNull(types.StringType),
				Field6: types.MapNull(types.StringType),
			},
		},
		"slice or map of primitive types Source and Collection of primitive types Target": {
			Source: &awsCollectionsOfPrimitiveElements{
				Field1: []string{"a", "b"},
				Field2: aws.StringSlice([]string{"a", "b"}),
				Field3: []string{"a", "b"},
				Field4: aws.StringSlice([]string{"a", "b"}),
				Field5: map[string]string{"A": "a", "B": "b"},
				Field6: aws.StringMap(map[string]string{"A": "a", "B": "b"}),
			},
			Target: &tfCollectionsOfPrimitiveElements{},
			WantTarget: &tfCollectionsOfPrimitiveElements{
				Field1: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field2: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field3: types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field4: types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field5: types.MapValueMust(types.StringType, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
				Field6: types.MapValueMust(types.StringType, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
			},
		},
		"zero value slice or map of string type Source and Collection of string types Target": {
			Source: &awsCollectionsOfPrimitiveElements{},
			Target: &tfTypedCollectionsOfPrimitiveElements{},
			WantTarget: &tfTypedCollectionsOfPrimitiveElements{
				Field1: fwtypes.NewListValueOfNull[types.String](ctx),
				Field2: fwtypes.NewListValueOfNull[types.String](ctx),
				Field3: fwtypes.NewSetValueOfNull[types.String](ctx),
				Field4: fwtypes.NewSetValueOfNull[types.String](ctx),
				Field5: fwtypes.NewMapValueOfNull[types.String](ctx),
				Field6: fwtypes.NewMapValueOfNull[types.String](ctx),
			},
		},
		"slice or map of string types Source and Collection of string types Target": {
			Source: &awsCollectionsOfPrimitiveElements{
				Field1: []string{"a", "b"},
				Field2: aws.StringSlice([]string{"a", "b"}),
				Field3: []string{"a", "b"},
				Field4: aws.StringSlice([]string{"a", "b"}),
				Field5: map[string]string{"A": "a", "B": "b"},
				Field6: aws.StringMap(map[string]string{"A": "a", "B": "b"}),
			},
			Target: &tfTypedCollectionsOfPrimitiveElements{},
			WantTarget: &tfTypedCollectionsOfPrimitiveElements{
				Field1: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field2: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field3: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field4: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field5: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
				Field6: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
}

type awsSimpleStringValueSlice struct {
	Field1 []string
}

type tfSimpleList struct {
	Field1 types.List `tfsdk:"field1"`
}

type tfSimpleListLegacy struct {
	Field1 types.List `tfsdk:"field1" autoflex:",legacy"`
}

func TestFlattenSimpleListOfPrimitiveValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"regular": {
			"values": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{"a", "b"},
				},
				Target: &tfSimpleList{},
				WantTarget: &tfSimpleList{
					Field1: types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("a"),
						types.StringValue("b"),
					}),
				},
			},

			"empty": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{},
				},
				Target: &tfSimpleList{},
				WantTarget: &tfSimpleList{
					Field1: types.ListValueMust(types.StringType, []attr.Value{}),
				},
			},

			"null": {
				Source: awsSimpleStringValueSlice{
					Field1: nil,
				},
				Target: &tfSimpleList{},
				WantTarget: &tfSimpleList{
					Field1: types.ListNull(types.StringType),
				},
			},
		},

		"legacy": {
			"values": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{"a", "b"},
				},
				Target: &tfSimpleListLegacy{},
				WantTarget: &tfSimpleListLegacy{
					Field1: types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("a"),
						types.StringValue("b"),
					}),
				},
			},

			"empty": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{},
				},
				Target: &tfSimpleListLegacy{},
				WantTarget: &tfSimpleListLegacy{
					Field1: types.ListValueMust(types.StringType, []attr.Value{}),
				},
			},

			"null": {
				Source: awsSimpleStringValueSlice{
					Field1: nil,
				},
				Target: &tfSimpleListLegacy{},
				WantTarget: &tfSimpleListLegacy{
					Field1: types.ListValueMust(types.StringType, []attr.Value{}),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

type tfSimpleSet struct {
	Field1 types.Set `tfsdk:"field1"`
}

type tfSimpleSetLegacy struct {
	Field1 types.Set `tfsdk:"field1" autoflex:",legacy"`
}

func TestFlattenSimpleSetOfPrimitiveValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"regular": {
			"values": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{"a", "b"},
				},
				Target: &tfSimpleSet{},
				WantTarget: &tfSimpleSet{
					Field1: types.SetValueMust(types.StringType, []attr.Value{
						types.StringValue("a"),
						types.StringValue("b"),
					}),
				},
			},

			"empty": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{},
				},
				Target: &tfSimpleSet{},
				WantTarget: &tfSimpleSet{
					Field1: types.SetValueMust(types.StringType, []attr.Value{}),
				},
			},

			"null": {
				Source: awsSimpleStringValueSlice{
					Field1: nil,
				},
				Target: &tfSimpleSet{},
				WantTarget: &tfSimpleSet{
					Field1: types.SetNull(types.StringType),
				},
			},
		},

		"legacy": {
			"values": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{"a", "b"},
				},
				Target: &tfSimpleSetLegacy{},
				WantTarget: &tfSimpleSetLegacy{
					Field1: types.SetValueMust(types.StringType, []attr.Value{
						types.StringValue("a"),
						types.StringValue("b"),
					}),
				},
			},

			"empty": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{},
				},
				Target: &tfSimpleSetLegacy{},
				WantTarget: &tfSimpleSetLegacy{
					Field1: types.SetValueMust(types.StringType, []attr.Value{}),
				},
			},

			"null": {
				Source: awsSimpleStringValueSlice{
					Field1: nil,
				},
				Target: &tfSimpleSetLegacy{},
				WantTarget: &tfSimpleSetLegacy{
					Field1: types.SetValueMust(types.StringType, []attr.Value{}),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

func TestFlattenIgnoreStructTag(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"from value": {
			Source: awsSingleStringValue{
				Field1: "value1",
			},
			Target:     &tfSingleStringFieldIgnore{},
			WantTarget: &tfSingleStringFieldIgnore{},
		},
		"from pointer": {
			Source: awsSingleStringPointer{
				Field1: aws.String("value1"),
			},
			Target:     &tfSingleStringFieldIgnore{},
			WantTarget: &tfSingleStringFieldIgnore{},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenStructListOfStringEnum(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"struct with list of string enum": {
			"valid value": {
				Source: awsSliceOfStringEnum{
					Field1: []testEnum{testEnumScalar, testEnumList},
				},
				Target: &tfListOfStringEnum{},
				WantTarget: &tfListOfStringEnum{
					Field1: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
						fwtypes.StringEnumValue(testEnumScalar),
						fwtypes.StringEnumValue(testEnumList),
					}),
				},
			},
			"empty value": {
				Source: awsSliceOfStringEnum{
					Field1: []testEnum{},
				},
				Target: &tfListOfStringEnum{},
				WantTarget: &tfListOfStringEnum{
					Field1: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
				},
			},
			"null value": {
				Source: awsSliceOfStringEnum{},
				Target: &tfListOfStringEnum{},
				WantTarget: &tfListOfStringEnum{
					Field1: fwtypes.NewListValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

func TestFlattenStructSetOfStringEnum(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"struct with set of string enum": {
			"valid value": {
				Source: awsSliceOfStringEnum{
					Field1: []testEnum{testEnumScalar, testEnumList},
				},
				Target: &tfSetOfStringEnum{},
				WantTarget: &tfSetOfStringEnum{
					Field1: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
						fwtypes.StringEnumValue(testEnumScalar),
						fwtypes.StringEnumValue(testEnumList),
					}),
				},
			},
			"empty value": {
				Source: awsSliceOfStringEnum{
					Field1: []testEnum{},
				},
				Target: &tfSetOfStringEnum{},
				WantTarget: &tfSetOfStringEnum{
					Field1: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
				},
			},
			"null value": {
				Source: awsSliceOfStringEnum{},
				Target: &tfSetOfStringEnum{},
				WantTarget: &tfSetOfStringEnum{
					Field1: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

func TestFlattenEmbeddedStruct(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"exported": {
			Source: &awsEmbeddedStruct{
				Field1: "a",
				Field2: "b",
			},
			Target: &tfExportedEmbeddedStruct{},
			WantTarget: &tfExportedEmbeddedStruct{
				TFExportedStruct: TFExportedStruct{
					Field1: types.StringValue("a"),
				},
				Field2: types.StringValue("b"),
			},
		},
		"unexported": {
			Source: &awsEmbeddedStruct{
				Field1: "a",
				Field2: "b",
			},
			Target: &tfUnexportedEmbeddedStruct{},
			WantTarget: &tfUnexportedEmbeddedStruct{
				tfSingleStringField: tfSingleStringField{
					Field1: types.StringValue("a"),
				},
				Field2: types.StringValue("b"),
			},
		},
	}
	// cmp.Diff cannot handle an unexported field
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: false, CompareTarget: true, SkipGoldenLogs: true}, cmpopts.EquateComparable(tfUnexportedEmbeddedStruct{}))
}
