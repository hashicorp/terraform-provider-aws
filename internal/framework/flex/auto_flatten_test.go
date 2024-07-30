// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	smithyjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestFlatten(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var (
		typedNilSource *TestFlex00
		typedNilTarget *TestFlex00
	)

	testString := "test"

	testARN := "arn:aws:securityhub:us-west-2:1234567890:control/cis-aws-foundations-benchmark/v/1.2.0/1.1" //lintignore:AWSAT003,AWSAT005

	testTimeStr := "2013-09-25T09:34:01Z"
	testTimeTime := errs.Must(time.Parse(time.RFC3339, testTimeStr))
	var zeroTime time.Time

	testCases := autoFlexTestCases{
		{
			TestName: "nil Source",
			Target:   &TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Cannot flatten nil source"),
				diag.NewErrorDiagnostic("AutoFlEx", "Flatten[<nil>, *flex.TestFlex00]"),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(nil, reflect.TypeFor[*TestFlex00]()),
			},
		},
		{
			TestName: "typed nil Source",
			Source:   typedNilSource,
			Target:   &TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Cannot flatten nil source"),
				diag.NewErrorDiagnostic("AutoFlEx", "Flatten[*flex.TestFlex00, *flex.TestFlex00]"),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlex00](), reflect.TypeFor[*TestFlex00]()),
			},
		},
		{
			TestName: "nil Target",
			Source:   TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Target cannot be nil"),
				diag.NewErrorDiagnostic("AutoFlEx", "Flatten[flex.TestFlex00, <nil>]"),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[TestFlex00](), nil),
			},
		},
		{
			TestName: "typed nil Target",
			Source:   TestFlex00{},
			Target:   typedNilTarget,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Target cannot be nil"),
				diag.NewErrorDiagnostic("AutoFlEx", "Flatten[flex.TestFlex00, *flex.TestFlex00]"),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[TestFlex00](), reflect.TypeFor[*TestFlex00]()),
			},
		},
		{
			TestName: "non-pointer Target",
			Source:   TestFlex00{},
			Target:   0,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "target (int): int, want pointer"),
				diag.NewErrorDiagnostic("AutoFlEx", "Flatten[flex.TestFlex00, int]"),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[TestFlex00](), reflect.TypeFor[int]()),
			},
		},
		{
			TestName: "non-struct Source",
			Source:   testString,
			Target:   &TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "does not implement attr.Value: struct"),
				diag.NewErrorDiagnostic("AutoFlEx", "Flatten[string, *flex.TestFlex00]"),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[string](), reflect.TypeFor[*TestFlex00]()),
			},
		},
		{
			TestName: "non-struct Target",
			Source:   TestFlex00{},
			Target:   &testString,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "does not implement attr.Value: string"),
				diag.NewErrorDiagnostic("AutoFlEx", "Flatten[flex.TestFlex00, *string]"),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[TestFlex00](), reflect.TypeFor[*string]()),
			},
		},
		{
			TestName: "json interface Source string Target",
			Source: &TestFlexAWS19{
				Field1: &testJSONDocument{
					Value: &struct {
						Test string `json:"test"`
					}{
						Test: "a",
					},
				},
			},
			Target: &TestFlexTF19{},
			WantTarget: &TestFlexTF19{
				Field1: types.StringValue(`{"test":"a"}`),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS19](), reflect.TypeFor[*TestFlexTF19]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS19](), "Field1", reflect.TypeFor[*TestFlexTF19]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName: "json interface Source JSONValue Target",
			Source: &TestFlexAWS19{
				Field1: &testJSONDocument{
					Value: &struct {
						Test string `json:"test"`
					}{
						Test: "a",
					},
				},
			},
			Target: &TestFlexTF20{},
			WantTarget: &TestFlexTF20{
				Field1: fwtypes.SmithyJSONValue(`{"test":"a"}`, newTestJSONDocument),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS19](), reflect.TypeFor[*TestFlexTF20]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS19](), "Field1", reflect.TypeFor[*TestFlexTF20]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[fwtypes.SmithyJSON[smithyjson.JSONStringer]]()),
			},
		},
		{
			TestName:   "empty struct Source and Target",
			Source:     TestFlex00{},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[TestFlex00](), reflect.TypeFor[*TestFlex00]()),
			},
		},
		{
			TestName:   "empty struct pointer Source and Target",
			Source:     &TestFlex00{},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlex00](), reflect.TypeFor[*TestFlex00]()),
			},
		},
		{
			TestName:   "single string struct pointer Source and empty Target",
			Source:     &TestFlexAWS01{Field1: "a"},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS01](), reflect.TypeFor[*TestFlex00]()),
				noCorrespondingFieldLogLine(reflect.TypeFor[*TestFlexAWS01](), "Field1", reflect.TypeFor[*TestFlex00]()),
			},
		},
		{
			TestName: "does not implement attr.Value Target",
			Source:   &TestFlexAWS01{Field1: "a"},
			Target:   &TestFlexAWS01{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "does not implement attr.Value: string"),
				diag.NewErrorDiagnostic("AutoFlEx", "convert (Field1)"),
				diag.NewErrorDiagnostic("AutoFlEx", "Flatten[*flex.TestFlexAWS01, *flex.TestFlexAWS01]"),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS01](), reflect.TypeFor[*TestFlexAWS01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1", reflect.TypeFor[*TestFlexAWS01]()),
			},
		},
		{
			TestName:   "single empty string Source and single string Target",
			Source:     &TestFlexAWS01{},
			Target:     &TestFlexTF01{},
			WantTarget: &TestFlexTF01{Field1: types.StringValue("")},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS01](), reflect.TypeFor[*TestFlexTF01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName:   "single string Source and single string Target",
			Source:     &TestFlexAWS01{Field1: "a"},
			Target:     &TestFlexTF01{},
			WantTarget: &TestFlexTF01{Field1: types.StringValue("a")},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS01](), reflect.TypeFor[*TestFlexTF01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName:   "single nil *string Source and single string Target",
			Source:     &TestFlexAWS02{},
			Target:     &TestFlexTF01{},
			WantTarget: &TestFlexTF01{Field1: types.StringNull()},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS02](), reflect.TypeFor[*TestFlexTF01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS02](), "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName:   "single *string Source and single string Target",
			Source:     &TestFlexAWS02{Field1: aws.String("a")},
			Target:     &TestFlexTF01{},
			WantTarget: &TestFlexTF01{Field1: types.StringValue("a")},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS02](), reflect.TypeFor[*TestFlexTF01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS02](), "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName:   "single string Source and single int64 Target",
			Source:     &TestFlexAWS01{Field1: "a"},
			Target:     &TestFlexTF02{},
			WantTarget: &TestFlexTF02{},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS01](), reflect.TypeFor[*TestFlexTF02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1", reflect.TypeFor[*TestFlexTF02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.Int64]()),
				{
					"@level":             "info",
					"@module":            "provider.autoflex",
					"@message":           "AutoFlex Flatten; incompatible types",
					"from":               float64(reflect.String),
					"to":                 map[string]any{},
					logAttrKeySourcePath: "Field1",
					logAttrKeySourceType: fullTypeName(reflect.TypeFor[string]()),
					logAttrKeyTargetPath: "Field1",
					logAttrKeyTargetType: fullTypeName(reflect.TypeFor[types.Int64]()),
				},
			},
		},
		{
			TestName: "zero value primtive types Source and primtive types Target",
			Source:   &TestFlexAWS04{},
			Target:   &TestFlexTF03{},
			WantTarget: &TestFlexTF03{
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
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS04](), reflect.TypeFor[*TestFlexTF03]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS04](), "Field1", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*TestFlexAWS04](), "Field2", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[*string](), "Field2", reflect.TypeFor[types.String]()),
				matchedFieldsLogLine("Field3", reflect.TypeFor[*TestFlexAWS04](), "Field3", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field3", reflect.TypeFor[int32](), "Field3", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field4", reflect.TypeFor[*TestFlexAWS04](), "Field4", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field4", reflect.TypeFor[*int32](), "Field4", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field5", reflect.TypeFor[*TestFlexAWS04](), "Field5", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field5", reflect.TypeFor[int64](), "Field5", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field6", reflect.TypeFor[*TestFlexAWS04](), "Field6", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field6", reflect.TypeFor[*int64](), "Field6", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field7", reflect.TypeFor[*TestFlexAWS04](), "Field7", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field7", reflect.TypeFor[float32](), "Field7", reflect.TypeFor[types.Float64]()),
				matchedFieldsLogLine("Field8", reflect.TypeFor[*TestFlexAWS04](), "Field8", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field8", reflect.TypeFor[*float32](), "Field8", reflect.TypeFor[types.Float64]()),
				matchedFieldsLogLine("Field9", reflect.TypeFor[*TestFlexAWS04](), "Field9", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field9", reflect.TypeFor[float64](), "Field9", reflect.TypeFor[types.Float64]()),
				matchedFieldsLogLine("Field10", reflect.TypeFor[*TestFlexAWS04](), "Field10", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field10", reflect.TypeFor[*float64](), "Field10", reflect.TypeFor[types.Float64]()),
				matchedFieldsLogLine("Field11", reflect.TypeFor[*TestFlexAWS04](), "Field11", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field11", reflect.TypeFor[bool](), "Field11", reflect.TypeFor[types.Bool]()),
				matchedFieldsLogLine("Field12", reflect.TypeFor[*TestFlexAWS04](), "Field12", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field12", reflect.TypeFor[*bool](), "Field12", reflect.TypeFor[types.Bool]()),
			},
		},
		{
			TestName: "primtive types Source and primtive types Target",
			Source: &TestFlexAWS04{
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
			Target: &TestFlexTF03{},
			WantTarget: &TestFlexTF03{
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
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS04](), reflect.TypeFor[*TestFlexTF03]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS04](), "Field1", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*TestFlexAWS04](), "Field2", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[*string](), "Field2", reflect.TypeFor[types.String]()),
				matchedFieldsLogLine("Field3", reflect.TypeFor[*TestFlexAWS04](), "Field3", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field3", reflect.TypeFor[int32](), "Field3", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field4", reflect.TypeFor[*TestFlexAWS04](), "Field4", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field4", reflect.TypeFor[*int32](), "Field4", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field5", reflect.TypeFor[*TestFlexAWS04](), "Field5", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field5", reflect.TypeFor[int64](), "Field5", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field6", reflect.TypeFor[*TestFlexAWS04](), "Field6", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field6", reflect.TypeFor[*int64](), "Field6", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field7", reflect.TypeFor[*TestFlexAWS04](), "Field7", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field7", reflect.TypeFor[float32](), "Field7", reflect.TypeFor[types.Float64]()),
				matchedFieldsLogLine("Field8", reflect.TypeFor[*TestFlexAWS04](), "Field8", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field8", reflect.TypeFor[*float32](), "Field8", reflect.TypeFor[types.Float64]()),
				matchedFieldsLogLine("Field9", reflect.TypeFor[*TestFlexAWS04](), "Field9", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field9", reflect.TypeFor[float64](), "Field9", reflect.TypeFor[types.Float64]()),
				matchedFieldsLogLine("Field10", reflect.TypeFor[*TestFlexAWS04](), "Field10", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field10", reflect.TypeFor[*float64](), "Field10", reflect.TypeFor[types.Float64]()),
				matchedFieldsLogLine("Field11", reflect.TypeFor[*TestFlexAWS04](), "Field11", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field11", reflect.TypeFor[bool](), "Field11", reflect.TypeFor[types.Bool]()),
				matchedFieldsLogLine("Field12", reflect.TypeFor[*TestFlexAWS04](), "Field12", reflect.TypeFor[*TestFlexTF03]()),
				convertingWithPathLogLine("Field12", reflect.TypeFor[*bool](), "Field12", reflect.TypeFor[types.Bool]()),
			},
		},
		{
			TestName: "zero value slice or map of primtive types Source and Collection of primtive types Target",
			Source:   &TestFlexAWS05{},
			Target:   &TestFlexTF04{},
			WantTarget: &TestFlexTF04{
				Field1: types.ListNull(types.StringType),
				Field2: types.ListNull(types.StringType),
				Field3: types.SetNull(types.StringType),
				Field4: types.SetNull(types.StringType),
				Field5: types.MapNull(types.StringType),
				Field6: types.MapNull(types.StringType),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS05](), reflect.TypeFor[*TestFlexTF04]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS05](), "Field1", reflect.TypeFor[*TestFlexTF04]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*TestFlexAWS05](), "Field2", reflect.TypeFor[*TestFlexTF04]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[[]*string](), "Field2", reflect.TypeFor[types.List]()),
				matchedFieldsLogLine("Field3", reflect.TypeFor[*TestFlexAWS05](), "Field3", reflect.TypeFor[*TestFlexTF04]()),
				convertingWithPathLogLine("Field3", reflect.TypeFor[[]string](), "Field3", reflect.TypeFor[types.Set]()),
				matchedFieldsLogLine("Field4", reflect.TypeFor[*TestFlexAWS05](), "Field4", reflect.TypeFor[*TestFlexTF04]()),
				convertingWithPathLogLine("Field4", reflect.TypeFor[[]*string](), "Field4", reflect.TypeFor[types.Set]()),
				matchedFieldsLogLine("Field5", reflect.TypeFor[*TestFlexAWS05](), "Field5", reflect.TypeFor[*TestFlexTF04]()),
				convertingWithPathLogLine("Field5", reflect.TypeFor[map[string]string](), "Field5", reflect.TypeFor[types.Map]()),
				flattenMapNullLogLine("Field5", reflect.TypeFor[map[string]string](), "Field5", reflect.TypeFor[types.Map]()),
				matchedFieldsLogLine("Field6", reflect.TypeFor[*TestFlexAWS05](), "Field6", reflect.TypeFor[*TestFlexTF04]()),
				convertingWithPathLogLine("Field6", reflect.TypeFor[map[string]*string](), "Field6", reflect.TypeFor[types.Map]()),
				flattenMapNullLogLine("Field6", reflect.TypeFor[map[string]*string](), "Field6", reflect.TypeFor[types.Map]()),
			},
		},
		{
			TestName: "slice or map of primtive types Source and Collection of primtive types Target",
			Source: &TestFlexAWS05{
				Field1: []string{"a", "b"},
				Field2: aws.StringSlice([]string{"a", "b"}),
				Field3: []string{"a", "b"},
				Field4: aws.StringSlice([]string{"a", "b"}),
				Field5: map[string]string{"A": "a", "B": "b"},
				Field6: aws.StringMap(map[string]string{"A": "a", "B": "b"}),
			},
			Target: &TestFlexTF04{},
			WantTarget: &TestFlexTF04{
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
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS05](), reflect.TypeFor[*TestFlexTF04]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS05](), "Field1", reflect.TypeFor[*TestFlexTF04]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*TestFlexAWS05](), "Field2", reflect.TypeFor[*TestFlexTF04]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[[]*string](), "Field2", reflect.TypeFor[types.List]()),
				matchedFieldsLogLine("Field3", reflect.TypeFor[*TestFlexAWS05](), "Field3", reflect.TypeFor[*TestFlexTF04]()),
				convertingWithPathLogLine("Field3", reflect.TypeFor[[]string](), "Field3", reflect.TypeFor[types.Set]()),
				matchedFieldsLogLine("Field4", reflect.TypeFor[*TestFlexAWS05](), "Field4", reflect.TypeFor[*TestFlexTF04]()),
				convertingWithPathLogLine("Field4", reflect.TypeFor[[]*string](), "Field4", reflect.TypeFor[types.Set]()),
				matchedFieldsLogLine("Field5", reflect.TypeFor[*TestFlexAWS05](), "Field5", reflect.TypeFor[*TestFlexTF04]()),
				convertingWithPathLogLine("Field5", reflect.TypeFor[map[string]string](), "Field5", reflect.TypeFor[types.Map]()),
				flattenMapValueLogLine("Field5", reflect.TypeFor[map[string]string](), 2, "Field5", reflect.TypeFor[types.Map]()),
				matchedFieldsLogLine("Field6", reflect.TypeFor[*TestFlexAWS05](), "Field6", reflect.TypeFor[*TestFlexTF04]()),
				convertingWithPathLogLine("Field6", reflect.TypeFor[map[string]*string](), "Field6", reflect.TypeFor[types.Map]()),
				flattenMapValueLogLine("Field6", reflect.TypeFor[map[string]*string](), 2, "Field6", reflect.TypeFor[types.Map]()),
			},
		},
		{
			TestName: "zero value slice or map of string type Source and Collection of string types Target",
			Source:   &TestFlexAWS05{},
			Target:   &TestFlexTF18{},
			WantTarget: &TestFlexTF18{
				Field1: fwtypes.NewListValueOfNull[types.String](ctx),
				Field2: fwtypes.NewListValueOfNull[types.String](ctx),
				Field3: fwtypes.NewSetValueOfNull[types.String](ctx),
				Field4: fwtypes.NewSetValueOfNull[types.String](ctx),
				Field5: fwtypes.NewMapValueOfNull[types.String](ctx),
				Field6: fwtypes.NewMapValueOfNull[types.String](ctx),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS05](), reflect.TypeFor[*TestFlexTF18]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS05](), "Field1", reflect.TypeFor[*TestFlexTF18]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*TestFlexAWS05](), "Field2", reflect.TypeFor[*TestFlexTF18]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[[]*string](), "Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				matchedFieldsLogLine("Field3", reflect.TypeFor[*TestFlexAWS05](), "Field3", reflect.TypeFor[*TestFlexTF18]()),
				convertingWithPathLogLine("Field3", reflect.TypeFor[[]string](), "Field3", reflect.TypeFor[fwtypes.SetValueOf[types.String]]()),
				matchedFieldsLogLine("Field4", reflect.TypeFor[*TestFlexAWS05](), "Field4", reflect.TypeFor[*TestFlexTF18]()),
				convertingWithPathLogLine("Field4", reflect.TypeFor[[]*string](), "Field4", reflect.TypeFor[fwtypes.SetValueOf[types.String]]()),
				matchedFieldsLogLine("Field5", reflect.TypeFor[*TestFlexAWS05](), "Field5", reflect.TypeFor[*TestFlexTF18]()),
				convertingWithPathLogLine("Field5", reflect.TypeFor[map[string]string](), "Field5", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				flattenMapNullLogLine("Field5", reflect.TypeFor[map[string]string](), "Field5", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				matchedFieldsLogLine("Field6", reflect.TypeFor[*TestFlexAWS05](), "Field6", reflect.TypeFor[*TestFlexTF18]()),
				convertingWithPathLogLine("Field6", reflect.TypeFor[map[string]*string](), "Field6", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				flattenMapNullLogLine("Field6", reflect.TypeFor[map[string]*string](), "Field6", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
		{
			TestName: "slice or map of string types Source and Collection of string types Target",
			Source: &TestFlexAWS05{
				Field1: []string{"a", "b"},
				Field2: aws.StringSlice([]string{"a", "b"}),
				Field3: []string{"a", "b"},
				Field4: aws.StringSlice([]string{"a", "b"}),
				Field5: map[string]string{"A": "a", "B": "b"},
				Field6: aws.StringMap(map[string]string{"A": "a", "B": "b"}),
			},
			Target: &TestFlexTF18{},
			WantTarget: &TestFlexTF18{
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
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS05](), reflect.TypeFor[*TestFlexTF18]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS05](), "Field1", reflect.TypeFor[*TestFlexTF18]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*TestFlexAWS05](), "Field2", reflect.TypeFor[*TestFlexTF18]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[[]*string](), "Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				matchedFieldsLogLine("Field3", reflect.TypeFor[*TestFlexAWS05](), "Field3", reflect.TypeFor[*TestFlexTF18]()),
				convertingWithPathLogLine("Field3", reflect.TypeFor[[]string](), "Field3", reflect.TypeFor[fwtypes.SetValueOf[types.String]]()),
				matchedFieldsLogLine("Field4", reflect.TypeFor[*TestFlexAWS05](), "Field4", reflect.TypeFor[*TestFlexTF18]()),
				convertingWithPathLogLine("Field4", reflect.TypeFor[[]*string](), "Field4", reflect.TypeFor[fwtypes.SetValueOf[types.String]]()),
				matchedFieldsLogLine("Field5", reflect.TypeFor[*TestFlexAWS05](), "Field5", reflect.TypeFor[*TestFlexTF18]()),
				convertingWithPathLogLine("Field5", reflect.TypeFor[map[string]string](), "Field5", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				flattenMapValueLogLine("Field5", reflect.TypeFor[map[string]string](), 2, "Field5", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				matchedFieldsLogLine("Field6", reflect.TypeFor[*TestFlexAWS05](), "Field6", reflect.TypeFor[*TestFlexTF18]()),
				convertingWithPathLogLine("Field6", reflect.TypeFor[map[string]*string](), "Field6", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				flattenMapValueLogLine("Field6", reflect.TypeFor[map[string]*string](), 2, "Field6", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
		{
			TestName: "plural ordinary field names",
			Source: &TestFlexAWS10{
				Fields: []TestFlexAWS01{{Field1: "a"}},
			},
			Target: &TestFlexTF08{},
			WantTarget: &TestFlexTF08{
				Field: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &TestFlexTF01{
					Field1: types.StringValue("a"),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS10](), reflect.TypeFor[*TestFlexTF08]()),
				matchedFieldsLogLine("Fields", reflect.TypeFor[*TestFlexAWS10](), "Field", reflect.TypeFor[*TestFlexTF08]()),
				convertingWithPathLogLine("Fields", reflect.TypeFor[[]TestFlexAWS01](), "Field", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
				matchedFieldsWithPathLogLine("Fields[0]", "Field1", reflect.TypeFor[TestFlexAWS01](), "Field[0]", "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Fields[0].Field1", reflect.TypeFor[string](), "Field[0].Field1", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName: "plural field names",
			Source: &TestFlexAWS11{
				Cities: []*string{
					aws.String("paris"),
					aws.String("london"),
				},
				Coaches: []*string{
					aws.String("guardiola"),
					aws.String("mourinho"),
				},
				Tomatoes: []*string{
					aws.String("brandywine"),
					aws.String("roma"),
				},
				Vertices: []*string{
					aws.String("ab"),
					aws.String("bc"),
				},
				Criteria: []*string{
					aws.String("votes"),
					aws.String("editors"),
				},
				Data: []*string{
					aws.String("d1282f78-fa99-5d9d-bd51-e6f0173eb74a"),
					aws.String("0f10cb10-2076-5254-bd21-d3f62fe66303"),
				},
				Hives: []*string{
					aws.String("Cegieme"),
					aws.String("Fahumvid"),
				},
			},
			Target: &TestFlexTF09{},
			WantTarget: &TestFlexTF09{
				City: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("paris"),
					types.StringValue("london"),
				}),
				Coach: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("guardiola"),
					types.StringValue("mourinho"),
				}),
				Tomato: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("brandywine"),
					types.StringValue("roma"),
				}),
				Vertex: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("ab"),
					types.StringValue("bc"),
				}),
				Criterion: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("votes"),
					types.StringValue("editors"),
				}),
				Datum: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("d1282f78-fa99-5d9d-bd51-e6f0173eb74a"),
					types.StringValue("0f10cb10-2076-5254-bd21-d3f62fe66303"),
				}),
				Hive: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("Cegieme"),
					types.StringValue("Fahumvid"),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS11](), reflect.TypeFor[*TestFlexTF09]()),
				matchedFieldsLogLine("Cities", reflect.TypeFor[*TestFlexAWS11](), "City", reflect.TypeFor[*TestFlexTF09]()),
				convertingWithPathLogLine("Cities", reflect.TypeFor[[]*string](), "City", reflect.TypeFor[types.List]()),
				matchedFieldsLogLine("Coaches", reflect.TypeFor[*TestFlexAWS11](), "Coach", reflect.TypeFor[*TestFlexTF09]()),
				convertingWithPathLogLine("Coaches", reflect.TypeFor[[]*string](), "Coach", reflect.TypeFor[types.List]()),
				matchedFieldsLogLine("Tomatoes", reflect.TypeFor[*TestFlexAWS11](), "Tomato", reflect.TypeFor[*TestFlexTF09]()),
				convertingWithPathLogLine("Tomatoes", reflect.TypeFor[[]*string](), "Tomato", reflect.TypeFor[types.List]()),
				matchedFieldsLogLine("Vertices", reflect.TypeFor[*TestFlexAWS11](), "Vertex", reflect.TypeFor[*TestFlexTF09]()),
				convertingWithPathLogLine("Vertices", reflect.TypeFor[[]*string](), "Vertex", reflect.TypeFor[types.List]()),
				matchedFieldsLogLine("Criteria", reflect.TypeFor[*TestFlexAWS11](), "Criterion", reflect.TypeFor[*TestFlexTF09]()),
				convertingWithPathLogLine("Criteria", reflect.TypeFor[[]*string](), "Criterion", reflect.TypeFor[types.List]()),
				matchedFieldsLogLine("Data", reflect.TypeFor[*TestFlexAWS11](), "Datum", reflect.TypeFor[*TestFlexTF09]()),
				convertingWithPathLogLine("Data", reflect.TypeFor[[]*string](), "Datum", reflect.TypeFor[types.List]()),
				matchedFieldsLogLine("Hives", reflect.TypeFor[*TestFlexAWS11](), "Hive", reflect.TypeFor[*TestFlexTF09]()),
				convertingWithPathLogLine("Hives", reflect.TypeFor[[]*string](), "Hive", reflect.TypeFor[types.List]()),
			},
		},
		{
			TestName: "strange plurality",
			Source: &TestFlexPluralityAWS01{
				Value:  "a",
				Values: "b",
			},
			Target: &TestFlexPluralityTF01{},
			WantTarget: &TestFlexPluralityTF01{
				Value: types.StringValue("a"),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexPluralityAWS01](), reflect.TypeFor[*TestFlexPluralityTF01]()),
				matchedFieldsLogLine("Value", reflect.TypeFor[*TestFlexPluralityAWS01](), "Value", reflect.TypeFor[*TestFlexPluralityTF01]()),
				convertingWithPathLogLine("Value", reflect.TypeFor[string](), "Value", reflect.TypeFor[types.String]()),
				noCorrespondingFieldLogLine(reflect.TypeFor[*TestFlexPluralityAWS01](), "Values", reflect.TypeFor[*TestFlexPluralityTF01]()),
			},
		},
		{
			TestName: "capitalization field names",
			Source: &TestFlexAWS12{
				FieldUrl: aws.String("h"),
			},
			Target: &TestFlexTF10{},
			WantTarget: &TestFlexTF10{
				FieldURL: types.StringValue("h"),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS12](), reflect.TypeFor[*TestFlexTF10]()),
				matchedFieldsLogLine("FieldUrl", reflect.TypeFor[*TestFlexAWS12](), "FieldURL", reflect.TypeFor[*TestFlexTF10]()),
				convertingWithPathLogLine("FieldUrl", reflect.TypeFor[*string](), "FieldURL", reflect.TypeFor[types.String]()),
			},
		},
		{
			ContextFn: func(ctx context.Context) context.Context { return context.WithValue(ctx, ResourcePrefix, "Intent") },
			TestName:  "resource name prefix",
			Source: &TestFlexAWS18{
				IntentName: aws.String("Ovodoghen"),
			},
			Target: &TestFlexTF16{},
			WantTarget: &TestFlexTF16{
				Name: types.StringValue("Ovodoghen"),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS18](), reflect.TypeFor[*TestFlexTF16]()),
				matchedFieldsLogLine("IntentName", reflect.TypeFor[*TestFlexAWS18](), "Name", reflect.TypeFor[*TestFlexTF16]()),
				convertingWithPathLogLine("IntentName", reflect.TypeFor[*string](), "Name", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName:   "single string Source and single ARN Target",
			Source:     &TestFlexAWS01{Field1: testARN},
			Target:     &TestFlexTF17{},
			WantTarget: &TestFlexTF17{Field1: fwtypes.ARNValue(testARN)},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS01](), reflect.TypeFor[*TestFlexTF17]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1", reflect.TypeFor[*TestFlexTF17]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[fwtypes.ARN]()),
			},
		},
		{
			TestName:   "single *string Source and single ARN Target",
			Source:     &TestFlexAWS02{Field1: aws.String(testARN)},
			Target:     &TestFlexTF17{},
			WantTarget: &TestFlexTF17{Field1: fwtypes.ARNValue(testARN)},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS02](), reflect.TypeFor[*TestFlexTF17]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS02](), "Field1", reflect.TypeFor[*TestFlexTF17]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[fwtypes.ARN]()),
			},
		},
		{
			TestName:   "single nil *string Source and single ARN Target",
			Source:     &TestFlexAWS02{},
			Target:     &TestFlexTF17{},
			WantTarget: &TestFlexTF17{Field1: fwtypes.ARNNull()},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS02](), reflect.TypeFor[*TestFlexTF17]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS02](), "Field1", reflect.TypeFor[*TestFlexTF17]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[fwtypes.ARN]()),
			},
		},
		{
			TestName: "timestamp",
			Source: &TestFlexTimeAWS02{
				CreationDateTime: testTimeTime,
			},
			Target: &TestFlexTimeTF01{},
			WantTarget: &TestFlexTimeTF01{
				CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexTimeAWS02](), reflect.TypeFor[*TestFlexTimeTF01]()),
				matchedFieldsLogLine("CreationDateTime", reflect.TypeFor[*TestFlexTimeAWS02](), "CreationDateTime", reflect.TypeFor[*TestFlexTimeTF01]()),
				convertingWithPathLogLine("CreationDateTime", reflect.TypeFor[time.Time](), "CreationDateTime", reflect.TypeFor[timetypes.RFC3339]()),
			},
		},
		{
			TestName: "timestamp pointer",
			Source: &TestFlexTimeAWS01{
				CreationDateTime: &testTimeTime,
			},
			Target: &TestFlexTimeTF01{},
			WantTarget: &TestFlexTimeTF01{
				CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexTimeAWS01](), reflect.TypeFor[*TestFlexTimeTF01]()),
				matchedFieldsLogLine("CreationDateTime", reflect.TypeFor[*TestFlexTimeAWS01](), "CreationDateTime", reflect.TypeFor[*TestFlexTimeTF01]()),
				convertingWithPathLogLine("CreationDateTime", reflect.TypeFor[*time.Time](), "CreationDateTime", reflect.TypeFor[timetypes.RFC3339]()),
			},
		},
		{
			TestName: "timestamp nil",
			Source:   &TestFlexTimeAWS01{},
			Target:   &TestFlexTimeTF01{},
			WantTarget: &TestFlexTimeTF01{
				CreationDateTime: timetypes.NewRFC3339Null(),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexTimeAWS01](), reflect.TypeFor[*TestFlexTimeTF01]()),
				matchedFieldsLogLine("CreationDateTime", reflect.TypeFor[*TestFlexTimeAWS01](), "CreationDateTime", reflect.TypeFor[*TestFlexTimeTF01]()),
				convertingWithPathLogLine("CreationDateTime", reflect.TypeFor[*time.Time](), "CreationDateTime", reflect.TypeFor[timetypes.RFC3339]()),
			},
		},
		{
			TestName: "timestamp empty",
			Source:   &TestFlexTimeAWS02{},
			Target:   &TestFlexTimeTF01{},
			WantTarget: &TestFlexTimeTF01{
				CreationDateTime: timetypes.NewRFC3339TimeValue(zeroTime),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexTimeAWS02](), reflect.TypeFor[*TestFlexTimeTF01]()),
				matchedFieldsLogLine("CreationDateTime", reflect.TypeFor[*TestFlexTimeAWS02](), "CreationDateTime", reflect.TypeFor[*TestFlexTimeTF01]()),
				convertingWithPathLogLine("CreationDateTime", reflect.TypeFor[time.Time](), "CreationDateTime", reflect.TypeFor[timetypes.RFC3339]()),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenGeneric(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		{
			TestName:   "nil *struct Source and single list Target",
			Source:     &TestFlexAWS06{},
			Target:     &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx)},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS06](), reflect.TypeFor[*TestFlexTF05]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS06](), "Field1", reflect.TypeFor[*TestFlexTF05]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
			},
		},
		{
			TestName:   "*struct Source and single list Target",
			Source:     &TestFlexAWS06{Field1: &TestFlexAWS01{Field1: "a"}},
			Target:     &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &TestFlexTF01{Field1: types.StringValue("a")})},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS06](), reflect.TypeFor[*TestFlexTF05]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS06](), "Field1", reflect.TypeFor[*TestFlexTF05]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
				matchedFieldsWithPathLogLine("Field1", "Field1", reflect.TypeFor[TestFlexAWS01](), "Field1", "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1.Field1", reflect.TypeFor[string](), "Field1.Field1", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName:   "*struct Source and single set Target",
			Source:     &TestFlexAWS06{Field1: &TestFlexAWS01{Field1: "a"}},
			Target:     &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &TestFlexTF01{Field1: types.StringValue("a")})},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS06](), reflect.TypeFor[*TestFlexTF06]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS06](), "Field1", reflect.TypeFor[*TestFlexTF06]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]]()),
				matchedFieldsWithPathLogLine("Field1", "Field1", reflect.TypeFor[TestFlexAWS01](), "Field1", "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1.Field1", reflect.TypeFor[string](), "Field1.Field1", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName:   "nil []struct and null list Target",
			Source:     &TestFlexAWS08{},
			Target:     &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx)},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS08](), reflect.TypeFor[*TestFlexTF05]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS08](), "Field1", reflect.TypeFor[*TestFlexTF05]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
			},
		},
		{
			TestName:   "nil []struct and null set Target",
			Source:     &TestFlexAWS08{},
			Target:     &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfNull[TestFlexTF01](ctx)},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS08](), reflect.TypeFor[*TestFlexTF06]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS08](), "Field1", reflect.TypeFor[*TestFlexTF06]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]]()),
			},
		},
		{
			TestName:   "empty []struct and empty list Target",
			Source:     &TestFlexAWS08{Field1: []TestFlexAWS01{}},
			Target:     &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{})},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS08](), reflect.TypeFor[*TestFlexTF05]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS08](), "Field1", reflect.TypeFor[*TestFlexTF05]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
			},
		},
		{
			TestName:   "empty []struct and empty set Target",
			Source:     &TestFlexAWS08{Field1: []TestFlexAWS01{}},
			Target:     &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{})},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS08](), reflect.TypeFor[*TestFlexTF06]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS08](), "Field1", reflect.TypeFor[*TestFlexTF06]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]]()),
			},
		},
		{
			TestName: "non-empty []struct and non-empty list Target",
			Source: &TestFlexAWS08{Field1: []TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
			Target: &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS08](), reflect.TypeFor[*TestFlexTF05]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS08](), "Field1", reflect.TypeFor[*TestFlexTF05]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
				matchedFieldsWithPathLogLine("Field1[0]", "Field1", reflect.TypeFor[TestFlexAWS01](), "Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1[0].Field1", reflect.TypeFor[string](), "Field1[0].Field1", reflect.TypeFor[types.String]()),
				matchedFieldsWithPathLogLine("Field1[1]", "Field1", reflect.TypeFor[TestFlexAWS01](), "Field1[1]", "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1[1].Field1", reflect.TypeFor[string](), "Field1[1].Field1", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName: "non-empty []struct and non-empty set Target",
			Source: &TestFlexAWS08{Field1: []TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
			Target: &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS08](), reflect.TypeFor[*TestFlexTF06]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS08](), "Field1", reflect.TypeFor[*TestFlexTF06]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]]()),
				matchedFieldsWithPathLogLine("Field1[0]", "Field1", reflect.TypeFor[TestFlexAWS01](), "Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1[0].Field1", reflect.TypeFor[string](), "Field1[0].Field1", reflect.TypeFor[types.String]()),
				matchedFieldsWithPathLogLine("Field1[1]", "Field1", reflect.TypeFor[TestFlexAWS01](), "Field1[1]", "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1[1].Field1", reflect.TypeFor[string](), "Field1[1].Field1", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName:   "nil []*struct and null list Target",
			Source:     &TestFlexAWS07{},
			Target:     &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx)},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS07](), reflect.TypeFor[*TestFlexTF05]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS07](), "Field1", reflect.TypeFor[*TestFlexTF05]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]*TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
			},
		},
		{
			TestName:   "nil []*struct and null set Target",
			Source:     &TestFlexAWS07{},
			Target:     &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfNull[TestFlexTF01](ctx)},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS07](), reflect.TypeFor[*TestFlexTF06]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS07](), "Field1", reflect.TypeFor[*TestFlexTF06]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]*TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]]()),
			},
		},
		{
			TestName:   "empty []*struct and empty list Target",
			Source:     &TestFlexAWS07{Field1: []*TestFlexAWS01{}},
			Target:     &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*TestFlexTF01{})},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS07](), reflect.TypeFor[*TestFlexTF05]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS07](), "Field1", reflect.TypeFor[*TestFlexTF05]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]*TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
			},
		},
		{
			TestName:   "empty []*struct and empty set Target",
			Source:     &TestFlexAWS07{Field1: []*TestFlexAWS01{}},
			Target:     &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*TestFlexTF01{})},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS07](), reflect.TypeFor[*TestFlexTF06]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS07](), "Field1", reflect.TypeFor[*TestFlexTF06]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]*TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]]()),
			},
		},
		{
			TestName: "non-empty []*struct and non-empty list Target",
			Source: &TestFlexAWS07{Field1: []*TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
			Target: &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS07](), reflect.TypeFor[*TestFlexTF05]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS07](), "Field1", reflect.TypeFor[*TestFlexTF05]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]*TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
				matchedFieldsWithPathLogLine("Field1[0]", "Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1[0].Field1", reflect.TypeFor[string](), "Field1[0].Field1", reflect.TypeFor[types.String]()),
				matchedFieldsWithPathLogLine("Field1[1]", "Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1[1]", "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1[1].Field1", reflect.TypeFor[string](), "Field1[1].Field1", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName: "non-empty []*struct and non-empty set Target",
			Source: &TestFlexAWS07{Field1: []*TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
			Target: &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS07](), reflect.TypeFor[*TestFlexTF06]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS07](), "Field1", reflect.TypeFor[*TestFlexTF06]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]*TestFlexAWS01](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]]()),
				matchedFieldsWithPathLogLine("Field1[0]", "Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1[0].Field1", reflect.TypeFor[string](), "Field1[0].Field1", reflect.TypeFor[types.String]()),
				matchedFieldsWithPathLogLine("Field1[1]", "Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1[1]", "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field1[1].Field1", reflect.TypeFor[string](), "Field1[1].Field1", reflect.TypeFor[types.String]()),
			},
		},
		{
			TestName: "complex Source and complex Target",
			Source: &TestFlexAWS09{
				Field1: "m",
				Field2: &TestFlexAWS06{Field1: &TestFlexAWS01{Field1: "n"}},
				Field3: aws.StringMap(map[string]string{"X": "x", "Y": "y"}),
				Field4: []TestFlexAWS03{{Field1: 100}, {Field1: 2000}, {Field1: 30000}},
			},
			Target: &TestFlexTF07{},
			WantTarget: &TestFlexTF07{
				Field1: types.StringValue("m"),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &TestFlexTF05{
					Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &TestFlexTF01{
						Field1: types.StringValue("n"),
					}),
				}),
				Field3: types.MapValueMust(types.StringType, map[string]attr.Value{
					"X": types.StringValue("x"),
					"Y": types.StringValue("y"),
				}),
				Field4: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF02{
					{Field1: types.Int64Value(100)},
					{Field1: types.Int64Value(2000)},
					{Field1: types.Int64Value(30000)},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS09](), reflect.TypeFor[*TestFlexTF07]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS09](), "Field1", reflect.TypeFor[*TestFlexTF07]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),

				matchedFieldsLogLine("Field2", reflect.TypeFor[*TestFlexAWS09](), "Field2", reflect.TypeFor[*TestFlexTF07]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[*TestFlexAWS06](), "Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF05]]()),
				matchedFieldsWithPathLogLine("Field2", "Field1", reflect.TypeFor[TestFlexAWS06](), "Field2", "Field1", reflect.TypeFor[*TestFlexTF05]()),
				convertingWithPathLogLine("Field2.Field1", reflect.TypeFor[*TestFlexAWS01](), "Field2.Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
				matchedFieldsWithPathLogLine("Field2.Field1", "Field1", reflect.TypeFor[TestFlexAWS01](), "Field2.Field1", "Field1", reflect.TypeFor[*TestFlexTF01]()),
				convertingWithPathLogLine("Field2.Field1.Field1", reflect.TypeFor[string](), "Field2.Field1.Field1", reflect.TypeFor[types.String]()),

				matchedFieldsLogLine("Field3", reflect.TypeFor[*TestFlexAWS09](), "Field3", reflect.TypeFor[*TestFlexTF07]()),
				convertingWithPathLogLine("Field3", reflect.TypeFor[map[string]*string](), "Field3", reflect.TypeFor[basetypes.MapValue]()),
				flattenMapValueLogLine("Field3", reflect.TypeFor[map[string]*string](), 2, "Field3", reflect.TypeFor[basetypes.MapValue]()),

				matchedFieldsLogLine("Field4", reflect.TypeFor[*TestFlexAWS09](), "Field4", reflect.TypeFor[*TestFlexTF07]()),
				convertingWithPathLogLine("Field4", reflect.TypeFor[[]TestFlexAWS03](), "Field4", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF02]]()),
				matchedFieldsWithPathLogLine("Field4[0]", "Field1", reflect.TypeFor[TestFlexAWS03](), "Field4[0]", "Field1", reflect.TypeFor[*TestFlexTF02]()),
				convertingWithPathLogLine("Field4[0].Field1", reflect.TypeFor[int64](), "Field4[0].Field1", reflect.TypeFor[types.Int64]()),
				matchedFieldsWithPathLogLine("Field4[1]", "Field1", reflect.TypeFor[TestFlexAWS03](), "Field4[1]", "Field1", reflect.TypeFor[*TestFlexTF02]()),
				convertingWithPathLogLine("Field4[1].Field1", reflect.TypeFor[int64](), "Field4[1].Field1", reflect.TypeFor[types.Int64]()),
				matchedFieldsWithPathLogLine("Field4[2]", "Field1", reflect.TypeFor[TestFlexAWS03](), "Field4[2]", "Field1", reflect.TypeFor[*TestFlexTF02]()),
				convertingWithPathLogLine("Field4[2].Field1", reflect.TypeFor[int64](), "Field4[2].Field1", reflect.TypeFor[types.Int64]()),
			},
		},
		{
			TestName: "map of string",
			Source: &TestFlexAWS13{
				FieldInner: map[string]string{
					"x": "y",
				},
			},
			Target: &TestFlexTF11{},
			WantTarget: &TestFlexTF11{
				FieldInner: fwtypes.NewMapValueOfMust[basetypes.StringValue](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS13](), reflect.TypeFor[*TestFlexTF11]()),
				matchedFieldsLogLine("FieldInner", reflect.TypeFor[*TestFlexAWS13](), "FieldInner", reflect.TypeFor[*TestFlexTF11]()),
				convertingWithPathLogLine("FieldInner", reflect.TypeFor[map[string]string](), "FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				flattenMapValueLogLine("FieldInner", reflect.TypeFor[map[string]string](), 1, "FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
		{
			TestName: "map of string pointer",
			Source: &awsMapOfStringPointer{
				FieldInner: map[string]*string{
					"x": aws.String("y"),
				},
			},
			Target: &TestFlexTF11{},
			WantTarget: &TestFlexTF11{
				FieldInner: fwtypes.NewMapValueOfMust[basetypes.StringValue](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*awsMapOfStringPointer](), reflect.TypeFor[*TestFlexTF11]()),
				matchedFieldsLogLine("FieldInner", reflect.TypeFor[*awsMapOfStringPointer](), "FieldInner", reflect.TypeFor[*TestFlexTF11]()),
				convertingWithPathLogLine("FieldInner", reflect.TypeFor[map[string]*string](), "FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				flattenMapValueLogLine("FieldInner", reflect.TypeFor[map[string]*string](), 1, "FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
		{
			TestName: "nested string map",
			Source: &TestFlexAWS16{
				FieldOuter: TestFlexAWS13{
					FieldInner: map[string]string{
						"x": "y",
					},
				},
			},
			Target: &TestFlexTF14{},
			WantTarget: &TestFlexTF14{
				FieldOuter: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &TestFlexTF11{
					FieldInner: fwtypes.NewMapValueOfMust[basetypes.StringValue](ctx, map[string]attr.Value{
						"x": types.StringValue("y"),
					}),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS16](), reflect.TypeFor[*TestFlexTF14]()),
				matchedFieldsLogLine("FieldOuter", reflect.TypeFor[*TestFlexAWS16](), "FieldOuter", reflect.TypeFor[*TestFlexTF14]()),
				convertingWithPathLogLine("FieldOuter", reflect.TypeFor[TestFlexAWS13](), "FieldOuter", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF11]]()),
				matchedFieldsWithPathLogLine("FieldOuter", "FieldInner", reflect.TypeFor[TestFlexAWS13](), "FieldOuter", "FieldInner", reflect.TypeFor[*TestFlexTF11]()),
				convertingWithPathLogLine("FieldOuter.FieldInner", reflect.TypeFor[map[string]string](), "FieldOuter.FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				flattenMapValueLogLine("FieldOuter.FieldInner", reflect.TypeFor[map[string]string](), 1, "FieldOuter.FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
		{
			TestName: "map of map of string",
			Source: &TestFlexAWS21{
				Field1: map[string]map[string]string{
					"x": {
						"y": "z",
					},
				},
			},
			Target: &TestFlexTF21{},
			WantTarget: &TestFlexTF21{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS21](), reflect.TypeFor[*TestFlexTF21]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS21](), "Field1", reflect.TypeFor[*TestFlexTF21]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[map[string]map[string]string](), "Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]]()),
				flattenMapLogLine("Field1", reflect.TypeFor[map[string]map[string]string](), 1, "Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]]()),
				flattenNewMapValueOfLogLine("Field1[\"x\"]", reflect.TypeFor[map[string]string](), 1, "Field1[\"x\"]", reflect.TypeFor[map[string]attr.Value]()),
			},
		},
		{
			TestName: "map of map of string pointer",
			Source: &TestFlexAWS22{
				Field1: map[string]map[string]*string{
					"x": {
						"y": aws.String("z"),
					},
				},
			},
			Target: &TestFlexTF21{},
			WantTarget: &TestFlexTF21{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexAWS22](), reflect.TypeFor[*TestFlexTF21]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS22](), "Field1", reflect.TypeFor[*TestFlexTF21]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[map[string]map[string]*string](), "Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]]()),
				flattenMapLogLine("Field1", reflect.TypeFor[map[string]map[string]*string](), 1, "Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]]()),
				flattenNewMapValueOfLogLine("Field1[\"x\"]", reflect.TypeFor[map[string]*string](), 1, "Field1[\"x\"]", reflect.TypeFor[map[string]attr.Value]()),
			},
		},
		{
			TestName: "map block key list",
			Source: &TestFlexMapBlockKeyAWS01{
				MapBlock: map[string]TestFlexMapBlockKeyAWS02{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &TestFlexMapBlockKeyTF01{},
			WantTarget: &TestFlexMapBlockKeyTF01{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust[TestFlexMapBlockKeyTF02](ctx, []TestFlexMapBlockKeyTF02{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexMapBlockKeyAWS01](), reflect.TypeFor[*TestFlexMapBlockKeyTF01]()),
				matchedFieldsLogLine("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS01](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF01]()),
				convertingWithPathLogLine("MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF02]]()),
				matchedFieldsWithPathLogLine("MapBlock", "Attr1", reflect.TypeFor[TestFlexMapBlockKeyAWS02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02]()),
				convertingWithPathLogLine("MapBlock.Attr1", reflect.TypeFor[string](), "MapBlock.Attr1", reflect.TypeFor[types.String]()), // TODO: fix source map index, fix target list index
				matchedFieldsWithPathLogLine("MapBlock", "Attr2", reflect.TypeFor[TestFlexMapBlockKeyAWS02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02]()),
				convertingWithPathLogLine("MapBlock.Attr2", reflect.TypeFor[string](), "MapBlock.Attr2", reflect.TypeFor[types.String]()), // TODO: fix source map index, fix target list index
			},
		},
		{
			TestName: "map block key set",
			Source: &TestFlexMapBlockKeyAWS01{
				MapBlock: map[string]TestFlexMapBlockKeyAWS02{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &TestFlexMapBlockKeyTF03{},
			WantTarget: &TestFlexMapBlockKeyTF03{
				MapBlock: fwtypes.NewSetNestedObjectValueOfValueSliceMust[TestFlexMapBlockKeyTF02](ctx, []TestFlexMapBlockKeyTF02{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexMapBlockKeyAWS01](), reflect.TypeFor[*TestFlexMapBlockKeyTF03]()),
				matchedFieldsLogLine("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS01](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF03]()),
				convertingWithPathLogLine("MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02](), "MapBlock", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexMapBlockKeyTF02]]()),
				matchedFieldsWithPathLogLine("MapBlock", "Attr1", reflect.TypeFor[TestFlexMapBlockKeyAWS02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02]()),
				convertingWithPathLogLine("MapBlock.Attr1", reflect.TypeFor[string](), "MapBlock.Attr1", reflect.TypeFor[types.String]()), // TODO: fix source map index, fix target list index
				matchedFieldsWithPathLogLine("MapBlock", "Attr2", reflect.TypeFor[TestFlexMapBlockKeyAWS02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02]()),
				convertingWithPathLogLine("MapBlock.Attr2", reflect.TypeFor[string](), "MapBlock.Attr2", reflect.TypeFor[types.String]()), // TODO: fix source map index, fix target list index
			},
		},
		{
			TestName: "map block key ptr source",
			Source: &TestFlexMapBlockKeyAWS03{
				MapBlock: map[string]*TestFlexMapBlockKeyAWS02{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &TestFlexMapBlockKeyTF01{},
			WantTarget: &TestFlexMapBlockKeyTF01{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust[TestFlexMapBlockKeyTF02](ctx, []TestFlexMapBlockKeyTF02{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexMapBlockKeyAWS03](), reflect.TypeFor[*TestFlexMapBlockKeyTF01]()),
				matchedFieldsLogLine("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS03](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF01]()),
				convertingWithPathLogLine("MapBlock", reflect.TypeFor[map[string]*TestFlexMapBlockKeyAWS02](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF02]]()),
				matchedFieldsWithPathLogLine("MapBlock", "Attr1", reflect.TypeFor[TestFlexMapBlockKeyAWS02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02]()),
				convertingWithPathLogLine("MapBlock.Attr1", reflect.TypeFor[string](), "MapBlock.Attr1", reflect.TypeFor[types.String]()), // TODO: fix source map index, fix target list index
				matchedFieldsWithPathLogLine("MapBlock", "Attr2", reflect.TypeFor[TestFlexMapBlockKeyAWS02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02]()),
				convertingWithPathLogLine("MapBlock.Attr2", reflect.TypeFor[string](), "MapBlock.Attr2", reflect.TypeFor[types.String]()), // TODO: fix source map index, fix target list index
			},
		},
		{
			TestName: "map block key ptr both",
			Source: &TestFlexMapBlockKeyAWS03{
				MapBlock: map[string]*TestFlexMapBlockKeyAWS02{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &TestFlexMapBlockKeyTF01{},
			WantTarget: &TestFlexMapBlockKeyTF01{
				MapBlock: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*TestFlexMapBlockKeyTF02{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexMapBlockKeyAWS03](), reflect.TypeFor[*TestFlexMapBlockKeyTF01]()),
				matchedFieldsLogLine("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS03](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF01]()),
				convertingWithPathLogLine("MapBlock", reflect.TypeFor[map[string]*TestFlexMapBlockKeyAWS02](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF02]]()),
				matchedFieldsWithPathLogLine("MapBlock", "Attr1", reflect.TypeFor[TestFlexMapBlockKeyAWS02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02]()),
				convertingWithPathLogLine("MapBlock.Attr1", reflect.TypeFor[string](), "MapBlock.Attr1", reflect.TypeFor[types.String]()), // TODO: fix source map index, fix target list index
				matchedFieldsWithPathLogLine("MapBlock", "Attr2", reflect.TypeFor[TestFlexMapBlockKeyAWS02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02]()),
				convertingWithPathLogLine("MapBlock.Attr2", reflect.TypeFor[string](), "MapBlock.Attr2", reflect.TypeFor[types.String]()), // TODO: fix source map index, fix target list index
			},
		},
		{
			TestName: "map block enum key",
			Source: &TestFlexMapBlockKeyAWS01{
				MapBlock: map[string]TestFlexMapBlockKeyAWS02{
					string(TestEnumList): {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &TestFlexMapBlockKeyTF04{},
			WantTarget: &TestFlexMapBlockKeyTF04{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust[TestFlexMapBlockKeyTF05](ctx, []TestFlexMapBlockKeyTF05{
					{
						MapBlockKey: fwtypes.StringEnumValue(TestEnumList),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*TestFlexMapBlockKeyAWS01](), reflect.TypeFor[*TestFlexMapBlockKeyTF04]()),
				matchedFieldsLogLine("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS01](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF04]()),
				convertingWithPathLogLine("MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF05]]()),
				matchedFieldsWithPathLogLine("MapBlock", "Attr1", reflect.TypeFor[TestFlexMapBlockKeyAWS02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF05]()),
				convertingWithPathLogLine("MapBlock.Attr1", reflect.TypeFor[string](), "MapBlock.Attr1", reflect.TypeFor[types.String]()), // TODO: fix source map index, fix target list index
				matchedFieldsWithPathLogLine("MapBlock", "Attr2", reflect.TypeFor[TestFlexMapBlockKeyAWS02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF05]()),
				convertingWithPathLogLine("MapBlock.Attr2", reflect.TypeFor[string](), "MapBlock.Attr2", reflect.TypeFor[types.String]()), // TODO: fix source map index, fix target list index
			},
		},
	}

	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenSimpleNestedBlockWithStringEnum(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Int64                  `tfsdk:"field1"`
		Field2 fwtypes.StringEnum[TestEnum] `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 int64
		Field2 TestEnum
	}

	testCases := autoFlexTestCases{
		{
			TestName: "single nested valid value",
			Source: &aws01{
				Field1: 1,
				Field2: TestEnumList,
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.Int64Value(1),
				Field2: fwtypes.StringEnumValue(TestEnumList),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*aws01](), "Field2", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[TestEnum](), "Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]]()),
			},
		},
		{
			TestName: "single nested empty value",
			Source: &aws01{
				Field1: 1,
				Field2: "",
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.Int64Value(1),
				Field2: fwtypes.StringEnumNull[TestEnum](),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*aws01](), "Field2", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[TestEnum](), "Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenComplexNestedBlockWithStringEnum(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field2 fwtypes.StringEnum[TestEnum] `tfsdk:"field2"`
	}
	type tf02 struct {
		Field1 types.Int64                           `tfsdk:"field1"`
		Field2 fwtypes.ListNestedObjectValueOf[tf01] `tfsdk:"field2"`
	}
	type aws02 struct {
		Field2 TestEnum
	}
	type aws01 struct {
		Field1 int64
		Field2 *aws02
	}

	ctx := context.Background()
	var zero fwtypes.StringEnum[TestEnum]
	testCases := autoFlexTestCases{
		{
			TestName: "single nested valid value",
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field2: TestEnumList,
				},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: fwtypes.StringEnumValue(TestEnumList),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*aws01](), "Field2", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[*aws02](), "Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]]()),
				matchedFieldsWithPathLogLine("Field2", "Field2", reflect.TypeFor[aws02](), "Field2", "Field2", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field2.Field2", reflect.TypeFor[TestEnum](), "Field2.Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]]()),
			},
		},
		{
			TestName: "single nested empty value",
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{Field2: ""},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: fwtypes.StringEnumNull[TestEnum](),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*aws01](), "Field2", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[*aws02](), "Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]]()),
				matchedFieldsWithPathLogLine("Field2", "Field2", reflect.TypeFor[aws02](), "Field2", "Field2", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field2.Field2", reflect.TypeFor[TestEnum](), "Field2.Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]]()),
			},
		},
		{
			TestName: "single nested zero value",
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
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*aws01](), "Field2", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[*aws02](), "Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]]()),
				matchedFieldsWithPathLogLine("Field2", "Field2", reflect.TypeFor[aws02](), "Field2", "Field2", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field2.Field2", reflect.TypeFor[TestEnum](), "Field2.Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
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
		{
			TestName: "single nested block pointer",
			Source: &aws02{
				Field1: &aws01{
					Field1: aws.String("a"),
					Field2: 1,
				},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{
					Field1: types.StringValue("a"),
					Field2: types.Int64Value(1),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws02](), reflect.TypeFor[*tf02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws02](), "Field1", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]]()),
				matchedFieldsWithPathLogLine("Field1", "Field1", reflect.TypeFor[aws01](), "Field1", "Field1", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1.Field1", reflect.TypeFor[*string](), "Field1.Field1", reflect.TypeFor[types.String]()),
				matchedFieldsWithPathLogLine("Field1", "Field2", reflect.TypeFor[aws01](), "Field1", "Field2", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1.Field2", reflect.TypeFor[int64](), "Field1.Field2", reflect.TypeFor[types.Int64]()),
			},
		},
		{
			TestName: "single nested block nil",
			Source:   &aws02{},
			Target:   &tf02{},
			WantTarget: &tf02{
				Field1: fwtypes.NewObjectValueOfNull[tf01](ctx),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws02](), reflect.TypeFor[*tf02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws02](), "Field1", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]]()),
			},
		},
		{
			TestName: "single nested block value",
			Source: &aws03{
				Field1: aws01{
					Field1: aws.String("a"),
					Field2: 1},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{
					Field1: types.StringValue("a"),
					Field2: types.Int64Value(1),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws03](), reflect.TypeFor[*tf02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws03](), "Field1", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]]()),
				matchedFieldsWithPathLogLine("Field1", "Field1", reflect.TypeFor[aws01](), "Field1", "Field1", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1.Field1", reflect.TypeFor[*string](), "Field1.Field1", reflect.TypeFor[types.String]()),
				matchedFieldsWithPathLogLine("Field1", "Field2", reflect.TypeFor[aws01](), "Field1", "Field2", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1.Field2", reflect.TypeFor[int64](), "Field1.Field2", reflect.TypeFor[types.Int64]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
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
		{
			TestName: "single nested block pointer",
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
				Field1: fwtypes.NewObjectValueOfMust[tf02](ctx, &tf02{
					Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{
						Field1: types.BoolValue(true),
						Field2: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
							types.StringValue("a"),
							types.StringValue("b"),
						}),
					}),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws03](), reflect.TypeFor[*tf03]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws03](), "Field1", reflect.TypeFor[*tf03]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*aws02](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf02]]()),
				matchedFieldsWithPathLogLine("Field1", "Field1", reflect.TypeFor[aws02](), "Field1", "Field1", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field1.Field1", reflect.TypeFor[*aws01](), "Field1.Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]]()),
				matchedFieldsWithPathLogLine("Field1.Field1", "Field1", reflect.TypeFor[aws01](), "Field1.Field1", "Field1", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1.Field1.Field1", reflect.TypeFor[bool](), "Field1.Field1.Field1", reflect.TypeFor[types.Bool]()),
				matchedFieldsWithPathLogLine("Field1.Field1", "Field2", reflect.TypeFor[aws01](), "Field1.Field1", "Field2", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1.Field1.Field2", reflect.TypeFor[[]string](), "Field1.Field1.Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
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
		{
			TestName:   "single nested valid value",
			Source:     &aws01{Field1: 1, Field2: aws.Float32(0.01)},
			Target:     &tf01{},
			WantTarget: &tf01{Field1: types.Int64Value(1), Field2: types.Float64Value(0.01)},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*aws01](), "Field2", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[*float32](), "Field2", reflect.TypeFor[types.Float64]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
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
		{
			TestName: "single nested valid value",
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
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*aws01](), "Field2", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[*aws02](), "Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]]()),
				matchedFieldsWithPathLogLine("Field2", "Field1", reflect.TypeFor[aws02](), "Field2", "Field1", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field2.Field1", reflect.TypeFor[float32](), "Field2.Field1", reflect.TypeFor[types.Float64]()),
				matchedFieldsWithPathLogLine("Field2", "Field2", reflect.TypeFor[aws02](), "Field2", "Field2", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field2.Field2", reflect.TypeFor[*float32](), "Field2.Field2", reflect.TypeFor[types.Float64]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
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
		{
			TestName: "single nested valid value",
			Source: &aws01{
				Field1: 1,
				Field2: aws.Float64(0.01),
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.Int64Value(1),
				Field2: types.Float64Value(0.01),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*aws01](), "Field2", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[*float64](), "Field2", reflect.TypeFor[types.Float64]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
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
		{
			TestName: "single nested valid value",
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field1: 1.11,
					Field2: aws.Float64(-2.22),
				},
			},
			Target:     &tf02{},
			WantTarget: &tf02{Field1: types.Int64Value(1), Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{Field1: types.Float64Value(1.11), Field2: types.Float64Value(-2.22)})},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*aws01](), "Field2", reflect.TypeFor[*tf02]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[*aws02](), "Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]]()),
				matchedFieldsWithPathLogLine("Field2", "Field1", reflect.TypeFor[aws02](), "Field2", "Field1", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field2.Field1", reflect.TypeFor[float64](), "Field2.Field1", reflect.TypeFor[types.Float64]()),
				matchedFieldsWithPathLogLine("Field2", "Field2", reflect.TypeFor[aws02](), "Field2", "Field2", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field2.Field2", reflect.TypeFor[*float64](), "Field2.Field2", reflect.TypeFor[types.Float64]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenOptions(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Bool                       `tfsdk:"field1"`
		Tags   fwtypes.MapValueOf[types.String] `tfsdk:"tags"`
	}
	type aws01 struct {
		Field1 bool
		Tags   map[string]string
	}

	// For test cases below where a field of `MapValue` type is ignored, the
	// result of `cmp.Diff` is intentionally not checked.
	//
	// When a target contains an ignored field of a `MapValue` type, the resulting
	// target will contain a zero value, which, because the `elementType` is nil, will
	// always return `false` from the `Equal` method, even when compared with another
	// zero value. In practice, this zeroed `MapValue` would be overwritten
	// by a subsequent step (ie. transparent tagging), and the temporary invalid
	// state of the zeroed `MapValue` will not appear in the final state.
	//
	// Example expected diff:
	// 	    unexpected diff (+wanted, -got):   &flex.tf01{
	//                 Field1: s"false",
	//         -       Tags:   types.MapValueOf[github.com/hashicorp/terraform-plugin-framework/types/basetypes.StringValue]{},
	//         +       Tags:   types.MapValueOf[github.com/hashicorp/terraform-plugin-framework/types/basetypes.StringValue]{MapValue: basetypes.MapValue{elementType: basetypes.StringType{}}},
	//           }
	ctx := context.Background()
	testCases := autoFlexTestCases{
		{
			TestName: "empty source with tags",
			Source:   &aws01{},
			Target:   &tf01{},
			WantTarget: &tf01{
				Field1: types.BoolValue(false),
				Tags:   fwtypes.NewMapValueOfNull[types.String](ctx),
			},
			WantDiff: true, // Ignored MapValue type, expect diff
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[bool](), "Field1", reflect.TypeFor[types.Bool]()),
				ignoredFieldLogLine(reflect.TypeFor[*aws01](), "Tags", reflect.TypeFor[*tf01]()),
			},
		},
		{
			TestName: "ignore tags by default",
			Source: &aws01{
				Field1: true,
				Tags:   map[string]string{"foo": "bar"},
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.BoolValue(true),
				Tags:   fwtypes.NewMapValueOfNull[types.String](ctx),
			},
			WantDiff: true, // Ignored MapValue type, expect diff
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[bool](), "Field1", reflect.TypeFor[types.Bool]()),
				ignoredFieldLogLine(reflect.TypeFor[*aws01](), "Tags", reflect.TypeFor[*tf01]()),
			},
		},
		{
			TestName: "include tags with option override",
			Options: []AutoFlexOptionsFunc{
				func(opts *AutoFlexOptions) {
					opts.SetIgnoredFields([]string{})
				},
			},
			Source: &aws01{
				Field1: true,
				Tags:   map[string]string{"foo": "bar"},
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.BoolValue(true),
				Tags: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"foo": types.StringValue("bar"),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[bool](), "Field1", reflect.TypeFor[types.Bool]()),
				matchedFieldsLogLine("Tags", reflect.TypeFor[*aws01](), "Tags", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Tags", reflect.TypeFor[map[string]string](), "Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				flattenMapValueLogLine("Tags", reflect.TypeFor[map[string]string](), 1, "Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
		{
			TestName: "ignore custom field",
			Options: []AutoFlexOptionsFunc{
				func(opts *AutoFlexOptions) {
					opts.SetIgnoredFields([]string{"Field1"})
				},
			},
			Source: &aws01{
				Field1: true,
				Tags:   map[string]string{"foo": "bar"},
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.BoolNull(),
				Tags: fwtypes.NewMapValueOfMust[types.String](
					ctx,
					map[string]attr.Value{
						"foo": types.StringValue("bar"),
					},
				),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				ignoredFieldLogLine(reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[*tf01]()),
				matchedFieldsLogLine("Tags", reflect.TypeFor[*aws01](), "Tags", reflect.TypeFor[*tf01]()),
				convertingWithPathLogLine("Tags", reflect.TypeFor[map[string]string](), "Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				flattenMapValueLogLine("Tags", reflect.TypeFor[map[string]string](), 1, "Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenInterface(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		{
			TestName: "nil interface Source and list Target",
			Source: testFlexAWSInterfaceSingle{
				Field1: nil,
			},
			Target: &testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{},
			WantTarget: &testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfNull[testFlexTFInterfaceFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSingle](), reflect.TypeFor[*testFlexTFListNestedObject[testFlexTFInterfaceFlexer]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSingle](), "Field1", reflect.TypeFor[*testFlexTFListNestedObject[testFlexTFInterfaceFlexer]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]]()),
			},
		},
		{
			TestName: "single interface Source and single list Target",
			Source: testFlexAWSInterfaceSingle{
				Field1: &testFlexAWSInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{},
			WantTarget: &testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSingle](), reflect.TypeFor[*testFlexTFListNestedObject[testFlexTFInterfaceFlexer]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSingle](), "Field1", reflect.TypeFor[*testFlexTFListNestedObject[testFlexTFInterfaceFlexer]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]]()),
				// StringValueToFramework in testFlexTFInterfaceFlexer.Flatten()
				flatteningWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
		{
			TestName: "nil interface Source and non-Flattener list Target",
			Source: testFlexAWSInterfaceSingle{
				Field1: nil,
			},
			Target: &testFlexTFListNestedObject[TestFlexTF01]{},
			WantTarget: &testFlexTFListNestedObject[TestFlexTF01]{
				Field1: fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSingle](), reflect.TypeFor[*testFlexTFListNestedObject[TestFlexTF01]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSingle](), "Field1", reflect.TypeFor[*testFlexTFListNestedObject[TestFlexTF01]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
			},
		},
		{
			TestName: "single interface Source and non-Flattener list Target",
			Source: testFlexAWSInterfaceSingle{
				Field1: &testFlexAWSInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &testFlexTFListNestedObject[TestFlexTF01]{},
			WantTarget: &testFlexTFListNestedObject[TestFlexTF01]{
				Field1: fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSingle](), reflect.TypeFor[*testFlexTFListNestedObject[TestFlexTF01]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSingle](), "Field1", reflect.TypeFor[*testFlexTFListNestedObject[TestFlexTF01]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
				{
					"@level":   "info",
					"@module":  "provider.autoflex",
					"@message": "AutoFlex Flatten; incompatible types",
					"from":     float64(reflect.Interface),
					"to": map[string]any{
						"ElemType": map[string]any{
							"AttrTypes": map[string]any{
								"field1": map[string]any{},
							},
						},
					},
					logAttrKeySourcePath: "Field1",
					logAttrKeySourceType: fullTypeName(reflect.TypeFor[testFlexAWSInterfaceInterface]()),
					logAttrKeyTargetPath: "Field1",
					logAttrKeyTargetType: fullTypeName(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]]()),
				},
			},
		},

		{
			TestName: "nil interface Source and set Target",
			Source: testFlexAWSInterfaceSingle{
				Field1: nil,
			},
			Target: &testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{},
			WantTarget: &testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfNull[testFlexTFInterfaceFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSingle](), reflect.TypeFor[*testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSingle](), "Field1", reflect.TypeFor[*testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]]()),
			},
		},
		{
			TestName: "single interface Source and single set Target",
			Source: testFlexAWSInterfaceSingle{
				Field1: &testFlexAWSInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{},
			WantTarget: &testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSingle](), reflect.TypeFor[*testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSingle](), "Field1", reflect.TypeFor[*testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]]()),
				// StringValueToFramework in testFlexTFInterfaceFlexer.Flatten()
				flatteningWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},

		{
			TestName: "nil interface list Source and empty list Target",
			Source: testFlexAWSInterfaceSlice{
				Field1: nil,
			},
			Target: &testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{},
			WantTarget: &testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfNull[testFlexTFInterfaceFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSlice](), reflect.TypeFor[*testFlexTFListNestedObject[testFlexTFInterfaceFlexer]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSlice](), "Field1", reflect.TypeFor[*testFlexTFListNestedObject[testFlexTFInterfaceFlexer]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]]()),
			},
		},
		{
			TestName: "empty interface list Source and empty list Target",
			Source: testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{},
			},
			Target: &testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{},
			WantTarget: &testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSlice](), reflect.TypeFor[*testFlexTFListNestedObject[testFlexTFInterfaceFlexer]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSlice](), "Field1", reflect.TypeFor[*testFlexTFListNestedObject[testFlexTFInterfaceFlexer]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]]()),
			},
		},
		{
			TestName: "non-empty interface list Source and non-empty list Target",
			Source: testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{
					&testFlexAWSInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&testFlexAWSInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			Target: &testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{},
			WantTarget: &testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSlice](), reflect.TypeFor[*testFlexTFListNestedObject[testFlexTFInterfaceFlexer]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSlice](), "Field1", reflect.TypeFor[*testFlexTFListNestedObject[testFlexTFInterfaceFlexer]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]]()),
				// StringValueToFramework in testFlexTFInterfaceFlexer.Flatten()
				flatteningWithPathLogLine("Field1[0]", reflect.TypeFor[string](), "Field1[0]", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
				// StringValueToFramework in testFlexTFInterfaceFlexer.Flatten()
				flatteningWithPathLogLine("Field1[1]", reflect.TypeFor[string](), "Field1[1]", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},

		{
			TestName: "nil interface list Source and empty set Target",
			Source: testFlexAWSInterfaceSlice{
				Field1: nil,
			},
			Target: &testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{},
			WantTarget: &testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfNull[testFlexTFInterfaceFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSlice](), reflect.TypeFor[*testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSlice](), "Field1", reflect.TypeFor[*testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]]()),
			},
		},
		{
			TestName: "empty interface list Source and empty set Target",
			Source: testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{},
			},
			Target: &testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{},
			WantTarget: &testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSlice](), reflect.TypeFor[*testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSlice](), "Field1", reflect.TypeFor[*testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]]()),
			},
		},
		{
			TestName: "non-empty interface list Source and non-empty set Target",
			Source: testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{
					&testFlexAWSInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&testFlexAWSInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			Target: &testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{},
			WantTarget: &testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSlice](), reflect.TypeFor[*testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSlice](), "Field1", reflect.TypeFor[*testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]]()),
				// StringValueToFramework in testFlexTFInterfaceFlexer.Flatten()
				flatteningWithPathLogLine("Field1[0]", reflect.TypeFor[string](), "Field1[0]", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
				// StringValueToFramework in testFlexTFInterfaceFlexer.Flatten()
				flatteningWithPathLogLine("Field1[1]", reflect.TypeFor[string](), "Field1[1]", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
		{
			TestName: "nil interface Source and nested object Target",
			Source: testFlexAWSInterfaceSingle{
				Field1: nil,
			},
			Target: &testFlexTFObjectValue[testFlexTFInterfaceFlexer]{},
			WantTarget: &testFlexTFObjectValue[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewObjectValueOfNull[testFlexTFInterfaceFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSingle](), reflect.TypeFor[*testFlexTFObjectValue[testFlexTFInterfaceFlexer]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSingle](), "Field1", reflect.TypeFor[*testFlexTFObjectValue[testFlexTFInterfaceFlexer]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFInterfaceFlexer]]()),
			},
		},
		{
			TestName: "interface Source and nested object Target",
			Source: testFlexAWSInterfaceSingle{
				Field1: &testFlexAWSInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &testFlexTFObjectValue[testFlexTFInterfaceFlexer]{},
			WantTarget: &testFlexTFObjectValue[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &testFlexTFInterfaceFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSInterfaceSingle](), reflect.TypeFor[*testFlexTFObjectValue[testFlexTFInterfaceFlexer]]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceSingle](), "Field1", reflect.TypeFor[*testFlexTFObjectValue[testFlexTFInterfaceFlexer]]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[testFlexAWSInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFInterfaceFlexer]]()),
				// StringValueToFramework in testFlexTFInterfaceFlexer.Flatten()
				flatteningWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenFlattener(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		{
			TestName: "top level struct Source",
			Source: testFlexAWSExpander{
				AWSField: "value1",
			},
			Target: &testFlexTFFlexer{},
			WantTarget: &testFlexTFFlexer{
				Field1: types.StringValue("value1"),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpander](), reflect.TypeFor[*testFlexTFFlexer]()),
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[*types.String]()), // TODO: fix path
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()),  // TODO: fix path
			},
		},
		{
			TestName: "top level incompatible struct Target",
			Source: testFlexAWSExpanderIncompatible{
				Incompatible: 123,
			},
			Target: &testFlexTFFlexer{},
			WantTarget: &testFlexTFFlexer{
				Field1: types.StringNull(),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderIncompatible](), reflect.TypeFor[*testFlexTFFlexer]()),
				// TODO: needs more logging
			},
		},
		{
			TestName: "single struct Source and single list Target",
			Source: testFlexAWSExpanderSingleStruct{
				Field1: testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			Target: &testFlexTFExpanderListNestedObject{},
			WantTarget: &testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderSingleStruct](), reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSExpanderSingleStruct](), "Field1", reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]]()),
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
		{
			TestName: "nil *struct Source and null list Target",
			Source: testFlexAWSExpanderSinglePtr{
				Field1: nil,
			},
			Target: &testFlexTFExpanderListNestedObject{},
			WantTarget: &testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfNull[testFlexTFFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderSinglePtr](), reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSExpanderSinglePtr](), "Field1", reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]]()),
			},
		},
		{
			TestName: "single struct Source and single set Target",
			Source: testFlexAWSExpanderSingleStruct{
				Field1: testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			Target: &testFlexTFExpanderSetNestedObject{},
			WantTarget: &testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderSingleStruct](), reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSExpanderSingleStruct](), "Field1", reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]]()),
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
		{
			TestName: "single *struct Source and single list Target",
			Source: testFlexAWSExpanderSinglePtr{
				Field1: &testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			Target: &testFlexTFExpanderListNestedObject{},
			WantTarget: &testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderSinglePtr](), reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSExpanderSinglePtr](), "Field1", reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]]()),
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
		{
			TestName: "single *struct Source and single set Target",
			Source: testFlexAWSExpanderSinglePtr{
				Field1: &testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			Target: &testFlexTFExpanderSetNestedObject{},
			WantTarget: &testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderSinglePtr](), reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSExpanderSinglePtr](), "Field1", reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]]()),
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
		{
			TestName: "nil *struct Source and null set Target",
			Source: testFlexAWSExpanderSinglePtr{
				Field1: nil,
			},
			Target: &testFlexTFExpanderSetNestedObject{},
			WantTarget: &testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfNull[testFlexTFFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderSinglePtr](), reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSExpanderSinglePtr](), "Field1", reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]]()),
			},
		},

		{
			TestName: "empty struct list Source and empty list Target",
			Source: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{},
			},
			Target: &testFlexTFExpanderListNestedObject{},
			WantTarget: &testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*testFlexAWSExpanderStructSlice](), reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice](), "Field1", reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]]()),
			},
		},
		{
			TestName: "non-empty struct list Source and non-empty list Target",
			Source: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			Target: &testFlexTFExpanderListNestedObject{},
			WantTarget: &testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*testFlexAWSExpanderStructSlice](), reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice](), "Field1", reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]]()),
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1[0]", reflect.TypeFor[string](), "Field1[0]", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1[1]", reflect.TypeFor[string](), "Field1[1]", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
		{
			TestName: "empty *struct list Source and empty list Target",
			Source: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{},
			},
			Target: &testFlexTFExpanderListNestedObject{},
			WantTarget: &testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*testFlexAWSExpanderPtrSlice](), reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice](), "Field1", reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]*testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]]()),
			},
		},
		{
			TestName: "non-empty *struct list Source and non-empty list Target",
			Source: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			Target: &testFlexTFExpanderListNestedObject{},
			WantTarget: &testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[*testFlexAWSExpanderPtrSlice](), reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice](), "Field1", reflect.TypeFor[*testFlexTFExpanderListNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]*testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]]()),
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1[0]", reflect.TypeFor[string](), "Field1[0]", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1[1]", reflect.TypeFor[string](), "Field1[1]", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
		{
			TestName: "empty struct list Source and empty set Target",
			Source: testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{},
			},
			Target: &testFlexTFExpanderSetNestedObject{},
			WantTarget: &testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderStructSlice](), reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSExpanderStructSlice](), "Field1", reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]]()),
			},
		},
		{
			TestName: "non-empty struct list Source and set Target",
			Source: testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			Target: &testFlexTFExpanderSetNestedObject{},
			WantTarget: &testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderStructSlice](), reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSExpanderStructSlice](), "Field1", reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]]()),
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1[0]", reflect.TypeFor[string](), "Field1[0]", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1[1]", reflect.TypeFor[string](), "Field1[1]", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
		{
			TestName: "empty *struct list Source and empty set Target",
			Source: testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{},
			},
			Target: &testFlexTFExpanderSetNestedObject{},
			WantTarget: &testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderPtrSlice](), reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSExpanderPtrSlice](), "Field1", reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]*testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]]()),
			},
		},
		{
			TestName: "non-empty *struct list Source and non-empty set Target",
			Source: testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			Target: &testFlexTFExpanderSetNestedObject{},
			WantTarget: &testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderPtrSlice](), reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSExpanderPtrSlice](), "Field1", reflect.TypeFor[*testFlexTFExpanderSetNestedObject]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[[]*testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]]()),
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1[0]", reflect.TypeFor[string](), "Field1[0]", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1[1]", reflect.TypeFor[string](), "Field1[1]", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
		{
			TestName: "struct Source and object value Target",
			Source: testFlexAWSExpanderSingleStruct{
				Field1: testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			Target: &testFlexTFExpanderObjectValue{},
			WantTarget: &testFlexTFExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &testFlexTFFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderSingleStruct](), reflect.TypeFor[*testFlexTFExpanderObjectValue]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSExpanderSingleStruct](), "Field1", reflect.TypeFor[*testFlexTFExpanderObjectValue]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFFlexer]]()),
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
		{
			TestName: "*struct Source and object value Target",
			Source: testFlexAWSExpanderSinglePtr{
				Field1: &testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			Target: &testFlexTFExpanderObjectValue{},
			WantTarget: &testFlexTFExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &testFlexTFFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			expectedLogLines: []map[string]any{
				flatteningLogLine(reflect.TypeFor[testFlexAWSExpanderSinglePtr](), reflect.TypeFor[*testFlexTFExpanderObjectValue]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexAWSExpanderSinglePtr](), "Field1", reflect.TypeFor[*testFlexTFExpanderObjectValue]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[*testFlexAWSExpander](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFFlexer]]()),
				// StringValueToFramework in testFlexTFFlexer.Flatten()
				flatteningWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[*types.String]()),
				convertingWithPathLogLine("", reflect.TypeFor[string](), "", reflect.TypeFor[types.String]()), // TODO: fix path
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func runAutoFlattenTestCases(t *testing.T, testCases autoFlexTestCases) {
	t.Helper()

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if testCase.ContextFn != nil {
				ctx = testCase.ContextFn(ctx)
			}

			var buf bytes.Buffer
			ctx = tflogtest.RootLogger(ctx, &buf)

			ctx = registerTestingLogger(ctx)

			diags := Flatten(ctx, testCase.Source, testCase.Target, testCase.Options...)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			lines, err := tflogtest.MultilineJSONDecode(&buf)
			if err != nil {
				t.Fatalf("Expand: decoding log lines: %s", err)
			}
			if diff := cmp.Diff(lines, testCase.expectedLogLines); diff != "" {
				t.Errorf("unexpected log lines diff (+wanted, -got): %s", diff)
			}

			if !diags.HasError() {
				less := func(a, b any) bool { return fmt.Sprintf("%+v", a) < fmt.Sprintf("%+v", b) }
				if diff := cmp.Diff(testCase.Target, testCase.WantTarget, cmpopts.SortSlices(less)); diff != "" {
					if !testCase.WantDiff {
						t.Errorf("unexpected diff (+wanted, -got): %s", diff)
					}
				}
			}
		})
	}
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
		testCase := testCase
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
