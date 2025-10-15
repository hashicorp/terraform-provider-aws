// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type tfSingleStringFieldLegacy struct {
	Field1 types.String `tfsdk:"field1" autoflex:",legacy"`
}

type tfSingleFloat64Field struct {
	Field1 types.Float64 `tfsdk:"field1"`
}

type tfSingleFloat64FieldLegacy struct {
	Field1 types.Float64 `tfsdk:"field1" autoflex:",legacy"`
}

type tfSingleFloat32Field struct {
	Field1 types.Float32 `tfsdk:"field1"`
}

type tfSingleFloat32FieldLegacy struct {
	Field1 types.Float32 `tfsdk:"field1" autoflex:",legacy"`
}

type tfSingleInt64Field struct {
	Field1 types.Int64 `tfsdk:"field1"`
}

type tfSingleInt64FieldLegacy struct {
	Field1 types.Int64 `tfsdk:"field1" autoflex:",legacy"`
}

type tfSingleInt32Field struct {
	Field1 types.Int32 `tfsdk:"field1"`
}

type tfSingleInt32FieldLegacy struct {
	Field1 types.Int32 `tfsdk:"field1" autoflex:",legacy"`
}

type tfSingleBoolField struct {
	Field1 types.Bool `tfsdk:"field1"`
}

type tfSingleBoolFieldLegacy struct {
	Field1 types.Bool `tfsdk:"field1" autoflex:",legacy"`
}

// All primitive types.
type tfAllThePrimitiveFields struct {
	Field1  types.String  `tfsdk:"field1"`
	Field2  types.String  `tfsdk:"field2"`
	Field3  types.Int64   `tfsdk:"field3"`
	Field4  types.Int64   `tfsdk:"field4"`
	Field5  types.Int64   `tfsdk:"field5"`
	Field6  types.Int64   `tfsdk:"field6"`
	Field7  types.Float64 `tfsdk:"field7"`
	Field8  types.Float64 `tfsdk:"field8"`
	Field9  types.Float64 `tfsdk:"field9"`
	Field10 types.Float64 `tfsdk:"field10"`
	Field11 types.Bool    `tfsdk:"field11"`
	Field12 types.Bool    `tfsdk:"field12"`
}

type awsAllThePrimitiveFields struct {
	Field1  string
	Field2  *string
	Field3  int32
	Field4  *int32
	Field5  int64
	Field6  *int64
	Field7  float32
	Field8  *float32
	Field9  float64
	Field10 *float64
	Field11 bool
	Field12 *bool
}

type awsSingleBoolValue struct {
	Field1 bool
}

type awsSingleBoolPointer struct {
	Field1 *bool
}

type awsSingleStringValue struct {
	Field1 string
}

type awsSingleStringPointer struct {
	Field1 *string
}

type awsSingleByteSliceValue struct {
	Field1 []byte
}

type awsSingleFloat64Value struct {
	Field1 float64
}

type awsSingleFloat64Pointer struct {
	Field1 *float64
}

type awsSingleFloat32Value struct {
	Field1 float32
}

type awsSingleFloat32Pointer struct {
	Field1 *float32
}

type awsSingleInt64Value struct {
	Field1 int64
}

type awsSingleInt64Pointer struct {
	Field1 *int64
}

type awsSingleInt32Value struct {
	Field1 int32
}

type awsSingleInt32Pointer struct {
	Field1 *int32
}

type testEnum string

// Enum values for SlotShape
const (
	testEnumScalar testEnum = "Scalar"
	testEnumList   testEnum = "List"
)

func (testEnum) Values() []testEnum {
	return []testEnum{
		testEnumScalar,
		testEnumList,
	}
}

func TestExpandString(t *testing.T) {
	t.Parallel()

	testString := "test"
	testStringResult := "a"

	testByteSlice := []byte("test")
	testByteSliceResult := []byte("a")

	testCases := map[string]autoFlexTestCases{
		"String to string": {
			"value": {
				Source: tfSingleStringField{
					Field1: types.StringValue("value"),
				},
				Target: &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{
					Field1: "value",
				},
			},
			"empty": {
				Source: tfSingleStringField{
					Field1: types.StringValue(""),
				},
				Target: &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{
					Field1: "",
				},
			},
			"null": {
				Source: tfSingleStringField{
					Field1: types.StringNull(),
				},
				Target: &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{
					Field1: "",
				},
			},
		},

		"legacy String to string": {
			"value": {
				Source: tfSingleStringFieldLegacy{
					Field1: types.StringValue("value"),
				},
				Target: &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{
					Field1: "value",
				},
			},
			"empty": {
				Source: tfSingleStringFieldLegacy{
					Field1: types.StringValue(""),
				},
				Target: &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{
					Field1: "",
				},
			},
			"null": {
				Source: tfSingleStringFieldLegacy{
					Field1: types.StringNull(),
				},
				Target: &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{
					Field1: "",
				},
			},
		},

		"String to *string": {
			"value": {
				Source: tfSingleStringField{
					Field1: types.StringValue("value"),
				},
				Target: &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{
					Field1: aws.String("value"),
				},
			},
			"empty": {
				Source: tfSingleStringField{
					Field1: types.StringValue(""),
				},
				Target: &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{
					Field1: aws.String(""),
				},
			},
			"null": {
				Source: tfSingleStringField{
					Field1: types.StringNull(),
				},
				Target: &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{
					Field1: nil,
				},
			},
		},

		"legacy String to *string": {
			"value": {
				Source: tfSingleStringFieldLegacy{
					Field1: types.StringValue("value"),
				},
				Target: &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{
					Field1: aws.String("value"),
				},
			},
			"empty": {
				Source: tfSingleStringFieldLegacy{
					Field1: types.StringValue(""),
				},
				Target: &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{
					Field1: nil,
				},
			},
			"null": {
				Source: tfSingleStringFieldLegacy{
					Field1: types.StringNull(),
				},
				Target: &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{
					Field1: nil,
				},
			},
		},

		"types.String to string": {
			"types.String to string": {
				Source:     types.StringValue("a"),
				Target:     &testString,
				WantTarget: &testStringResult,
			},
			"types.String to byte slice": {
				Source:     types.StringValue("a"),
				Target:     &testByteSlice,
				WantTarget: &testByteSliceResult,
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
		})
	}
}

