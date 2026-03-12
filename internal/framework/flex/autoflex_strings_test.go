// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

// Tests AutoFlex's Expand/Flatten of strings and string-like types.
// Additional, foundational string tests are in autoflex_primitives_test.go.

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type tfSingleStringFieldLegacy struct {
	Field1 types.String `tfsdk:"field1" autoflex:",legacy"`
}

type awsSingleStringValue struct {
	Field1 string
}

type awsSingleStringPointer struct {
	Field1 *string
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
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
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
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

type tfSingleStringFieldOmitEmpty struct {
	Field1 types.String `tfsdk:"field1" autoflex:",omitempty"`
}

func TestFlattenString(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
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

			runAutoFlattenTestCases(t, cases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
		})
	}
}

func TestFlattenStringSpecial(t *testing.T) {
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
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true})
}

func TestFlattenTopLevelStringPtr(t *testing.T) {
	t.Parallel()

	testCases := toplevelTestCases[*string, types.String]{
		"value": {
			source:        aws.String("value"),
			expectedValue: types.StringValue("value"),
			ExpectedDiags: diagAFEmpty(),
		},

		"empty": {
			source:        aws.String(""),
			expectedValue: types.StringValue(""),
			ExpectedDiags: diagAFEmpty(),
		},

		"nil": {
			source:        nil,
			expectedValue: types.StringNull(),
			ExpectedDiags: diagAFEmpty(),
		},
	}

	runTopLevelTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}