func TestExpandPrimitives(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"primitive types Source and primitive types Target": {
			Source: &tfAllThePrimitiveFields{
				Field1:  types.StringValue("field1"),
				Field2:  types.StringValue("field2"),
				Field3:  types.Int64Value(3),
				Field4:  types.Int64Value(-4),
				Field5:  types.Int64Value(5),
				Field6:  types.Int64Value(-6),
				Field7:  types.Float64Value(7.7),
				Field8:  types.Float64Value(-8.8),
				Field9:  types.Float64Value(9.99),
				Field10: types.Float64Value(-10.101),
				Field11: types.BoolValue(true),
				Field12: types.BoolValue(false),
			},
			Target: &awsAllThePrimitiveFields{},
			WantTarget: &awsAllThePrimitiveFields{
				Field1:  "field1",
				Field2:  aws.String("field2"),
				Field3:  3,
				Field4:  aws.Int32(-4),
				Field5:  5,
				Field6:  aws.Int64(-6),
				Field7:  7.7,
				Field8:  aws.Float32(-8.8),
				Field9:  9.99,
				Field10: aws.Float64(-10.101),
				Field11: true,
				Field12: aws.Bool(false),
			},
		},
		"single string struct pointer Source and empty Target": {
			Source:     &tfSingleStringField{Field1: types.StringValue("a")},
			Target:     &emptyStruct{},
			WantTarget: &emptyStruct{},
		},
		"single string Source and single string Target": {
			Source:     &tfSingleStringField{Field1: types.StringValue("a")},
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{Field1: "a"},
		},
		"single string Source and byte slice Target": {
			Source:     &tfSingleStringField{Field1: types.StringValue("a")},
			Target:     &awsSingleByteSliceValue{},
			WantTarget: &awsSingleByteSliceValue{Field1: []byte("a")},
		},
		"single string Source and single *string Target": {
			Source:     &tfSingleStringField{Field1: types.StringValue("a")},
			Target:     &awsSingleStringPointer{},
			WantTarget: &awsSingleStringPointer{Field1: aws.String("a")},
		},
		"single string Source and single int64 Target": {
			Source:     &tfSingleStringField{Field1: types.StringValue("a")},
			Target:     &awsSingleInt64Value{},
			WantTarget: &awsSingleInt64Value{},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: true})
}

func TestExpandBool(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"Bool to bool": {
			"true": {
				Source: tfSingleBoolField{
					Field1: types.BoolValue(true),
				},
				Target: &awsSingleBoolValue{},
				WantTarget: &awsSingleBoolValue{
					Field1: true,
				},
			},
			"false": {
				Source: tfSingleBoolField{
					Field1: types.BoolValue(false),
				},
				Target: &awsSingleBoolValue{},
				WantTarget: &awsSingleBoolValue{
					Field1: false,
				},
			},
			"null": {
				Source: tfSingleBoolField{
					Field1: types.BoolNull(),
				},
				Target: &awsSingleBoolValue{},
				WantTarget: &awsSingleBoolValue{
					Field1: false,
				},
			},
		},

		"legacy Bool to bool": {
			"true": {
				Source: tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(true),
				},
				Target: &awsSingleBoolValue{},
				WantTarget: &awsSingleBoolValue{
					Field1: true,
				},
			},
			"false": {
				Source: tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(false),
				},
				Target: &awsSingleBoolValue{},
				WantTarget: &awsSingleBoolValue{
					Field1: false,
				},
			},
			"null": {
				Source: tfSingleBoolFieldLegacy{
					Field1: types.BoolNull(),
				},
				Target: &awsSingleBoolValue{},
				WantTarget: &awsSingleBoolValue{
					Field1: false,
				},
			},
		},

		"Bool to *bool": {
			"true": {
				Source: tfSingleBoolField{
					Field1: types.BoolValue(true),
				},
				Target: &awsSingleBoolPointer{},
				WantTarget: &awsSingleBoolPointer{
					Field1: aws.Bool(true),
				},
			},
			"false": {
				Source: tfSingleBoolField{
					Field1: types.BoolValue(false),
				},
				Target: &awsSingleBoolPointer{},
				WantTarget: &awsSingleBoolPointer{
					Field1: aws.Bool(false),
				},
			},
			"null": {
				Source: tfSingleBoolField{
					Field1: types.BoolNull(),
				},
				Target: &awsSingleBoolPointer{},
				WantTarget: &awsSingleBoolPointer{
					Field1: nil,
				},
			},
		},

		"legacy Bool to *bool": {
			"true": {
				Source: tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(true),
				},
				Target: &awsSingleBoolPointer{},
				WantTarget: &awsSingleBoolPointer{
					Field1: aws.Bool(true),
				},
			},
			"false": {
				Source: tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(false),
				},
				Target: &awsSingleBoolPointer{},
				WantTarget: &awsSingleBoolPointer{
					Field1: nil,
				},
			},
			"null": {
				Source: tfSingleBoolFieldLegacy{
					Field1: types.BoolNull(),
				},
				Target: &awsSingleBoolPointer{},
				WantTarget: &awsSingleBoolPointer{
					Field1: nil,
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
		})
	}
}

func TestExpandFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"Float64 to float64": {
			"value": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 42,
				},
			},
			"zero": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 0,
				},
			},
			"null": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 0,
				},
			},
		},

		"legacy Float64 to float64": {
			"value": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 42,
				},
			},
			"zero": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 0,
				},
			},
			"null": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 0,
				},
			},
		},

		"Float64 to *float64": {
			"value": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: aws.Float64(42),
				},
			},
			"zero": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: aws.Float64(0),
				},
			},
			"null": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: nil,
				},
			},
		},

		"legacy Float64 to *float64": {
			"value": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: aws.Float64(42),
				},
			},
			"zero": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: nil,
				},
			},
			"null": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: nil,
				},
			},
		},

		// For historical reasons, Float64 can be expanded to float32 values
		"Float64 to float32": {
			"value": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 42,
				},
			},
			"zero": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
			},
			"null": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
			},
		},

		"legacy Float64 to float32": {
			"value": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 42,
				},
			},
			"zero": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
			},
			"null": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
			},
		},

		"Float64 to *float32": {
			"value": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
			},
			"zero": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: aws.Float32(0),
				},
			},
			"null": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: nil,
				},
			},
		},

		"legacy Float64 to *float32": {
			"value": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
			},
			"zero": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: nil,
				},
			},
			"null": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: nil,
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
		})
	}
}

func TestExpandFloat32(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"Float32 to float32": {
			"value": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(42),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 42,
				},
			},
			"zero": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(0),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
			},
			"null": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
			},
		},

		"legacy Float32 to float32": {
			"value": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(42),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 42,
				},
			},
			"zero": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(0),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
			},
			"null": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
			},
		},

		"Float32 to *float32": {
			"value": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(42),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
			},
			"zero": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(0),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: aws.Float32(0),
				},
			},
			"null": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: nil,
				},
			},
		},

		"legacy Float32 to *float32": {
			"value": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(42),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
			},
			"zero": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(0),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: nil,
				},
			},
			"null": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: nil,
				},
			},
		},

		// Float32 cannot be expanded to float64
		"Float32 to float64": {
			"value": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(42),
				},
				Target:        &awsSingleFloat64Value{},
				expectedDiags: diagAF2[types.Float32, float64](diagExpandingIncompatibleTypes),
			},
			"zero": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(0),
				},
				Target:        &awsSingleFloat64Value{},
				expectedDiags: diagAF2[types.Float32, float64](diagExpandingIncompatibleTypes),
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleFloat32Field{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 0,
				},
			},
		},

		"legacy Float32 to float64": {
			"value": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(42),
				},
				Target:        &awsSingleFloat64Value{},
				expectedDiags: diagAF2[types.Float32, float64](diagExpandingIncompatibleTypes),
			},
			"zero": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(0),
				},
				Target:        &awsSingleFloat64Value{},
				expectedDiags: diagAF2[types.Float32, float64](diagExpandingIncompatibleTypes),
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 0,
				},
			},
		},

		"Float32 to *float64": {
			"value": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(42),
				},
				Target:        &awsSingleFloat64Pointer{},
				expectedDiags: diagAF2[types.Float32, *float64](diagExpandingIncompatibleTypes),
			},
			"zero": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(0),
				},
				Target:        &awsSingleFloat64Pointer{},
				expectedDiags: diagAF2[types.Float32, *float64](diagExpandingIncompatibleTypes),
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleFloat32Field{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: nil,
				},
			},
		},

		"legacy Float32 to *float64": {
			"value": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(42),
				},
				Target:        &awsSingleFloat64Pointer{},
				expectedDiags: diagAF2[types.Float32, *float64](diagExpandingIncompatibleTypes),
			},
			"zero": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(0),
				},
				Target:        &awsSingleFloat64Pointer{},
				expectedDiags: diagAF2[types.Float32, *float64](diagExpandingIncompatibleTypes),
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: nil,
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
		})
	}
}

func TestExpandInt64(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"Int64 to int64": {
			"value": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{
					Field1: 42,
				},
			},
			"zero": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{
					Field1: 0,
				},
			},
			"null": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{
					Field1: 0,
				},
			},
		},

		"legacy Int64 to int64": {
			"value": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{
					Field1: 42,
				},
			},
			"zero": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{
					Field1: 0,
				},
			},
			"null": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{
					Field1: 0,
				},
			},
		},

		"Int64 to *int64": {
			"value": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{
					Field1: aws.Int64(42),
				},
			},
			"zero": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{
					Field1: aws.Int64(0),
				},
			},
			"null": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{
					Field1: nil,
				},
			},
		},

		"legacy Int64 to *int64": {
			"value": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{
					Field1: aws.Int64(42),
				},
			},
			"zero": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{
					Field1: nil,
				},
			},
			"null": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{
					Field1: nil,
				},
			},
		},

		// For historical reasons, Int64 can be expanded to int32 values
		"Int64 to int32": {
			"value": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 42,
				},
			},
			"zero": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
			},
			"null": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
			},
		},

		"legacy Int64 to int32": {
			"value": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 42,
				},
			},
			"zero": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
			},
			"null": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
			},
		},

		"Int64 to *int32": {
			"value": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
			},
			"zero": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: aws.Int32(0),
				},
			},
			"null": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: nil,
				},
			},
		},

		"legacy Int64 to *int32": {
			"value": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
			},
			"zero": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: nil,
				},
			},
			"null": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: nil,
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
		})
	}
}

func TestExpandInt32(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"Int32 to int32": {
			"value": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(42),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 42,
				},
			},
			"zero": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(0),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
			},
			"null": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Null(),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
			},
		},

		"legacy Int32 to int32": {
			"value": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(42),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 42,
				},
			},
			"zero": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(0),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
			},
			"null": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Null(),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
			},
		},

		"Int32 to *int32": {
			"value": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(42),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
			},
			"zero": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(0),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: aws.Int32(0),
				},
			},
			"null": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Null(),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: nil,
				},
			},
		},

		"legacy Int32 to *int32": {
			"value": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(42),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
			},
			"zero": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(0),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: nil,
				},
			},
			"null": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Null(),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: nil,
				},
			},
		},

		// Int32 cannot be expanded to int64
		"Int32 to int64": {
			"value": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(42),
				},
				Target:        &awsSingleInt64Value{},
				expectedDiags: diagAF2[types.Int32, int64](diagExpandingIncompatibleTypes),
			},
			"zero": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(0),
				},
				Target:        &awsSingleInt64Value{},
				expectedDiags: diagAF2[types.Int32, int64](diagExpandingIncompatibleTypes),
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleInt32Field{
					Field1: types.Int32Null(),
				},
				Target:     &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{},
			},
		},

		"legacy Int32 to int64": {
			"value": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(42),
				},
				Target:        &awsSingleInt64Value{},
				expectedDiags: diagAF2[types.Int32, int64](diagExpandingIncompatibleTypes),
			},
			"zero": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(0),
				},
				Target:        &awsSingleInt64Value{},
				expectedDiags: diagAF2[types.Int32, int64](diagExpandingIncompatibleTypes),
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Null(),
				},
				Target:     &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{},
			},
		},

		"Int32 to *int64": {
			"value": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(42),
				},
				Target:        &awsSingleInt64Pointer{},
				expectedDiags: diagAF2[types.Int32, *int64](diagExpandingIncompatibleTypes),
			},
			"zero": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(0),
				},
				Target:        &awsSingleInt64Pointer{},
				expectedDiags: diagAF2[types.Int32, *int64](diagExpandingIncompatibleTypes),
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleInt32Field{
					Field1: types.Int32Null(),
				},
				Target:     &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{},
			},
		},

		"legacy Int32 to *int64": {
			"value": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(42),
				},
				Target:        &awsSingleInt64Pointer{},
				expectedDiags: diagAF2[types.Int32, *int64](diagExpandingIncompatibleTypes),
			},
			"zero": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(0),
				},
				Target:        &awsSingleInt64Pointer{},
				expectedDiags: diagAF2[types.Int32, *int64](diagExpandingIncompatibleTypes),
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Null(),
				},
				Target:     &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
		})
	}
}

func TestExpandStringEnum(t *testing.T) {
	t.Parallel()

	var enum testEnum
	enumList := testEnumList

	testCases := autoFlexTestCases{
		"valid value": {
			Source:     fwtypes.StringEnumValue(testEnumList),
			Target:     &enum,
			WantTarget: &enumList,
		},
		"empty value": {
			Source:     fwtypes.StringEnumNull[testEnum](),
			Target:     &enum,
			WantTarget: &enum,
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
}

func TestFlattenPrimitives(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"single empty string Source and single string Target": {
			Source:     &awsSingleStringValue{},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{Field1: types.StringValue("")},
		},
		"single string Source and single string Target": {
			Source:     &awsSingleStringValue{Field1: "a"},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{Field1: types.StringValue("a")},
		},
		"single byte slice Source and single string Target": {
			Source:     &awsSingleByteSliceValue{Field1: []byte("a")},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{Field1: types.StringValue("a")},
		},
		"single nil *string Source and single string Target": {
			Source:     &awsSingleStringPointer{},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{Field1: types.StringNull()},
		},
		"single *string Source and single string Target": {
			Source:     &awsSingleStringPointer{Field1: aws.String("a")},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{Field1: types.StringValue("a")},
		},
		"single string Source and single int64 Target": {
			Source:     &awsSingleStringValue{Field1: "a"},
			Target:     &tfSingleInt64Field{},
			WantTarget: &tfSingleInt64Field{},
		},
		"single string struct pointer Source and empty Target": {
			Source:     &awsSingleStringValue{Field1: "a"},
			Target:     &emptyStruct{},
			WantTarget: &emptyStruct{},
		},
		"primitive pack zero ok": {
			Source: &awsAllThePrimitiveFields{},
			Target: &tfAllThePrimitiveFields{},
			WantTarget: &tfAllThePrimitiveFields{
				Field1:  types.StringValue(""),
				Field2:  types.StringNull(),
				Field3:  types.Int64Value(0),
				Field4:  types.Int64Null(),
				Field5:  types.Int64Value(0),
				Field6:  types.Int64Null(),
				Field7:  types.Float64Value(0),
				Field8:  types.Float64Null(),
				Field9:  types.Float64Value(0),
				Field10: types.Float64Null(),
				Field11: types.BoolValue(false),
				Field12: types.BoolNull(),
			},
		},
		"primitive pack ok": {
			Source: &awsAllThePrimitiveFields{
				Field1:  "field1",
				Field2:  aws.String("field2"),
				Field3:  3,
				Field4:  aws.Int32(-4),
				Field5:  5,
				Field6:  aws.Int64(-6),
				Field7:  7.7,
				Field8:  aws.Float32(-8.8),
				Field9:  9.99,
				Field10: aws.Float64(-10.101),
				Field11: true,
				Field12: aws.Bool(false),
			},
			Target: &tfAllThePrimitiveFields{},
			WantTarget: &tfAllThePrimitiveFields{
				Field1:  types.StringValue("field1"),
				Field2:  types.StringValue("field2"),
				Field3:  types.Int64Value(3),
				Field4:  types.Int64Value(-4),
				Field5:  types.Int64Value(5),
				Field6:  types.Int64Value(-6),
				Field7:  types.Float64Value(7.7),
				Field8:  types.Float64Value(-8.8),
				Field9:  types.Float64Value(9.99),
				Field10: types.Float64Value(-10.101),
				Field11: types.BoolValue(true),
				Field12: types.BoolValue(false),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, GoldenLogs: true})
}

func TestFlattenBool(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"bool to Bool": {
			"true": {
				Source: awsSingleBoolValue{
					Field1: true,
				},
				Target: &tfSingleBoolField{},
				WantTarget: &tfSingleBoolField{
					Field1: types.BoolValue(true),
				},
			},
			"false": {
				Source: awsSingleBoolValue{
					Field1: false,
				},
				Target: &tfSingleBoolField{},
				WantTarget: &tfSingleBoolField{
					Field1: types.BoolValue(false),
				},
			},
		},

		"*bool to Bool": {
			"true": {
				Source: awsSingleBoolPointer{
					Field1: aws.Bool(true),
				},
				Target: &tfSingleBoolField{},
				WantTarget: &tfSingleBoolField{
					Field1: types.BoolValue(true),
				},
			},
			"false": {
				Source: awsSingleBoolPointer{
					Field1: aws.Bool(false),
				},
				Target: &tfSingleBoolField{},
				WantTarget: &tfSingleBoolField{
					Field1: types.BoolValue(false),
				},
			},
			"null": {
				Source: awsSingleBoolPointer{
					Field1: nil,
				},
				Target: &tfSingleBoolField{},
				WantTarget: &tfSingleBoolField{
					Field1: types.BoolNull(),
				},
			},
		},

		"legacy *bool to Bool": {
			"true": {
				Source: awsSingleBoolPointer{
					Field1: aws.Bool(true),
				},
				Target: &tfSingleBoolFieldLegacy{},
				WantTarget: &tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(true),
				},
			},
			"false": {
				Source: awsSingleBoolPointer{
					Field1: aws.Bool(false),
				},
				Target: &tfSingleBoolFieldLegacy{},
				WantTarget: &tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(false),
				},
			},
			"null": {
				Source: awsSingleBoolPointer{
					Field1: nil,
				},
				Target: &tfSingleBoolFieldLegacy{},
				WantTarget: &tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(false),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
		})
	}
}

func TestFlattenFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"float64 to Float64": {
			"value": {
				Source: awsSingleFloat64Value{
					Field1: 42,
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
			},
			"zero": {
				Source: awsSingleFloat64Value{
					Field1: 0,
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
			},
		},

		"*float64 to Float64": {
			"value": {
				Source: awsSingleFloat64Pointer{
					Field1: aws.Float64(42),
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
			},
			"zero": {
				Source: awsSingleFloat64Pointer{
					Field1: aws.Float64(0),
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
			},
			"null": {
				Source: awsSingleFloat64Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Null(),
				},
			},
		},

		"legacy *float64 to Float64": {
			"value": {
				Source: awsSingleFloat64Pointer{
					Field1: aws.Float64(42),
				},
				Target: &tfSingleFloat64FieldLegacy{},
				WantTarget: &tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(42),
				},
			},
			"zero": {
				Source: awsSingleFloat64Pointer{
					Field1: aws.Float64(0),
				},
				Target: &tfSingleFloat64FieldLegacy{},
				WantTarget: &tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
			},
			"null": {
				Source: awsSingleFloat64Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat64FieldLegacy{},
				WantTarget: &tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
			},
		},

		// For historical reasons, float32 can be flattened to Float64 values
		"float32 to Float64": {
			"value": {
				Source: awsSingleFloat32Value{
					Field1: 42,
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
			},
			"zero": {
				Source: awsSingleFloat32Value{
					Field1: 0,
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
			},
		},

		"*float32 to Float64": {
			"value": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
			},
			"zero": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(0),
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
			},
			"null": {
				Source: awsSingleFloat32Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Null(),
				},
			},
		},

		"legacy *float32 to Float64": {
			"value": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
				Target: &tfSingleFloat64FieldLegacy{},
				WantTarget: &tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(42),
				},
			},
			"zero": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(0),
				},
				Target: &tfSingleFloat64FieldLegacy{},
				WantTarget: &tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
			},
			"null": {
				Source: awsSingleFloat32Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat64FieldLegacy{},
				WantTarget: &tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
		})
	}
}

func TestFlattenFloat32(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"float32 to Float32": {
			"value": {
				Source: awsSingleFloat32Value{
					Field1: 42,
				},
				Target: &tfSingleFloat32Field{},
				WantTarget: &tfSingleFloat32Field{
					Field1: types.Float32Value(42),
				},
			},
			"zero": {
				Source: awsSingleFloat32Value{
					Field1: 0,
				},
				Target: &tfSingleFloat32Field{},
				WantTarget: &tfSingleFloat32Field{
					Field1: types.Float32Value(0),
				},
			},
		},

		"*float32 to Float32": {
			"value": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
				Target: &tfSingleFloat32Field{},
				WantTarget: &tfSingleFloat32Field{
					Field1: types.Float32Value(42),
				},
			},
			"zero": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(0),
				},
				Target: &tfSingleFloat32Field{},
				WantTarget: &tfSingleFloat32Field{
					Field1: types.Float32Value(0),
				},
			},
			"null": {
				Source: awsSingleFloat32Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat32Field{},
				WantTarget: &tfSingleFloat32Field{
					Field1: types.Float32Null(),
				},
			},
		},

		"legacy *float32 to Float32": {
			"value": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
				Target: &tfSingleFloat32FieldLegacy{},
				WantTarget: &tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(42),
				},
			},
			"zero": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(0),
				},
				Target: &tfSingleFloat32FieldLegacy{},
				WantTarget: &tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(0),
				},
			},
			"null": {
				Source: awsSingleFloat32Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat32FieldLegacy{},
				WantTarget: &tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(0),
				},
			},
		},

		// float64 cannot be flattened to Float32
		"float64 to Float32": {
			"value": {
				Source: awsSingleFloat64Value{
					Field1: 42,
				},
				Target:        &tfSingleFloat32Field{},
				expectedDiags: diagAF2[float64, types.Float32](DiagFlatteningIncompatibleTypes),
			},
			"zero": {
				Source: awsSingleFloat64Value{
					Field1: 0,
				},
				Target:        &tfSingleFloat32Field{},
				expectedDiags: diagAF2[float64, types.Float32](DiagFlatteningIncompatibleTypes),
			},
		},

		"*float64 to Float32": {
			"value": {
				Source: awsSingleFloat64Pointer{
					Field1: aws.Float64(42),
				},
				Target:        &tfSingleFloat32Field{},
				expectedDiags: diagAF2[*float64, types.Float32](DiagFlatteningIncompatibleTypes),
			},
			"zero": {
				Source: awsSingleFloat64Pointer{
					Field1: aws.Float64(0),
				},
				Target:        &tfSingleFloat32Field{},
				expectedDiags: diagAF2[*float64, types.Float32](DiagFlatteningIncompatibleTypes),
			},
			"null": {
				Source: awsSingleFloat64Pointer{
					Field1: nil,
				},
				Target:        &tfSingleFloat32Field{},
				expectedDiags: diagAF2[*float64, types.Float32](DiagFlatteningIncompatibleTypes),
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
		})
	}
}

func TestFlattenInt64(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"int64 to Int64": {
			"value": {
				Source: awsSingleInt64Value{
					Field1: 42,
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
			},
			"zero": {
				Source: awsSingleInt64Value{
					Field1: 0,
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
			},
		},

		"*int64 to Int64": {
			"value": {
				Source: awsSingleInt64Pointer{
					Field1: aws.Int64(42),
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
			},
			"zero": {
				Source: awsSingleInt64Pointer{
					Field1: aws.Int64(0),
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
			},
			"null": {
				Source: awsSingleInt64Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Null(),
				},
			},
		},

		"legacy *int64 to Int64": {
			"value": {
				Source: awsSingleInt64Pointer{
					Field1: aws.Int64(42),
				},
				Target: &tfSingleInt64FieldLegacy{},
				WantTarget: &tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(42),
				},
			},
			"zero": {
				Source: awsSingleInt64Pointer{
					Field1: aws.Int64(0),
				},
				Target: &tfSingleInt64FieldLegacy{},
				WantTarget: &tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
			},
			"null": {
				Source: awsSingleInt64Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt64FieldLegacy{},
				WantTarget: &tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
			},
		},

		// For historical reasons, int32 can be flattened to Int64 values
		"int32 to Int64": {
			"value": {
				Source: awsSingleInt32Value{
					Field1: 42,
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
			},
			"zero": {
				Source: awsSingleInt32Value{
					Field1: 0,
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
			},
		},

		"*int32 to Int64": {
			"value": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
			},
			"zero": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(0),
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
			},
			"null": {
				Source: awsSingleInt32Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Null(),
				},
			},
		},

		"legacy *int32 to Int64": {
			"value": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
				Target: &tfSingleInt64FieldLegacy{},
				WantTarget: &tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(42),
				},
			},
			"zero": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(0),
				},
				Target: &tfSingleInt64FieldLegacy{},
				WantTarget: &tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
			},
			"null": {
				Source: awsSingleInt32Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt64FieldLegacy{},
				WantTarget: &tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
		})
	}
}

func TestFlattenInt32(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"int32 to Int32": {
			"value": {
				Source: awsSingleInt32Value{
					Field1: 42,
				},
				Target: &tfSingleInt32Field{},
				WantTarget: &tfSingleInt32Field{
					Field1: types.Int32Value(42),
				},
			},
			"zero": {
				Source: awsSingleInt32Value{
					Field1: 0,
				},
				Target: &tfSingleInt32Field{},
				WantTarget: &tfSingleInt32Field{
					Field1: types.Int32Value(0),
				},
			},
		},

		"*int32 to Int32": {
			"value": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
				Target: &tfSingleInt32Field{},
				WantTarget: &tfSingleInt32Field{
					Field1: types.Int32Value(42),
				},
			},
			"zero": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(0),
				},
				Target: &tfSingleInt32Field{},
				WantTarget: &tfSingleInt32Field{
					Field1: types.Int32Value(0),
				},
			},
			"null": {
				Source: awsSingleInt32Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt32Field{},
				WantTarget: &tfSingleInt32Field{
					Field1: types.Int32Null(),
				},
			},
		},

		"legacy *int32 to Int32": {
			"value": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
				Target: &tfSingleInt32FieldLegacy{},
				WantTarget: &tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(42),
				},
			},
			"zero": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(0),
				},
				Target: &tfSingleInt32FieldLegacy{},
				WantTarget: &tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(0),
				},
			},
			"null": {
				Source: awsSingleInt32Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt32FieldLegacy{},
				WantTarget: &tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(0),
				},
			},
		},

		// int64 cannot be flattened to Int32
		"int64 to Int32": {
			"value": {
				Source: awsSingleInt64Value{
					Field1: 42,
				},
				Target:        &tfSingleInt32Field{},
				expectedDiags: diagAF2[int64, types.Int32](DiagFlatteningIncompatibleTypes),
			},
			"zero": {
				Source: awsSingleInt64Value{
					Field1: 0,
				},
				Target:        &tfSingleInt32Field{},
				expectedDiags: diagAF2[int64, types.Int32](DiagFlatteningIncompatibleTypes),
			},
		},

		"*int64 to Int32": {
			"value": {
				Source: awsSingleInt64Pointer{
					Field1: aws.Int64(42),
				},
				Target:        &tfSingleInt32Field{},
				expectedDiags: diagAF2[*int64, types.Int32](DiagFlatteningIncompatibleTypes),
			},
			"zero": {
				Source: awsSingleInt64Pointer{
					Field1: aws.Int64(0),
				},
				Target:        &tfSingleInt32Field{},
				expectedDiags: diagAF2[*int64, types.Int32](DiagFlatteningIncompatibleTypes),
			},
			"null": {
				Source: awsSingleInt64Pointer{
					Field1: nil,
				},
				Target:        &tfSingleInt32Field{},
				expectedDiags: diagAF2[*int64, types.Int32](DiagFlatteningIncompatibleTypes),
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
		})
	}
}

type tfSingleStringFieldOmitEmpty struct {
	Field1 types.String `tfsdk:"field1" autoflex:",omitempty"`
}

func TestFlattenString(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"string to String": {
			"value": {
				Source: awsSingleStringValue{
					Field1: "a",
				},
				Target: &tfSingleStringField{},
				WantTarget: &tfSingleStringField{
					Field1: types.StringValue("a"),
				},
			},
			"zero": {
				Source: awsSingleStringValue{
					Field1: "",
				},
				Target: &tfSingleStringField{},
				WantTarget: &tfSingleStringField{
					Field1: types.StringValue(""),
				},
			},
		},

		"*string to String": {
			"value": {
				Source: awsSingleStringPointer{
					Field1: aws.String("a"),
				},
				Target: &tfSingleStringField{},
				WantTarget: &tfSingleStringField{
					Field1: types.StringValue("a"),
				},
			},
			"zero": {
				Source: awsSingleStringPointer{
					Field1: aws.String(""),
				},
				Target: &tfSingleStringField{},
				WantTarget: &tfSingleStringField{
					Field1: types.StringValue(""),
				},
			},
			"null": {
				Source: awsSingleStringPointer{
					Field1: nil,
				},
				Target: &tfSingleStringField{},
				WantTarget: &tfSingleStringField{
					Field1: types.StringNull(),
				},
			},
		},

		"omitempty string to String": {
			"value": {
				Source: awsSingleStringValue{
					Field1: "a",
				},
				Target: &tfSingleStringFieldOmitEmpty{},
				WantTarget: &tfSingleStringFieldOmitEmpty{
					Field1: types.StringValue("a"),
				},
			},
			"zero": {
				Source: awsSingleStringValue{
					Field1: "",
				},
				Target: &tfSingleStringFieldOmitEmpty{},
				WantTarget: &tfSingleStringFieldOmitEmpty{
					Field1: types.StringNull(),
				},
			},
		},

		"omitempty *string to String": {
			"value": {
				Source: awsSingleStringPointer{
					Field1: aws.String("a"),
				},
				Target: &tfSingleStringFieldOmitEmpty{},
				WantTarget: &tfSingleStringFieldOmitEmpty{
					Field1: types.StringValue("a"),
				},
			},
			"zero": {
				Source: awsSingleStringPointer{
					Field1: aws.String(""),
				},
				Target: &tfSingleStringFieldOmitEmpty{},
				WantTarget: &tfSingleStringFieldOmitEmpty{
					Field1: types.StringNull(),
				},
			},
			"null": {
				Source: awsSingleStringPointer{
					Field1: nil,
				},
				Target: &tfSingleStringFieldOmitEmpty{},
				WantTarget: &tfSingleStringFieldOmitEmpty{
					Field1: types.StringNull(),
				},
			},
		},

		"legacy *string to String": {
			"value": {
				Source: awsSingleStringPointer{
					Field1: aws.String("a"),
				},
				Target: &tfSingleStringFieldLegacy{},
				WantTarget: &tfSingleStringFieldLegacy{
					Field1: types.StringValue("a"),
				},
			},
			"zero": {
				Source: awsSingleStringPointer{
					Field1: aws.String(""),
				},
				Target: &tfSingleStringFieldLegacy{},
				WantTarget: &tfSingleStringFieldLegacy{
					Field1: types.StringValue(""),
				},
			},
			"null": {
				Source: awsSingleStringPointer{
					Field1: nil,
				},
				Target: &tfSingleStringFieldLegacy{},
				WantTarget: &tfSingleStringFieldLegacy{
					Field1: types.StringValue(""),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, CompareLogs: false})
		})
	}
}

func TestFlattenTopLevelStringPtr(t *testing.T) {
	t.Parallel()

	testCases := toplevelTestCases[*string, types.String]{
		"value": {
			source:        aws.String("value"),
			expectedValue: types.StringValue("value"),
			expectedDiags: diagAFEmpty(),
		},

		"empty": {
			source:        aws.String(""),
			expectedValue: types.StringValue(""),
			expectedDiags: diagAFEmpty(),
		},

		"nil": {
			source:        nil,
			expectedValue: types.StringNull(),
			expectedDiags: diagAFEmpty(),
		},
	}

	runTopLevelTestCases(t, testCases, runChecks{CompareDiags: true, CompareLogs: false, CompareTarget: true})
}

func TestFlattenTopLevelInt64Ptr(t *testing.T) {
	t.Parallel()

	testCases := toplevelTestCases[*int64, types.Int64]{
		"value": {
			source:        aws.Int64(42),
			expectedValue: types.Int64Value(42),
			expectedDiags: diagAFEmpty(),
		},

		"empty": {
			source:        aws.Int64(0),
			expectedValue: types.Int64Value(0),
			expectedDiags: diagAFEmpty(),
		},

		"nil": {
			source:        nil,
			expectedValue: types.Int64Null(),
			expectedDiags: diagAFEmpty(),
		},
	}

	runTopLevelTestCases(t, testCases, runChecks{CompareDiags: true, CompareLogs: false, CompareTarget: true})
}
