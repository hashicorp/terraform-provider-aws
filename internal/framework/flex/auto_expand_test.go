// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	smithyjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestExpand(t *testing.T) {
	t.Parallel()

	testString := "test"
	testStringResult := "a"

	var (
		typedNilSource *TestFlex00
		typedNilTarget *TestFlex00
	)

	testARN := "arn:aws:securityhub:us-west-2:1234567890:control/cis-aws-foundations-benchmark/v/1.2.0/1.1" //lintignore:AWSAT003,AWSAT005

	testTimeStr := "2013-09-25T09:34:01Z"
	testTimeTime := errs.Must(time.Parse(time.RFC3339, testTimeStr))

	testCases := autoFlexTestCases{
		"nil Source": {
			Target: &TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Cannot expand nil source"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[<nil>, *flex.TestFlex00]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(nil, reflect.TypeFor[*TestFlex00]()),
			},
		},
		"typed nil Source": {
			Source: typedNilSource,
			Target: &TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Cannot expand nil source"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[*flex.TestFlex00, *flex.TestFlex00]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlex00](), reflect.TypeFor[*TestFlex00]()),
			},
		},
		"nil Target": {
			Source: TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Target cannot be nil"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.TestFlex00, <nil>]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[TestFlex00](), nil),
			},
		},
		"typed nil Target": {
			Source: TestFlex00{},
			Target: typedNilTarget,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Target cannot be nil"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.TestFlex00, *flex.TestFlex00]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[TestFlex00](), reflect.TypeFor[*TestFlex00]()),
			},
		},
		"non-pointer Target": {
			Source: TestFlex00{},
			Target: 0,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "target (int): int, want pointer"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.TestFlex00, int]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[TestFlex00](), reflect.TypeFor[int]()),
			},
		},
		"non-struct Source": {
			Source: testString,
			Target: &TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "does not implement attr.Value: string"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[string, *flex.TestFlex00]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[string](), reflect.TypeFor[*TestFlex00]()),
				infoConverting(reflect.TypeFor[string](), reflect.TypeFor[TestFlex00]()),
			},
		},
		"non-struct Target": {
			Source: TestFlex00{},
			Target: &testString,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "does not implement attr.Value: struct"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.TestFlex00, *string]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[TestFlex00](), reflect.TypeFor[*string]()),
				infoConverting(reflect.TypeFor[TestFlex00](), reflect.TypeFor[string]()),
			},
		},
		"types.String to string": {
			Source:     types.StringValue("a"),
			Target:     &testString,
			WantTarget: &testStringResult,
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.String](), reflect.TypeFor[*string]()),
				infoConverting(reflect.TypeFor[types.String](), reflect.TypeFor[string]()),
			},
		},
		"empty struct Source and Target": {
			Source:     TestFlex00{},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[TestFlex00](), reflect.TypeFor[*TestFlex00]()),
				infoConverting(reflect.TypeFor[TestFlex00](), reflect.TypeFor[TestFlex00]()),
			},
		},
		"empty struct pointer Source and Target": {
			Source:     &TestFlex00{},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlex00](), reflect.TypeFor[*TestFlex00]()),
				infoConverting(reflect.TypeFor[TestFlex00](), reflect.TypeFor[TestFlex00]()),
			},
		},
		"single string struct pointer Source and empty Target": {
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF01](), reflect.TypeFor[*TestFlex00]()),
				infoConverting(reflect.TypeFor[TestFlexTF01](), reflect.TypeFor[TestFlex00]()),
				debugNoCorrespondingField(reflect.TypeFor[*TestFlexTF01](), "Field1", reflect.TypeFor[*TestFlex00]()),
			},
		},
		"does not implement attr.Value Source": {
			Source: &TestFlexAWS01{Field1: "a"},
			Target: &TestFlexAWS01{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "does not implement attr.Value: string"),
				diag.NewErrorDiagnostic("AutoFlEx", "convert (Field1)"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[*flex.TestFlexAWS01, *flex.TestFlexAWS01]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexAWS01](), reflect.TypeFor[*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[TestFlexAWS01](), reflect.TypeFor[TestFlexAWS01]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[string]()),
			},
		},
		"single string Source and single string Target": {
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{Field1: "a"},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF01](), reflect.TypeFor[*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[TestFlexTF01](), reflect.TypeFor[TestFlexAWS01]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF01](), "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
			},
		},
		"single string Source and single *string Target": {
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlexAWS02{},
			WantTarget: &TestFlexAWS02{Field1: aws.String("a")},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF01](), reflect.TypeFor[*TestFlexAWS02]()),
				infoConverting(reflect.TypeFor[TestFlexTF01](), reflect.TypeFor[TestFlexAWS02]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF01](), "Field1", reflect.TypeFor[*TestFlexAWS02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
			},
		},
		"single string Source and single int64 Target": {
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlexAWS03{},
			WantTarget: &TestFlexAWS03{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF01](), reflect.TypeFor[*TestFlexAWS03]()),
				infoConverting(reflect.TypeFor[TestFlexTF01](), reflect.TypeFor[TestFlexAWS03]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF01](), "Field1", reflect.TypeFor[*TestFlexAWS03]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[int64]()),
				{
					"@level":             "info",
					"@module":            "provider.autoflex",
					"@message":           "AutoFlex Expand; incompatible types",
					"from":               map[string]any{},
					"to":                 float64(reflect.Int64),
					logAttrKeySourcePath: "Field1",
					logAttrKeySourceType: fullTypeName(reflect.TypeFor[types.String]()),
					logAttrKeyTargetPath: "Field1",
					logAttrKeyTargetType: fullTypeName(reflect.TypeFor[int64]()),
				},
			},
		},
		"primitive types Source and primtive types Target": {
			Source: &TestFlexTF03{
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
			Target: &TestFlexAWS04{},
			WantTarget: &TestFlexAWS04{
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF03](), reflect.TypeFor[*TestFlexAWS04]()),
				infoConverting(reflect.TypeFor[TestFlexTF03](), reflect.TypeFor[TestFlexAWS04]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF03](), "Field1", reflect.TypeFor[*TestFlexAWS04]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				traceMatchedFields("Field2", reflect.TypeFor[*TestFlexTF03](), "Field2", reflect.TypeFor[*TestFlexAWS04]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[types.String](), "Field2", reflect.TypeFor[*string]()),
				traceMatchedFields("Field3", reflect.TypeFor[*TestFlexTF03](), "Field3", reflect.TypeFor[*TestFlexAWS04]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[types.Int64](), "Field3", reflect.TypeFor[int32]()),
				traceMatchedFields("Field4", reflect.TypeFor[*TestFlexTF03](), "Field4", reflect.TypeFor[*TestFlexAWS04]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[types.Int64](), "Field4", reflect.TypeFor[*int32]()),
				traceMatchedFields("Field5", reflect.TypeFor[*TestFlexTF03](), "Field5", reflect.TypeFor[*TestFlexAWS04]()),
				infoConvertingWithPath("Field5", reflect.TypeFor[types.Int64](), "Field5", reflect.TypeFor[int64]()),
				traceMatchedFields("Field6", reflect.TypeFor[*TestFlexTF03](), "Field6", reflect.TypeFor[*TestFlexAWS04]()),
				infoConvertingWithPath("Field6", reflect.TypeFor[types.Int64](), "Field6", reflect.TypeFor[*int64]()),
				traceMatchedFields("Field7", reflect.TypeFor[*TestFlexTF03](), "Field7", reflect.TypeFor[*TestFlexAWS04]()),
				infoConvertingWithPath("Field7", reflect.TypeFor[types.Float64](), "Field7", reflect.TypeFor[float32]()),
				traceMatchedFields("Field8", reflect.TypeFor[*TestFlexTF03](), "Field8", reflect.TypeFor[*TestFlexAWS04]()),
				infoConvertingWithPath("Field8", reflect.TypeFor[types.Float64](), "Field8", reflect.TypeFor[*float32]()),
				traceMatchedFields("Field9", reflect.TypeFor[*TestFlexTF03](), "Field9", reflect.TypeFor[*TestFlexAWS04]()),
				infoConvertingWithPath("Field9", reflect.TypeFor[types.Float64](), "Field9", reflect.TypeFor[float64]()),
				traceMatchedFields("Field10", reflect.TypeFor[*TestFlexTF03](), "Field10", reflect.TypeFor[*TestFlexAWS04]()),
				infoConvertingWithPath("Field10", reflect.TypeFor[types.Float64](), "Field10", reflect.TypeFor[*float64]()),
				traceMatchedFields("Field11", reflect.TypeFor[*TestFlexTF03](), "Field11", reflect.TypeFor[*TestFlexAWS04]()),
				infoConvertingWithPath("Field11", reflect.TypeFor[types.Bool](), "Field11", reflect.TypeFor[bool]()),
				traceMatchedFields("Field12", reflect.TypeFor[*TestFlexTF03](), "Field12", reflect.TypeFor[*TestFlexAWS04]()),
				infoConvertingWithPath("Field12", reflect.TypeFor[types.Bool](), "Field12", reflect.TypeFor[*bool]()),
			},
		},
		"Collection of primitive types Source and slice or map of primtive types Target": {
			Source: &TestFlexTF04{
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
			Target: &TestFlexAWS05{},
			WantTarget: &TestFlexAWS05{
				Field1: []string{"a", "b"},
				Field2: aws.StringSlice([]string{"a", "b"}),
				Field3: []string{"a", "b"},
				Field4: aws.StringSlice([]string{"a", "b"}),
				Field5: map[string]string{"A": "a", "B": "b"},
				Field6: aws.StringMap(map[string]string{"A": "a", "B": "b"}),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF04](), reflect.TypeFor[*TestFlexAWS05]()),
				infoConverting(reflect.TypeFor[TestFlexTF04](), reflect.TypeFor[TestFlexAWS05]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF04](), "Field1", reflect.TypeFor[*TestFlexAWS05]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.List](), "Field1", reflect.TypeFor[[]string]()),
				traceExpandingWithElementsAs("Field1", reflect.TypeFor[types.List](), 2, "Field1", reflect.TypeFor[[]string]()),
				traceMatchedFields("Field2", reflect.TypeFor[*TestFlexTF04](), "Field2", reflect.TypeFor[*TestFlexAWS05]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[types.List](), "Field2", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Field2", reflect.TypeFor[types.List](), 2, "Field2", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Field3", reflect.TypeFor[*TestFlexTF04](), "Field3", reflect.TypeFor[*TestFlexAWS05]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[types.Set](), "Field3", reflect.TypeFor[[]string]()),
				traceExpandingWithElementsAs("Field3", reflect.TypeFor[types.Set](), 2, "Field3", reflect.TypeFor[[]string]()),
				traceMatchedFields("Field4", reflect.TypeFor[*TestFlexTF04](), "Field4", reflect.TypeFor[*TestFlexAWS05]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[types.Set](), "Field4", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Field4", reflect.TypeFor[types.Set](), 2, "Field4", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Field5", reflect.TypeFor[*TestFlexTF04](), "Field5", reflect.TypeFor[*TestFlexAWS05]()),
				infoConvertingWithPath("Field5", reflect.TypeFor[types.Map](), "Field5", reflect.TypeFor[map[string]string]()),
				traceExpandingWithElementsAs("Field5", reflect.TypeFor[types.Map](), 2, "Field5", reflect.TypeFor[map[string]string]()),
				traceMatchedFields("Field6", reflect.TypeFor[*TestFlexTF04](), "Field6", reflect.TypeFor[*TestFlexAWS05]()),
				infoConvertingWithPath("Field6", reflect.TypeFor[types.Map](), "Field6", reflect.TypeFor[map[string]*string]()),
				traceExpandingWithElementsAs("Field6", reflect.TypeFor[types.Map](), 2, "Field6", reflect.TypeFor[map[string]*string]()),
			},
		},
		"plural field names": {
			Source: &TestFlexTF09{
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
			Target: &TestFlexAWS11{},
			WantTarget: &TestFlexAWS11{
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF09](), reflect.TypeFor[*TestFlexAWS11]()),
				infoConverting(reflect.TypeFor[TestFlexTF09](), reflect.TypeFor[TestFlexAWS11]()),
				traceMatchedFields("City", reflect.TypeFor[*TestFlexTF09](), "Cities", reflect.TypeFor[*TestFlexAWS11]()),
				infoConvertingWithPath("City", reflect.TypeFor[types.List](), "Cities", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("City", reflect.TypeFor[types.List](), 2, "Cities", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Coach", reflect.TypeFor[*TestFlexTF09](), "Coaches", reflect.TypeFor[*TestFlexAWS11]()),
				infoConvertingWithPath("Coach", reflect.TypeFor[types.List](), "Coaches", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Coach", reflect.TypeFor[types.List](), 2, "Coaches", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Tomato", reflect.TypeFor[*TestFlexTF09](), "Tomatoes", reflect.TypeFor[*TestFlexAWS11]()),
				infoConvertingWithPath("Tomato", reflect.TypeFor[types.List](), "Tomatoes", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Tomato", reflect.TypeFor[types.List](), 2, "Tomatoes", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Vertex", reflect.TypeFor[*TestFlexTF09](), "Vertices", reflect.TypeFor[*TestFlexAWS11]()),
				infoConvertingWithPath("Vertex", reflect.TypeFor[types.List](), "Vertices", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Vertex", reflect.TypeFor[types.List](), 2, "Vertices", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Criterion", reflect.TypeFor[*TestFlexTF09](), "Criteria", reflect.TypeFor[*TestFlexAWS11]()),
				infoConvertingWithPath("Criterion", reflect.TypeFor[types.List](), "Criteria", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Criterion", reflect.TypeFor[types.List](), 2, "Criteria", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Datum", reflect.TypeFor[*TestFlexTF09](), "Data", reflect.TypeFor[*TestFlexAWS11]()),
				infoConvertingWithPath("Datum", reflect.TypeFor[types.List](), "Data", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Datum", reflect.TypeFor[types.List](), 2, "Data", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Hive", reflect.TypeFor[*TestFlexTF09](), "Hives", reflect.TypeFor[*TestFlexAWS11]()),
				infoConvertingWithPath("Hive", reflect.TypeFor[types.List](), "Hives", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Hive", reflect.TypeFor[types.List](), 2, "Hives", reflect.TypeFor[[]*string]()),
			},
		},
		"capitalization field names": {
			Source: &TestFlexTF10{
				FieldURL: types.StringValue("h"),
			},
			Target: &TestFlexAWS12{},
			WantTarget: &TestFlexAWS12{
				FieldUrl: aws.String("h"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF10](), reflect.TypeFor[*TestFlexAWS12]()),
				infoConverting(reflect.TypeFor[TestFlexTF10](), reflect.TypeFor[TestFlexAWS12]()),
				traceMatchedFields("FieldURL", reflect.TypeFor[*TestFlexTF10](), "FieldUrl", reflect.TypeFor[*TestFlexAWS12]()),
				infoConvertingWithPath("FieldURL", reflect.TypeFor[types.String](), "FieldUrl", reflect.TypeFor[*string]()),
			},
		},
		"resource name prefix": {
			Options: []AutoFlexOptionsFunc{WithFieldNamePrefix("Intent")},
			Source: &TestFlexTF16{
				Name: types.StringValue("Ovodoghen"),
			},
			Target: &TestFlexAWS18{},
			WantTarget: &TestFlexAWS18{
				IntentName: aws.String("Ovodoghen"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF16](), reflect.TypeFor[*TestFlexAWS18]()),
				infoConverting(reflect.TypeFor[TestFlexTF16](), reflect.TypeFor[TestFlexAWS18]()),
				traceMatchedFields("Name", reflect.TypeFor[*TestFlexTF16](), "IntentName", reflect.TypeFor[*TestFlexAWS18]()),
				infoConvertingWithPath("Name", reflect.TypeFor[types.String](), "IntentName", reflect.TypeFor[*string]()),
			},
		},
		"single ARN Source and single string Target": {
			Source:     &TestFlexTF17{Field1: fwtypes.ARNValue(testARN)},
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{Field1: testARN},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF17](), reflect.TypeFor[*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[TestFlexTF17](), reflect.TypeFor[TestFlexAWS01]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF17](), "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ARN](), "Field1", reflect.TypeFor[string]()),
			},
		},
		"single ARN Source and single *string Target": {
			Source:     &TestFlexTF17{Field1: fwtypes.ARNValue(testARN)},
			Target:     &TestFlexAWS02{},
			WantTarget: &TestFlexAWS02{Field1: aws.String(testARN)},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF17](), reflect.TypeFor[*TestFlexAWS02]()),
				infoConverting(reflect.TypeFor[TestFlexTF17](), reflect.TypeFor[TestFlexAWS02]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF17](), "Field1", reflect.TypeFor[*TestFlexAWS02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ARN](), "Field1", reflect.TypeFor[*string]()),
			},
		},
		"timestamp pointer": {
			Source: &TestFlexTimeTF01{
				CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
			},
			Target: &TestFlexTimeAWS01{},
			WantTarget: &TestFlexTimeAWS01{
				CreationDateTime: &testTimeTime,
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTimeTF01](), reflect.TypeFor[*TestFlexTimeAWS01]()),
				infoConverting(reflect.TypeFor[TestFlexTimeTF01](), reflect.TypeFor[TestFlexTimeAWS01]()),
				traceMatchedFields("CreationDateTime", reflect.TypeFor[*TestFlexTimeTF01](), "CreationDateTime", reflect.TypeFor[*TestFlexTimeAWS01]()),
				infoConvertingWithPath("CreationDateTime", reflect.TypeFor[timetypes.RFC3339](), "CreationDateTime", reflect.TypeFor[*time.Time]()),
			},
		},
		"timestamp": {
			Source: &TestFlexTimeTF01{
				CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
			},
			Target: &TestFlexTimeAWS02{},
			WantTarget: &TestFlexTimeAWS02{
				CreationDateTime: testTimeTime,
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTimeTF01](), reflect.TypeFor[*TestFlexTimeAWS02]()),
				infoConverting(reflect.TypeFor[TestFlexTimeTF01](), reflect.TypeFor[TestFlexTimeAWS02]()),
				traceMatchedFields("CreationDateTime", reflect.TypeFor[*TestFlexTimeTF01](), "CreationDateTime", reflect.TypeFor[*TestFlexTimeAWS02]()),
				infoConvertingWithPath("CreationDateTime", reflect.TypeFor[timetypes.RFC3339](), "CreationDateTime", reflect.TypeFor[time.Time]()),
			},
		},
		"JSONValue Source to json interface Target": {
			Source: &TestFlexTF20{Field1: fwtypes.SmithyJSONValue(`{"field1": "a"}`, newTestJSONDocument)},
			Target: &TestFlexAWS19{},
			WantTarget: &TestFlexAWS19{
				Field1: &testJSONDocument{
					Value: map[string]any{
						"field1": "a",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF20](), reflect.TypeFor[*TestFlexAWS19]()),
				infoConverting(reflect.TypeFor[TestFlexTF20](), reflect.TypeFor[TestFlexAWS19]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF20](), "Field1", reflect.TypeFor[*TestFlexAWS19]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SmithyJSON[smithyjson.JSONStringer]](), "Field1", reflect.TypeFor[smithyjson.JSONStringer]()),
			},
		},
	}

	runAutoExpandTestCases(t, testCases)
}

func TestExpandGeneric(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"single list Source and *struct Target": {
			Source: &TestFlexTF05{
				Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &TestFlexTF01{
					Field1: types.StringValue("a"),
				}),
			},
			Target: &TestFlexAWS06{},
			WantTarget: &TestFlexAWS06{
				Field1: &TestFlexAWS01{
					Field1: "a",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF05](), reflect.TypeFor[*TestFlexAWS06]()),
				infoConverting(reflect.TypeFor[TestFlexTF05](), reflect.TypeFor[TestFlexAWS06]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF05](), "Field1", reflect.TypeFor[*TestFlexAWS06]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1.Field1", reflect.TypeFor[string]()),
			},
		},
		"single set Source and *struct Target": {
			Source: &TestFlexTF06{
				Field1: fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &TestFlexTF01{
					Field1: types.StringValue("a"),
				}),
			},
			Target: &TestFlexAWS06{},
			WantTarget: &TestFlexAWS06{
				Field1: &TestFlexAWS01{
					Field1: "a",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF06](), reflect.TypeFor[*TestFlexAWS06]()),
				infoConverting(reflect.TypeFor[TestFlexTF06](), reflect.TypeFor[TestFlexAWS06]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF06](), "Field1", reflect.TypeFor[*TestFlexAWS06]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1.Field1", reflect.TypeFor[string]()),
			},
		},
		"empty list Source and empty []struct Target": {
			Source:     &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{})},
			Target:     &TestFlexAWS08{},
			WantTarget: &TestFlexAWS08{Field1: []TestFlexAWS01{}},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF05](), reflect.TypeFor[*TestFlexAWS08]()),
				infoConverting(reflect.TypeFor[TestFlexTF05](), reflect.TypeFor[TestFlexAWS08]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF05](), "Field1", reflect.TypeFor[*TestFlexAWS08]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), 0, "Field1", reflect.TypeFor[[]TestFlexAWS01]()),
			},
		},
		"non-empty list Source and non-empty []struct Target": {
			Source: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
			Target: &TestFlexAWS08{},
			WantTarget: &TestFlexAWS08{Field1: []TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF05](), reflect.TypeFor[*TestFlexAWS08]()),
				infoConverting(reflect.TypeFor[TestFlexTF05](), reflect.TypeFor[TestFlexAWS08]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF05](), "Field1", reflect.TypeFor[*TestFlexAWS08]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), 2, "Field1", reflect.TypeFor[[]TestFlexAWS01]()),
				traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1[0].Field1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[types.String](), "Field1[1].Field1", reflect.TypeFor[string]()),
			},
		},
		"empty list Source and empty []*struct Target": {
			Source:     &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*TestFlexTF01{})},
			Target:     &TestFlexAWS07{},
			WantTarget: &TestFlexAWS07{Field1: []*TestFlexAWS01{}},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF05](), reflect.TypeFor[*TestFlexAWS07]()),
				infoConverting(reflect.TypeFor[TestFlexTF05](), reflect.TypeFor[TestFlexAWS07]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF05](), "Field1", reflect.TypeFor[*TestFlexAWS07]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]*TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), 0, "Field1", reflect.TypeFor[[]*TestFlexAWS01]()),
			},
		},
		"non-empty list Source and non-empty []*struct Target": {
			Source: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
			Target: &TestFlexAWS07{},
			WantTarget: &TestFlexAWS07{Field1: []*TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF05](), reflect.TypeFor[*TestFlexAWS07]()),
				infoConverting(reflect.TypeFor[TestFlexTF05](), reflect.TypeFor[TestFlexAWS07]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF05](), "Field1", reflect.TypeFor[*TestFlexAWS07]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]*TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), 2, "Field1", reflect.TypeFor[[]*TestFlexAWS01]()),
				traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1[0].Field1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[types.String](), "Field1[1].Field1", reflect.TypeFor[string]()),
			},
		},
		"empty set Source and empty []*struct Target": {
			Source:     &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*TestFlexTF01{})},
			Target:     &TestFlexAWS07{},
			WantTarget: &TestFlexAWS07{Field1: []*TestFlexAWS01{}},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF06](), reflect.TypeFor[*TestFlexAWS07]()),
				infoConverting(reflect.TypeFor[TestFlexTF06](), reflect.TypeFor[TestFlexAWS07]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF06](), "Field1", reflect.TypeFor[*TestFlexAWS07]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]*TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), 0, "Field1", reflect.TypeFor[[]*TestFlexAWS01]()),
			},
		},
		"non-empty set Source and non-empty []*struct Target": {
			Source: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
			Target: &TestFlexAWS07{},
			WantTarget: &TestFlexAWS07{Field1: []*TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF06](), reflect.TypeFor[*TestFlexAWS07]()),
				infoConverting(reflect.TypeFor[TestFlexTF06](), reflect.TypeFor[TestFlexAWS07]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF06](), "Field1", reflect.TypeFor[*TestFlexAWS07]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]*TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), 2, "Field1", reflect.TypeFor[[]*TestFlexAWS01]()),
				traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1[0].Field1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[types.String](), "Field1[1].Field1", reflect.TypeFor[string]()),
			},
		},
		"non-empty set Source and non-empty []struct Target": {
			Source: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
			Target: &TestFlexAWS08{},
			WantTarget: &TestFlexAWS08{Field1: []TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF06](), reflect.TypeFor[*TestFlexAWS08]()),
				infoConverting(reflect.TypeFor[TestFlexTF06](), reflect.TypeFor[TestFlexAWS08]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF06](), "Field1", reflect.TypeFor[*TestFlexAWS08]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), 2, "Field1", reflect.TypeFor[[]TestFlexAWS01]()),
				traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1[0].Field1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[types.String](), "Field1[1].Field1", reflect.TypeFor[string]()),
			},
		},
		"complex Source and complex Target": {
			Source: &TestFlexTF07{
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
			Target: &TestFlexAWS09{},
			WantTarget: &TestFlexAWS09{
				Field1: "m",
				Field2: &TestFlexAWS06{
					Field1: &TestFlexAWS01{
						Field1: "n",
					},
				},
				Field3: aws.StringMap(map[string]string{
					"X": "x",
					"Y": "y",
				}),
				Field4: []TestFlexAWS03{
					{Field1: 100},
					{Field1: 2000},
					{Field1: 30000},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF07](), reflect.TypeFor[*TestFlexAWS09]()),
				infoConverting(reflect.TypeFor[TestFlexTF07](), reflect.TypeFor[TestFlexAWS09]()),

				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF07](), "Field1", reflect.TypeFor[*TestFlexAWS09]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				traceMatchedFields("Field2", reflect.TypeFor[*TestFlexTF07](), "Field2", reflect.TypeFor[*TestFlexAWS09]()),

				infoConvertingWithPath("Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF05]](), "Field2", reflect.TypeFor[*TestFlexAWS06]()),
				traceMatchedFieldsWithPath("Field2[0]", "Field1", reflect.TypeFor[*TestFlexTF05](), "Field2", "Field1", reflect.TypeFor[*TestFlexAWS06]()),
				infoConvertingWithPath("Field2[0].Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field2.Field1", reflect.TypeFor[*TestFlexAWS01]()),
				traceMatchedFieldsWithPath("Field2[0].Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field2.Field1", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("Field2[0].Field1[0].Field1", reflect.TypeFor[types.String](), "Field2.Field1.Field1", reflect.TypeFor[string]()),

				traceMatchedFields("Field3", reflect.TypeFor[*TestFlexTF07](), "Field3", reflect.TypeFor[*TestFlexAWS09]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[types.Map](), "Field3", reflect.TypeFor[map[string]*string]()),
				traceExpandingWithElementsAs("Field3", reflect.TypeFor[types.Map](), 2, "Field3", reflect.TypeFor[map[string]*string]()),

				traceMatchedFields("Field4", reflect.TypeFor[*TestFlexTF07](), "Field4", reflect.TypeFor[*TestFlexAWS09]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF02]](), "Field4", reflect.TypeFor[[]TestFlexAWS03]()),
				traceExpandingNestedObjectCollection("Field4", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF02]](), 3, "Field4", reflect.TypeFor[[]TestFlexAWS03]()),
				traceMatchedFieldsWithPath("Field4[0]", "Field1", reflect.TypeFor[*TestFlexTF02](), "Field4[0]", "Field1", reflect.TypeFor[*TestFlexAWS03]()),
				infoConvertingWithPath("Field4[0].Field1", reflect.TypeFor[types.Int64](), "Field4[0].Field1", reflect.TypeFor[int64]()),
				traceMatchedFieldsWithPath("Field4[1]", "Field1", reflect.TypeFor[*TestFlexTF02](), "Field4[1]", "Field1", reflect.TypeFor[*TestFlexAWS03]()),
				infoConvertingWithPath("Field4[1].Field1", reflect.TypeFor[types.Int64](), "Field4[1].Field1", reflect.TypeFor[int64]()),
				traceMatchedFieldsWithPath("Field4[2]", "Field1", reflect.TypeFor[*TestFlexTF02](), "Field4[2]", "Field1", reflect.TypeFor[*TestFlexAWS03]()),
				infoConvertingWithPath("Field4[2].Field1", reflect.TypeFor[types.Int64](), "Field4[2].Field1", reflect.TypeFor[int64]()),
			},
		},
		"map of string": {
			Source: &TestFlexTF11{
				FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			Target: &TestFlexAWS13{},
			WantTarget: &TestFlexAWS13{
				FieldInner: map[string]string{
					"x": "y",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF11](), reflect.TypeFor[*TestFlexAWS13]()),
				infoConverting(reflect.TypeFor[TestFlexTF11](), reflect.TypeFor[TestFlexAWS13]()),
				traceMatchedFields("FieldInner", reflect.TypeFor[*TestFlexTF11](), "FieldInner", reflect.TypeFor[*TestFlexAWS13]()),
				infoConvertingWithPath("FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), "FieldInner", reflect.TypeFor[map[string]string]()),
				traceExpandingWithElementsAs("FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), 1, "FieldInner", reflect.TypeFor[map[string]string]()),
			},
		},
		"map of string pointer": {
			Source: &TestFlexTF11{
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF11](), reflect.TypeFor[*awsMapOfStringPointer]()),
				infoConverting(reflect.TypeFor[TestFlexTF11](), reflect.TypeFor[awsMapOfStringPointer]()),
				traceMatchedFields("FieldInner", reflect.TypeFor[*TestFlexTF11](), "FieldInner", reflect.TypeFor[*awsMapOfStringPointer]()),
				infoConvertingWithPath("FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), "FieldInner", reflect.TypeFor[map[string]*string]()),
				traceExpandingWithElementsAs("FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), 1, "FieldInner", reflect.TypeFor[map[string]*string]()),
			},
		},
		"map of map of string": {
			Source: &TestFlexTF21{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			Target: &TestFlexAWS21{},
			WantTarget: &TestFlexAWS21{
				Field1: map[string]map[string]string{
					"x": {
						"y": "z",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF21](), reflect.TypeFor[*TestFlexAWS21]()),
				infoConverting(reflect.TypeFor[TestFlexTF21](), reflect.TypeFor[TestFlexAWS21]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF21](), "Field1", reflect.TypeFor[*TestFlexAWS21]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]](), "Field1", reflect.TypeFor[map[string]map[string]string]()),
				traceExpandingWithElementsAs("Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]](), 1, "Field1", reflect.TypeFor[map[string]map[string]string]()),
			},
		},
		"map of map of string pointer": {
			Source: &TestFlexTF21{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			Target: &TestFlexAWS22{},
			WantTarget: &TestFlexAWS22{
				Field1: map[string]map[string]*string{
					"x": {
						"y": aws.String("z"),
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF21](), reflect.TypeFor[*TestFlexAWS22]()),
				infoConverting(reflect.TypeFor[TestFlexTF21](), reflect.TypeFor[TestFlexAWS22]()),
				traceMatchedFields("Field1", reflect.TypeFor[*TestFlexTF21](), "Field1", reflect.TypeFor[*TestFlexAWS22]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]](), "Field1", reflect.TypeFor[map[string]map[string]*string]()),
				traceExpandingWithElementsAs("Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]](), 1, "Field1", reflect.TypeFor[map[string]map[string]*string]()),
			},
		},
		"nested string map": {
			Source: &TestFlexTF14{
				FieldOuter: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &TestFlexTF11{
					FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"x": types.StringValue("y"),
					}),
				}),
			},
			Target: &TestFlexAWS16{},
			WantTarget: &TestFlexAWS16{
				FieldOuter: TestFlexAWS13{
					FieldInner: map[string]string{
						"x": "y",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexTF14](), reflect.TypeFor[*TestFlexAWS16]()),
				infoConverting(reflect.TypeFor[TestFlexTF14](), reflect.TypeFor[TestFlexAWS16]()),
				traceMatchedFields("FieldOuter", reflect.TypeFor[*TestFlexTF14](), "FieldOuter", reflect.TypeFor[*TestFlexAWS16]()),
				infoConvertingWithPath("FieldOuter", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF11]](), "FieldOuter", reflect.TypeFor[TestFlexAWS13]()),
				traceMatchedFieldsWithPath("FieldOuter[0]", "FieldInner", reflect.TypeFor[*TestFlexTF11](), "FieldOuter", "FieldInner", reflect.TypeFor[*TestFlexAWS13]()),
				infoConvertingWithPath("FieldOuter[0].FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), "FieldOuter.FieldInner", reflect.TypeFor[map[string]string]()),
				traceExpandingWithElementsAs("FieldOuter[0].FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), 1, "FieldOuter.FieldInner", reflect.TypeFor[map[string]string]()),
			},
		},
		"null map block key": {
			Source: &TestFlexMapBlockKeyTF01{
				MapBlock: fwtypes.NewListNestedObjectValueOfNull[TestFlexMapBlockKeyTF02](ctx),
			},
			Target: &TestFlexMapBlockKeyAWS01{},
			WantTarget: &TestFlexMapBlockKeyAWS01{
				MapBlock: nil,
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexMapBlockKeyTF01](), reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				infoConverting(reflect.TypeFor[TestFlexMapBlockKeyTF01](), reflect.TypeFor[TestFlexMapBlockKeyAWS01]()),
				traceMatchedFields("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF01](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF02]](), "MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02]()),
				traceExpandingNullValue("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF02]](), "MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02]()),
			},
		},
		"map block key list": {
			Source: &TestFlexMapBlockKeyTF01{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust[TestFlexMapBlockKeyTF02](ctx, []TestFlexMapBlockKeyTF02{
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
			Target: &TestFlexMapBlockKeyAWS01{},
			WantTarget: &TestFlexMapBlockKeyAWS01{
				MapBlock: map[string]TestFlexMapBlockKeyAWS02{
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexMapBlockKeyTF01](), reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				infoConverting(reflect.TypeFor[TestFlexMapBlockKeyTF01](), reflect.TypeFor[TestFlexMapBlockKeyAWS01]()),

				traceMatchedFields("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF01](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF02]](), "MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02]()),

				traceSkipMapBlockKey("MapBlock[0]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"x\"]", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"x\"]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"x\"]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr2", reflect.TypeFor[string]()),

				traceSkipMapBlockKey("MapBlock[1]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"y\"]", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"y\"]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"y\"]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr2", reflect.TypeFor[string]()),
			},
		},
		"map block key set": {
			Source: &TestFlexMapBlockKeyTF03{
				MapBlock: fwtypes.NewSetNestedObjectValueOfValueSliceMust[TestFlexMapBlockKeyTF02](ctx, []TestFlexMapBlockKeyTF02{
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
			Target: &TestFlexMapBlockKeyAWS01{},
			WantTarget: &TestFlexMapBlockKeyAWS01{
				MapBlock: map[string]TestFlexMapBlockKeyAWS02{
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexMapBlockKeyTF03](), reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				infoConverting(reflect.TypeFor[TestFlexMapBlockKeyTF03](), reflect.TypeFor[TestFlexMapBlockKeyAWS01]()),

				traceMatchedFields("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF03](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexMapBlockKeyTF02]](), "MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02]()),

				traceSkipMapBlockKey("MapBlock[0]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"x\"]", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"x\"]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"x\"]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr2", reflect.TypeFor[string]()),

				traceSkipMapBlockKey("MapBlock[1]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"y\"]", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"y\"]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"y\"]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr2", reflect.TypeFor[string]()),
			},
		},
		"map block key ptr source": {
			Source: &TestFlexMapBlockKeyTF01{
				MapBlock: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*TestFlexMapBlockKeyTF02{
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
			Target: &TestFlexMapBlockKeyAWS01{},
			WantTarget: &TestFlexMapBlockKeyAWS01{
				MapBlock: map[string]TestFlexMapBlockKeyAWS02{
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexMapBlockKeyTF01](), reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				infoConverting(reflect.TypeFor[TestFlexMapBlockKeyTF01](), reflect.TypeFor[TestFlexMapBlockKeyAWS01]()),

				traceMatchedFields("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF01](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF02]](), "MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02]()),

				traceSkipMapBlockKey("MapBlock[0]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"x\"]", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"x\"]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"x\"]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr2", reflect.TypeFor[string]()),

				traceSkipMapBlockKey("MapBlock[1]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"y\"]", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"y\"]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"y\"]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr2", reflect.TypeFor[string]()),
			},
		},
		"map block key ptr both": {
			Source: &TestFlexMapBlockKeyTF01{
				MapBlock: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*TestFlexMapBlockKeyTF02{
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
			Target: &TestFlexMapBlockKeyAWS03{},
			WantTarget: &TestFlexMapBlockKeyAWS03{
				MapBlock: map[string]*TestFlexMapBlockKeyAWS02{
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexMapBlockKeyTF01](), reflect.TypeFor[*TestFlexMapBlockKeyAWS03]()),
				infoConverting(reflect.TypeFor[TestFlexMapBlockKeyTF01](), reflect.TypeFor[TestFlexMapBlockKeyAWS03]()),

				traceMatchedFields("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF01](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS03]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF02]](), "MapBlock", reflect.TypeFor[map[string]*TestFlexMapBlockKeyAWS02]()),

				traceSkipMapBlockKey("MapBlock[0]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"x\"]", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"x\"]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"x\"]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr2", reflect.TypeFor[string]()),

				traceSkipMapBlockKey("MapBlock[1]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"y\"]", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"y\"]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock[\"y\"]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr2", reflect.TypeFor[string]()),
			},
		},
		"map block enum key": {
			Source: &TestFlexMapBlockKeyTF04{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust[TestFlexMapBlockKeyTF05](ctx, []TestFlexMapBlockKeyTF05{
					{
						MapBlockKey: fwtypes.StringEnumValue(TestEnumList),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
					{
						MapBlockKey: fwtypes.StringEnumValue(TestEnumScalar),
						Attr1:       types.StringValue("c"),
						Attr2:       types.StringValue("d"),
					},
				}),
			},
			Target: &TestFlexMapBlockKeyAWS01{},
			WantTarget: &TestFlexMapBlockKeyAWS01{
				MapBlock: map[string]TestFlexMapBlockKeyAWS02{
					string(TestEnumList): {
						Attr1: "a",
						Attr2: "b",
					},
					string(TestEnumScalar): {
						Attr1: "c",
						Attr2: "d",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*TestFlexMapBlockKeyTF04](), reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				infoConverting(reflect.TypeFor[TestFlexMapBlockKeyTF04](), reflect.TypeFor[TestFlexMapBlockKeyAWS01]()),

				traceMatchedFields("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF04](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF05]](), "MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02]()),

				traceSkipMapBlockKey("MapBlock[0]", reflect.TypeFor[*TestFlexMapBlockKeyTF05](), "MapBlock[\"List\"]", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF05](), "MapBlock[\"List\"]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"List\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF05](), "MapBlock[\"List\"]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"List\"].Attr2", reflect.TypeFor[string]()),

				traceSkipMapBlockKey("MapBlock[1]", reflect.TypeFor[*TestFlexMapBlockKeyTF05](), "MapBlock[\"Scalar\"]", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF05](), "MapBlock[\"Scalar\"]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"Scalar\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF05](), "MapBlock[\"Scalar\"]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				infoConvertingWithPath("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"Scalar\"].Attr2", reflect.TypeFor[string]()),
			},
		},
	}

	runAutoExpandTestCases(t, testCases)
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
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{Field1: types.StringValue("a"), Field2: types.Int64Value(1)})},
			Target:     &aws02{},
			WantTarget: &aws02{Field1: &aws01{Field1: aws.String("a"), Field2: 1}},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws02]()),
				infoConverting(reflect.TypeFor[tf02](), reflect.TypeFor[aws02]()),
				traceMatchedFields("Field1", reflect.TypeFor[*tf02](), "Field1", reflect.TypeFor[*aws02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1", reflect.TypeFor[*aws01]()),
				traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[*tf01](), "Field1", "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1.Field1", reflect.TypeFor[types.String](), "Field1.Field1", reflect.TypeFor[*string]()),
				traceMatchedFieldsWithPath("Field1", "Field2", reflect.TypeFor[*tf01](), "Field1", "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1.Field2", reflect.TypeFor[types.Int64](), "Field1.Field2", reflect.TypeFor[int64]()),
			},
		},
		"single nested block nil": {
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfNull[tf01](ctx)},
			Target:     &aws02{},
			WantTarget: &aws02{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws02]()),
				infoConverting(reflect.TypeFor[tf02](), reflect.TypeFor[aws02]()),
				traceMatchedFields("Field1", reflect.TypeFor[*tf02](), "Field1", reflect.TypeFor[*aws02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1", reflect.TypeFor[*aws01]()),
				traceExpandingNullValue("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1", reflect.TypeFor[*aws01]()),
			},
		},
		"single nested block value": {
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{Field1: types.StringValue("a"), Field2: types.Int64Value(1)})},
			Target:     &aws03{},
			WantTarget: &aws03{Field1: aws01{Field1: aws.String("a"), Field2: 1}},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws03]()),
				infoConverting(reflect.TypeFor[tf02](), reflect.TypeFor[aws03]()),
				traceMatchedFields("Field1", reflect.TypeFor[*tf02](), "Field1", reflect.TypeFor[*aws03]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1", reflect.TypeFor[aws01]()),
				traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[*tf01](), "Field1", "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1.Field1", reflect.TypeFor[types.String](), "Field1.Field1", reflect.TypeFor[*string]()),
				traceMatchedFieldsWithPath("Field1", "Field2", reflect.TypeFor[*tf01](), "Field1", "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1.Field2", reflect.TypeFor[types.Int64](), "Field1.Field2", reflect.TypeFor[int64]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
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
				Field1: fwtypes.NewObjectValueOfMust[tf02](
					ctx,
					&tf02{
						Field1: fwtypes.NewObjectValueOfMust[tf01](
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf03](), reflect.TypeFor[*aws03]()),
				infoConverting(reflect.TypeFor[tf03](), reflect.TypeFor[aws03]()),
				traceMatchedFields("Field1", reflect.TypeFor[*tf03](), "Field1", reflect.TypeFor[*aws03]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf02]](), "Field1", reflect.TypeFor[*aws02]()),
				traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[*tf02](), "Field1", "Field1", reflect.TypeFor[*aws02]()),
				infoConvertingWithPath("Field1.Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1.Field1", reflect.TypeFor[*aws01]()),
				traceMatchedFieldsWithPath("Field1.Field1", "Field1", reflect.TypeFor[*tf01](), "Field1.Field1", "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1.Field1.Field1", reflect.TypeFor[types.Bool](), "Field1.Field1.Field1", reflect.TypeFor[bool]()),
				traceMatchedFieldsWithPath("Field1.Field1", "Field2", reflect.TypeFor[*tf01](), "Field1.Field1", "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1.Field1.Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]](), "Field1.Field1.Field2", reflect.TypeFor[[]string]()),
				traceExpandingWithElementsAs("Field1.Field1.Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]](), 2, "Field1.Field1.Field2", reflect.TypeFor[[]string]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandStringEnum(t *testing.T) {
	t.Parallel()

	var testEnum TestEnum
	testEnumList := TestEnumList

	testCases := autoFlexTestCases{
		"valid value": {
			Source:     fwtypes.StringEnumValue(TestEnumList),
			Target:     &testEnum,
			WantTarget: &testEnumList,
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), reflect.TypeFor[*TestEnum]()),
				infoConverting(reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), reflect.TypeFor[TestEnum]()),
			},
		},
		"empty value": {
			Source:     fwtypes.StringEnumNull[TestEnum](),
			Target:     &testEnum,
			WantTarget: &testEnum,
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), reflect.TypeFor[*TestEnum]()),
				infoConverting(reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), reflect.TypeFor[TestEnum]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), "", reflect.TypeFor[TestEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]int64]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 2, "", reflect.TypeFor[[]int64]()),
			},
		},
		"empty value []int64": {
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]int64]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 0, "", reflect.TypeFor[[]int64]()),
			},
		},
		"null value []int64": {
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]int64]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]int64]()),
				traceExpandingNullValue("", reflect.TypeFor[types.List](), "", reflect.TypeFor[[]int64]()),
			},
		},
		"valid value []*int64": {
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{aws.Int64(1), aws.Int64(-1)},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]*int64]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]*int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 2, "", reflect.TypeFor[[]*int64]()),
			},
		},
		"empty value []*int64": {
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]*int64]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]*int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 0, "", reflect.TypeFor[[]*int64]()),
			},
		},
		"null value []*int64": {
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]*int64]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]*int64]()),
				traceExpandingNullValue("", reflect.TypeFor[types.List](), "", reflect.TypeFor[[]*int64]()),
			},
		},
		"valid value []int32": {
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int32{},
			WantTarget: &[]int32{1, -1},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]int32]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 2, "", reflect.TypeFor[[]int32]()),
			},
		},
		"empty value []int32": {
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]int32]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 0, "", reflect.TypeFor[[]int32]()),
			},
		},
		"null value []int32": {
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]int32]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]int32]()),
				traceExpandingNullValue("", reflect.TypeFor[types.List](), "", reflect.TypeFor[[]int32]()),
			},
		},
		"valid value []*int32": {
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{aws.Int32(1), aws.Int32(-1)},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]*int32]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]*int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 2, "", reflect.TypeFor[[]*int32]()),
			},
		},
		"empty value []*int32": {
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]*int32]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]*int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 0, "", reflect.TypeFor[[]*int32]()),
			},
		},
		"null value []*int32": {
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]*int32]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]*int32]()),
				traceExpandingNullValue("", reflect.TypeFor[types.List](), "", reflect.TypeFor[[]*int32]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]int64]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 2, "", reflect.TypeFor[[]int64]()),
			},
		},
		"empty value []int64": {
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]int64]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 0, "", reflect.TypeFor[[]int64]()),
			},
		},
		"null value []int64": {
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]int64]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]int64]()),
				traceExpandingNullValue("", reflect.TypeFor[types.Set](), "", reflect.TypeFor[[]int64]()),
			},
		},
		"valid value []*int64": {
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{aws.Int64(1), aws.Int64(-1)},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]*int64]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]*int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 2, "", reflect.TypeFor[[]*int64]()),
			},
		},
		"empty value []*int64": {
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]*int64]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]*int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 0, "", reflect.TypeFor[[]*int64]()),
			},
		},
		"null value []*int64": {
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]*int64]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]*int64]()),
				traceExpandingNullValue("", reflect.TypeFor[types.Set](), "", reflect.TypeFor[[]*int64]()),
			},
		},
		"valid value []int32": {
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int32{},
			WantTarget: &[]int32{1, -1},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]int32]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 2, "", reflect.TypeFor[[]int32]()),
			},
		},
		"empty value []int32": {
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]int32]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 0, "", reflect.TypeFor[[]int32]()),
			},
		},
		"null value []int32": {
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]int32]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]int32]()),
				traceExpandingNullValue("", reflect.TypeFor[types.Set](), "", reflect.TypeFor[[]int32]()),
			},
		},
		"valid value []*int32": {
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{aws.Int32(1), aws.Int32(-1)},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]*int32]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]*int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 2, "", reflect.TypeFor[[]*int32]()),
			},
		},
		"empty value []*int32": {
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]*int32]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]*int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 0, "", reflect.TypeFor[[]*int32]()),
			},
		},
		"null value []*int32": {
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]*int32]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]*int32]()),
				traceExpandingNullValue("", reflect.TypeFor[types.Set](), "", reflect.TypeFor[[]*int32]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandListOfStringEnum(t *testing.T) {
	t.Parallel()

	type testEnum string
	var testEnumFoo testEnum = "foo"
	var testEnumBar testEnum = "bar"

	testCases := autoFlexTestCases{
		"valid value": {
			Source: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue(string(testEnumFoo)),
				types.StringValue(string(testEnumBar)),
			}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{testEnumFoo, testEnumBar},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]testEnum]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]testEnum]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 2, "", reflect.TypeFor[[]testEnum]()),
			},
		},
		"empty value": {
			Source:     types.ListValueMust(types.StringType, []attr.Value{}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]testEnum]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]testEnum]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 0, "", reflect.TypeFor[[]testEnum]()),
			},
		},
		"null value": {
			Source:     types.ListNull(types.StringType),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]testEnum]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]testEnum]()),
				traceExpandingNullValue("", reflect.TypeFor[types.List](), "", reflect.TypeFor[[]testEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandSetOfStringEnum(t *testing.T) {
	t.Parallel()

	type testEnum string
	var testEnumFoo testEnum = "foo"
	var testEnumBar testEnum = "bar"

	testCases := autoFlexTestCases{
		"valid value": {
			Source: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue(string(testEnumFoo)),
				types.StringValue(string(testEnumBar)),
			}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{testEnumFoo, testEnumBar},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]testEnum]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]testEnum]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 2, "", reflect.TypeFor[[]testEnum]()),
			},
		},
		"empty value": {
			Source:     types.SetValueMust(types.StringType, []attr.Value{}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]testEnum]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]testEnum]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 0, "", reflect.TypeFor[[]testEnum]()),
			},
		},
		"null value": {
			Source:     types.SetNull(types.StringType),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]testEnum]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]testEnum]()),
				traceExpandingNullValue("", reflect.TypeFor[types.Set](), "", reflect.TypeFor[[]testEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandListOfNestedObject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid value to []struct": {
			Source: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			Target: &[]TestFlexAWS01{},
			WantTarget: &[]TestFlexAWS01{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), 2, "", reflect.TypeFor[[]TestFlexAWS01]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[types.String](), "[0].Field1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("[1].Field1", reflect.TypeFor[types.String](), "[1].Field1", reflect.TypeFor[string]()),
			},
		},
		"empty value to []struct": {
			Source:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{}),
			Target:     &[]TestFlexAWS01{},
			WantTarget: &[]TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), 0, "", reflect.TypeFor[[]TestFlexAWS01]()),
			},
		},
		"null value to []struct": {
			Source:     fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx),
			Target:     &[]TestFlexAWS01{},
			WantTarget: &[]TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]TestFlexAWS01]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "", reflect.TypeFor[[]TestFlexAWS01]()),
			},
		},

		"valid value to []*struct": {
			Source: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			Target: &[]*TestFlexAWS01{},
			WantTarget: &[]*TestFlexAWS01{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]*TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), 2, "", reflect.TypeFor[[]*TestFlexAWS01]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[types.String](), "[0].Field1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("[1].Field1", reflect.TypeFor[types.String](), "[1].Field1", reflect.TypeFor[string]()),
			},
		},
		"empty value to []*struct": {
			Source:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{}),
			Target:     &[]*TestFlexAWS01{},
			WantTarget: &[]*TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]*TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), 0, "", reflect.TypeFor[[]*TestFlexAWS01]()),
			},
		},
		"null value to []*struct": {
			Source:     fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx),
			Target:     &[]*TestFlexAWS01{},
			WantTarget: &[]*TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]*TestFlexAWS01]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "", reflect.TypeFor[[]*TestFlexAWS01]()),
			},
		},

		"single list value to single struct": {
			Source: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{
				{
					Field1: types.StringValue("value1"),
				},
			}),
			Target: &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{
				Field1: "value1",
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[TestFlexAWS01]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
			},
		},
		"empty list value to single struct": {
			Source:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{}),
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[TestFlexAWS01]()),
			},
		},
		"null value to single struct": {
			Source:     fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx),
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[TestFlexAWS01]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "", reflect.TypeFor[TestFlexAWS01]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandSetOfNestedObject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid value to []struct": {
			Source: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			Target: &[]TestFlexAWS01{},
			WantTarget: &[]TestFlexAWS01{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), 2, "", reflect.TypeFor[[]TestFlexAWS01]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[types.String](), "[0].Field1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("[1].Field1", reflect.TypeFor[types.String](), "[1].Field1", reflect.TypeFor[string]()),
			},
		},
		"empty value to []struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{}),
			Target:     &[]TestFlexAWS01{},
			WantTarget: &[]TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), 0, "", reflect.TypeFor[[]TestFlexAWS01]()),
			},
		},
		"null value to []struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfNull[TestFlexTF01](ctx),
			Target:     &[]TestFlexAWS01{},
			WantTarget: &[]TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]TestFlexAWS01]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), "", reflect.TypeFor[[]TestFlexAWS01]()),
			},
		},

		"valid value to []*struct": {
			Source: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			Target: &[]*TestFlexAWS01{},
			WantTarget: &[]*TestFlexAWS01{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]*TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), 2, "", reflect.TypeFor[[]*TestFlexAWS01]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[types.String](), "[0].Field1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("[1].Field1", reflect.TypeFor[types.String](), "[1].Field1", reflect.TypeFor[string]()),
			},
		},
		"empty value to []*struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{}),
			Target:     &[]*TestFlexAWS01{},
			WantTarget: &[]*TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]*TestFlexAWS01]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), 0, "", reflect.TypeFor[[]*TestFlexAWS01]()),
			},
		},
		"null value to []*struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfNull[TestFlexTF01](ctx),
			Target:     &[]*TestFlexAWS01{},
			WantTarget: &[]*TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]*TestFlexAWS01]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), "", reflect.TypeFor[[]*TestFlexAWS01]()),
			},
		},

		"single set value to single struct": {
			Source: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{
				{
					Field1: types.StringValue("value1"),
				},
			}),
			Target: &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{
				Field1: "value1",
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[TestFlexAWS01]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
			},
		},
		"empty set value to single struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{}),
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[TestFlexAWS01]()),
			},
		},
		"null value to single struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfNull[TestFlexTF01](ctx),
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*TestFlexAWS01]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[TestFlexAWS01]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), "", reflect.TypeFor[TestFlexAWS01]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandSimpleNestedBlockWithStringEnum(t *testing.T) {
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
		"single nested valid value": {
			Source: &tf01{
				Field1: types.Int64Value(1),
				Field2: fwtypes.StringEnumValue(TestEnumList),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: 1,
				Field2: TestEnumList,
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf01](), reflect.TypeFor[aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[*tf01](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[*tf01](), "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), "Field2", reflect.TypeFor[TestEnum]()),
			},
		},
		"single nested null value": {
			Source: &tf01{
				Field1: types.Int64Value(1),
				Field2: fwtypes.StringEnumNull[TestEnum](),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: 1,
				Field2: "",
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf01](), reflect.TypeFor[aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[*tf01](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[*tf01](), "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), "Field2", reflect.TypeFor[TestEnum]()),
				traceExpandingNullValue("Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), "Field2", reflect.TypeFor[TestEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandComplexNestedBlockWithStringEnum(t *testing.T) {
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
	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: fwtypes.StringEnumValue(TestEnumList),
				}),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field2: TestEnumList,
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf02](), reflect.TypeFor[aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[*tf02](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[*tf02](), "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]](), "Field2", reflect.TypeFor[*aws02]()),
				traceMatchedFieldsWithPath("Field2[0]", "Field2", reflect.TypeFor[*tf01](), "Field2", "Field2", reflect.TypeFor[*aws02]()),
				infoConvertingWithPath("Field2[0].Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), "Field2.Field2", reflect.TypeFor[TestEnum]()),
			},
		},
		"single nested null value": {
			Source: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: fwtypes.StringEnumNull[TestEnum](),
				}),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field2: "",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf02](), reflect.TypeFor[aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[*tf02](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[*tf02](), "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]](), "Field2", reflect.TypeFor[*aws02]()),
				traceMatchedFieldsWithPath("Field2[0]", "Field2", reflect.TypeFor[*tf01](), "Field2", "Field2", reflect.TypeFor[*aws02]()),
				infoConvertingWithPath("Field2[0].Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), "Field2.Field2", reflect.TypeFor[TestEnum]()),
				traceExpandingNullValue("Field2[0].Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), "Field2.Field2", reflect.TypeFor[TestEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandOptions(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Bool                       `tfsdk:"field1"`
		Tags   fwtypes.MapValueOf[types.String] `tfsdk:"tags"`
	}
	type aws01 struct {
		Field1 bool
		Tags   map[string]string
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"empty source with tags": {
			Source:     &tf01{},
			Target:     &aws01{},
			WantTarget: &aws01{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf01](), reflect.TypeFor[aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[*tf01](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				traceExpandingNullValue("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				traceSkipIgnoredField(reflect.TypeFor[*tf01](), "Tags", reflect.TypeFor[*aws01]()),
			},
		},
		"ignore tags by default": {
			Source: &tf01{
				Field1: types.BoolValue(true),
				Tags: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"foo": types.StringValue("bar"),
				},
				),
			},
			Target:     &aws01{},
			WantTarget: &aws01{Field1: true},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf01](), reflect.TypeFor[aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[*tf01](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				traceSkipIgnoredField(reflect.TypeFor[*tf01](), "Tags", reflect.TypeFor[*aws01]()),
			},
		},
		"include tags with option override": {
			Options: []AutoFlexOptionsFunc{WithNoIgnoredFieldNames()},
			Source: &tf01{
				Field1: types.BoolValue(true),
				Tags: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"foo": types.StringValue("bar"),
				},
				),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: true,
				Tags:   map[string]string{"foo": "bar"},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf01](), reflect.TypeFor[aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[*tf01](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				traceMatchedFields("Tags", reflect.TypeFor[*tf01](), "Tags", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), "Tags", reflect.TypeFor[map[string]string]()),
				traceExpandingWithElementsAs("Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), 1, "Tags", reflect.TypeFor[map[string]string]()),
			},
		},
		"ignore custom field": {
			Options: []AutoFlexOptionsFunc{WithIgnoredFieldNames([]string{"Field1"})},
			Source: &tf01{
				Field1: types.BoolValue(true),
				Tags: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"foo": types.StringValue("bar"),
				},
				),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Tags: map[string]string{"foo": "bar"},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf01](), reflect.TypeFor[aws01]()),
				traceSkipIgnoredField(reflect.TypeFor[*tf01](), "Field1", reflect.TypeFor[*aws01]()),
				traceMatchedFields("Tags", reflect.TypeFor[*tf01](), "Tags", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), "Tags", reflect.TypeFor[map[string]string]()),
				traceExpandingWithElementsAs("Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), 1, "Tags", reflect.TypeFor[map[string]string]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandInterface(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var targetInterface testFlexAWSInterfaceInterface

	testCases := autoFlexTestCases{
		"top level": {
			Source: testFlexTFInterfaceFlexer{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			WantTarget: testFlexAWSInterfaceInterfacePtr(&testFlexAWSInterfaceInterfaceImpl{
				AWSField: "value1",
			}),
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFInterfaceFlexer](), reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				infoConverting(reflect.TypeFor[testFlexTFInterfaceFlexer](), reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[testFlexTFInterfaceFlexer](), "", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceFlexer.Expand()
				infoExpandingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[*string]()), // TODO: fix path
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"top level return value does not implement target interface": {
			Source: testFlexTFInterfaceIncompatibleExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			expectedDiags: diag.Diagnostics{
				diagExpandedTypeDoesNotImplement(reflect.TypeFor[*testFlexAWSInterfaceIncompatibleImpl](), reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFInterfaceIncompatibleExpander, *flex.testFlexAWSInterfaceInterface]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFInterfaceIncompatibleExpander](), reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				infoConverting(reflect.TypeFor[testFlexTFInterfaceIncompatibleExpander](), reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[testFlexTFInterfaceIncompatibleExpander](), "", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceFlexer.Expand()
				infoExpandingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[*string]()), // TODO: fix path
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"single list Source and single interface Target": {
			Source: testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSInterfaceSingle{},
			WantTarget: &testFlexAWSInterfaceSingle{
				Field1: &testFlexAWSInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConverting(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[testFlexAWSInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[*testFlexTFInterfaceFlexer](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceFlexer.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()), // TODO: fix path
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()),                // TODO: fix path
			},
		},
		"single list non-Expander Source and single interface Target": {
			Source: testFlexTFListNestedObject[TestFlexTF01]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSInterfaceSingle{},
			WantTarget: &testFlexAWSInterfaceSingle{
				Field1: nil,
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFListNestedObject[TestFlexTF01]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConverting(reflect.TypeFor[testFlexTFListNestedObject[TestFlexTF01]](), reflect.TypeFor[testFlexAWSInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFListNestedObject[TestFlexTF01]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				{
					"@level":             "info",
					"@module":            "provider.autoflex",
					"@message":           "AutoFlex Expand; incompatible types",
					"from":               map[string]any{},
					"to":                 float64(reflect.Interface),
					logAttrKeySourcePath: "Field1[0]",
					logAttrKeySourceType: fullTypeName(reflect.TypeFor[*TestFlexTF01]()),
					logAttrKeyTargetPath: "Field1",
					logAttrKeyTargetType: fullTypeName(reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				},
			},
		},
		"single set Source and single interface Target": {
			Source: testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSInterfaceSingle{},
			WantTarget: &testFlexAWSInterfaceSingle{
				Field1: &testFlexAWSInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConverting(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[testFlexAWSInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[*testFlexTFInterfaceFlexer](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceFlexer.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"empty list Source and empty interface Target": {
			Source: testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[testFlexAWSInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]](), 0, "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
			},
		},
		"non-empty list Source and non-empty interface Target": {
			Source: testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{
					&testFlexAWSInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&testFlexAWSInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[testFlexAWSInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]](), 2, "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[*testFlexTFInterfaceFlexer](), "Field1[0]", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceFlexer.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1[0]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
				infoSourceImplementsFlexExpander("Field1[1]", reflect.TypeFor[*testFlexTFInterfaceFlexer](), "Field1[1]", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceFlexer.Expand()
				infoExpandingWithPath("Field1[1]", reflect.TypeFor[types.String](), "Field1[1]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"empty set Source and empty interface Target": {
			Source: testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[testFlexAWSInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]](), 0, "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
			},
		},
		"non-empty set Source and non-empty interface Target": {
			Source: testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{
					&testFlexAWSInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&testFlexAWSInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[testFlexAWSInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]](), 2, "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[*testFlexTFInterfaceFlexer](), "Field1[0]", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceFlexer.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1[0]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
				infoSourceImplementsFlexExpander("Field1[1]", reflect.TypeFor[*testFlexTFInterfaceFlexer](), "Field1[1]", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceFlexer.Expand()
				infoExpandingWithPath("Field1[1]", reflect.TypeFor[types.String](), "Field1[1]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"object value Source and struct Target": {
			Source: testFlexTFObjectValue[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &testFlexTFInterfaceFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &testFlexAWSInterfaceSingle{},
			WantTarget: &testFlexAWSInterfaceSingle{
				Field1: &testFlexAWSInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFObjectValue[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConverting(reflect.TypeFor[testFlexTFObjectValue[testFlexTFInterfaceFlexer]](), reflect.TypeFor[testFlexAWSInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFObjectValue[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexExpander("Field1", reflect.TypeFor[*testFlexTFInterfaceFlexer](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceFlexer.Expand()
				infoExpandingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func testFlexAWSInterfaceInterfacePtr(v testFlexAWSInterfaceInterface) *testFlexAWSInterfaceInterface { // nosemgrep:ci.aws-in-func-name
	return &v
}

func TestExpandExpander(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"top level struct Target": {
			Source: testFlexTFFlexer{
				Field1: types.StringValue("value1"),
			},
			Target: &testFlexAWSExpander{},
			WantTarget: &testFlexAWSExpander{
				AWSField: "value1",
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFFlexer](), reflect.TypeFor[*testFlexAWSExpander]()),
				infoConverting(reflect.TypeFor[testFlexTFFlexer](), reflect.TypeFor[testFlexAWSExpander]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[testFlexTFFlexer](), "", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"top level string Target": {
			Source: testFlexTFExpanderToString{
				Field1: types.StringValue("value1"),
			},
			Target:     aws.String(""),
			WantTarget: aws.String("value1"),
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderToString](), reflect.TypeFor[*string]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderToString](), reflect.TypeFor[string]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[testFlexTFExpanderToString](), "", reflect.TypeFor[string]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"top level incompatible struct Target": {
			Source: testFlexTFFlexer{
				Field1: types.StringValue("value1"),
			},
			Target: &testFlexAWSExpanderIncompatible{},
			expectedDiags: diag.Diagnostics{
				diagCannotBeAssigned(reflect.TypeFor[testFlexAWSExpander](), reflect.TypeFor[testFlexAWSExpanderIncompatible]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFFlexer, *flex.testFlexAWSExpanderIncompatible]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFFlexer](), reflect.TypeFor[*testFlexAWSExpanderIncompatible]()),
				infoConverting(reflect.TypeFor[testFlexTFFlexer](), reflect.TypeFor[testFlexAWSExpanderIncompatible]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[testFlexTFFlexer](), "", reflect.TypeFor[*testFlexAWSExpanderIncompatible]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[*string]()), // TODO: fix path
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"top level expands to nil": {
			Source: testFlexTFExpanderToNil{
				Field1: types.StringValue("value1"),
			},
			Target: &testFlexAWSExpander{},
			expectedDiags: diag.Diagnostics{
				diagExpandsToNil(reflect.TypeFor[testFlexTFExpanderToNil]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFExpanderToNil, *flex.testFlexAWSExpander]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderToNil](), reflect.TypeFor[*testFlexAWSExpander]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderToNil](), reflect.TypeFor[testFlexAWSExpander]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[testFlexTFExpanderToNil](), "", reflect.TypeFor[*testFlexAWSExpander]()),
			},
		},
		"top level incompatible non-struct Target": {
			Source: testFlexTFExpanderToString{
				Field1: types.StringValue("value1"),
			},
			Target: aws.Int64(0),
			expectedDiags: diag.Diagnostics{
				diagCannotBeAssigned(reflect.TypeFor[string](), reflect.TypeFor[int64]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFExpanderToString, *int64]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderToString](), reflect.TypeFor[*int64]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderToString](), reflect.TypeFor[int64]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[testFlexTFExpanderToString](), "", reflect.TypeFor[int64]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"single list Source and single struct Target": {
			Source: testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSExpanderSingleStruct{},
			WantTarget: &testFlexAWSExpanderSingleStruct{
				Field1: testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[testFlexAWSExpanderSingleStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[testFlexAWSExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[*testFlexTFFlexer](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"single set Source and single struct Target": {
			Source: testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSExpanderSingleStruct{},
			WantTarget: &testFlexAWSExpanderSingleStruct{
				Field1: testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[testFlexAWSExpanderSingleStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[testFlexAWSExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[*testFlexTFFlexer](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"single list Source and single *struct Target": {
			Source: testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSExpanderSinglePtr{},
			WantTarget: &testFlexAWSExpanderSinglePtr{
				Field1: &testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[testFlexAWSExpanderSinglePtr]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[*testFlexTFFlexer](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"single set Source and single *struct Target": {
			Source: testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSExpanderSinglePtr{},
			WantTarget: &testFlexAWSExpanderSinglePtr{
				Field1: &testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[testFlexAWSExpanderSinglePtr]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[*testFlexTFFlexer](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"empty list Source and empty struct Target": {
			Source: testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{}),
			},
			Target: &testFlexAWSExpanderStructSlice{},
			WantTarget: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[testFlexAWSExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), 0, "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
			},
		},
		"non-empty list Source and non-empty struct Target": {
			Source: testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &testFlexAWSExpanderStructSlice{},
			WantTarget: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[testFlexAWSExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), 2, "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[*testFlexTFFlexer](), "Field1[0]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1[0]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
				infoSourceImplementsFlexExpander("Field1[1]", reflect.TypeFor[*testFlexTFFlexer](), "Field1[1]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1[1]", reflect.TypeFor[types.String](), "Field1[1]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"empty list Source and empty *struct Target": {
			Source: testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{}),
			},
			Target: &testFlexAWSExpanderPtrSlice{},
			WantTarget: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[testFlexAWSExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), 0, "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
			},
		},
		"non-empty list Source and non-empty *struct Target": {
			Source: testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &testFlexAWSExpanderPtrSlice{},
			WantTarget: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[testFlexAWSExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), 2, "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[*testFlexTFFlexer](), "Field1[0]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1[0]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
				infoSourceImplementsFlexExpander("Field1[1]", reflect.TypeFor[*testFlexTFFlexer](), "Field1[1]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1[1]", reflect.TypeFor[types.String](), "Field1[1]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"empty set Source and empty struct Target": {
			Source: testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{}),
			},
			Target: &testFlexAWSExpanderStructSlice{},
			WantTarget: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[testFlexAWSExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), 0, "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
			},
		},
		"non-empty set Source and non-empty struct Target": {
			Source: testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &testFlexAWSExpanderStructSlice{},
			WantTarget: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[testFlexAWSExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), 2, "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[*testFlexTFFlexer](), "Field1[0]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1[0]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
				infoSourceImplementsFlexExpander("Field1[1]", reflect.TypeFor[*testFlexTFFlexer](), "Field1[1]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1[1]", reflect.TypeFor[types.String](), "Field1[1]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"empty set Source and empty *struct Target": {
			Source: testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{}),
			},
			Target: &testFlexAWSExpanderPtrSlice{},
			WantTarget: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[testFlexAWSExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), 0, "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
			},
		},
		"non-empty set Source and non-empty *struct Target": {
			Source: testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &testFlexAWSExpanderPtrSlice{},
			WantTarget: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[testFlexAWSExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), 2, "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[*testFlexTFFlexer](), "Field1[0]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1[0]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
				infoSourceImplementsFlexExpander("Field1[1]", reflect.TypeFor[*testFlexTFFlexer](), "Field1[1]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1[1]", reflect.TypeFor[types.String](), "Field1[1]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"object value Source and struct Target": {
			Source: testFlexTFExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &testFlexTFFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &testFlexAWSExpanderSingleStruct{},
			WantTarget: &testFlexAWSExpanderSingleStruct{
				Field1: testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderObjectValue](), reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderObjectValue](), reflect.TypeFor[testFlexAWSExpanderSingleStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderObjectValue](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[testFlexAWSExpander]()),
				infoSourceImplementsFlexExpander("Field1", reflect.TypeFor[*testFlexTFFlexer](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"object value Source and *struct Target": {
			Source: testFlexTFExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &testFlexTFFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &testFlexAWSExpanderSinglePtr{},
			WantTarget: &testFlexAWSExpanderSinglePtr{
				Field1: &testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFExpanderObjectValue](), reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				infoConverting(reflect.TypeFor[testFlexTFExpanderObjectValue](), reflect.TypeFor[testFlexAWSExpanderSinglePtr]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFExpanderObjectValue](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				infoSourceImplementsFlexExpander("Field1", reflect.TypeFor[*testFlexTFFlexer](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFFlexer.Expand()
				infoExpandingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandInterfaceTypedExpander(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var targetInterface testFlexAWSInterfaceInterface

	testCases := autoFlexTestCases{
		"top level": {
			Source: testFlexTFInterfaceTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			WantTarget: testFlexAWSInterfaceInterfacePtr(&testFlexAWSInterfaceInterfaceImpl{
				AWSField: "value1",
			}),
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFInterfaceTypedExpander](), reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				infoConverting(reflect.TypeFor[testFlexTFInterfaceTypedExpander](), reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("", reflect.TypeFor[testFlexTFInterfaceTypedExpander](), "", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceTypedExpander.Expand()
				infoExpandingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[*string]()), // TODO: fix path
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"top level return value does not implement target interface": {
			Source: testFlexTFInterfaceIncompatibleTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			expectedDiags: diag.Diagnostics{
				diagExpandedTypeDoesNotImplement(reflect.TypeFor[*testFlexAWSInterfaceIncompatibleImpl](), reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFInterfaceIncompatibleTypedExpander, *flex.testFlexAWSInterfaceInterface]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFInterfaceIncompatibleTypedExpander](), reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				infoConverting(reflect.TypeFor[testFlexTFInterfaceIncompatibleTypedExpander](), reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("", reflect.TypeFor[testFlexTFInterfaceIncompatibleTypedExpander](), "", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceTypedExpander.Expand()
				infoExpandingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[*string]()), // TODO: fix path
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"single list Source and single interface Target": {
			Source: testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSInterfaceSingle{},
			WantTarget: &testFlexAWSInterfaceSingle{
				Field1: &testFlexAWSInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConverting(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[testFlexAWSInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[*testFlexTFInterfaceTypedExpander](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceTypedExpander.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()), // TODO: fix path
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()),                // TODO: fix path
			},
		},
		"single list non-Expander Source and single interface Target": {
			Source: testFlexTFListNestedObject[TestFlexTF01]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSInterfaceSingle{},
			WantTarget: &testFlexAWSInterfaceSingle{
				Field1: nil,
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFListNestedObject[TestFlexTF01]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConverting(reflect.TypeFor[testFlexTFListNestedObject[TestFlexTF01]](), reflect.TypeFor[testFlexAWSInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFListNestedObject[TestFlexTF01]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				{
					"@level":             "info",
					"@module":            "provider.autoflex",
					"@message":           "AutoFlex Expand; incompatible types",
					"from":               map[string]any{},
					"to":                 float64(reflect.Interface),
					logAttrKeySourcePath: "Field1[0]",
					logAttrKeySourceType: fullTypeName(reflect.TypeFor[*TestFlexTF01]()),
					logAttrKeyTargetPath: "Field1",
					logAttrKeyTargetType: fullTypeName(reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				},
			},
		},
		"single set Source and single interface Target": {
			Source: testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSInterfaceSingle{},
			WantTarget: &testFlexAWSInterfaceSingle{
				Field1: &testFlexAWSInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConverting(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[testFlexAWSInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[*testFlexTFInterfaceTypedExpander](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceTypedExpander.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()), // TODO: fix path
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()),                // TODO: fix path
			},
		},
		"empty list Source and empty interface Target": {
			Source: testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceTypedExpander{}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[testFlexAWSInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), 0, "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
			},
		},
		"non-empty list Source and non-empty interface Target": {
			Source: testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{
					&testFlexAWSInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&testFlexAWSInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[testFlexAWSInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), 2, "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[*testFlexTFInterfaceTypedExpander](), "Field1[0]", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceTypedExpander.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1[0]", reflect.TypeFor[*string]()), // TODO: fix path
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()),                   // TODO: fix path
				infoSourceImplementsFlexTypedExpander("Field1[1]", reflect.TypeFor[*testFlexTFInterfaceTypedExpander](), "Field1[1]", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceTypedExpander.Expand()
				infoExpandingWithPath("Field1[1]", reflect.TypeFor[types.String](), "Field1[1]", reflect.TypeFor[*string]()), // TODO: fix path
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()),                   // TODO: fix path
			},
		},
		"empty set Source and empty interface Target": {
			Source: testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceTypedExpander{}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[testFlexAWSInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), 0, "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
			},
		},
		"non-empty set Source and non-empty interface Target": {
			Source: testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{
					&testFlexAWSInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&testFlexAWSInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[testFlexAWSInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), 2, "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[*testFlexTFInterfaceTypedExpander](), "Field1[0]", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceTypedExpander.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1[0]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
				infoSourceImplementsFlexTypedExpander("Field1[1]", reflect.TypeFor[*testFlexTFInterfaceTypedExpander](), "Field1[1]", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceTypedExpander.Expand()
				infoExpandingWithPath("Field1[1]", reflect.TypeFor[types.String](), "Field1[1]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"object value Source and struct Target": {
			Source: testFlexTFObjectValue[testFlexTFInterfaceTypedExpander]{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &testFlexTFInterfaceTypedExpander{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &testFlexAWSInterfaceSingle{},
			WantTarget: &testFlexAWSInterfaceSingle{
				Field1: &testFlexAWSInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFObjectValue[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConverting(reflect.TypeFor[testFlexTFObjectValue[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[testFlexAWSInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFObjectValue[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("Field1", reflect.TypeFor[*testFlexTFInterfaceTypedExpander](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				// StringValueFromFramework in testFlexTFInterfaceTypedExpander.Expand()
				infoExpandingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandTypedExpander(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"top level struct Target": {
			Source: testFlexTFTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &testFlexAWSExpander{},
			WantTarget: &testFlexAWSExpander{
				AWSField: "value1",
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpander](), reflect.TypeFor[*testFlexAWSExpander]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpander](), reflect.TypeFor[testFlexAWSExpander]()),
				infoSourceImplementsFlexTypedExpander("", reflect.TypeFor[testFlexTFTypedExpander](), "", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"top level incompatible struct Target": {
			Source: testFlexTFTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &testFlexAWSExpanderIncompatible{},
			expectedDiags: diag.Diagnostics{
				diagCannotBeAssigned(reflect.TypeFor[testFlexAWSExpander](), reflect.TypeFor[testFlexAWSExpanderIncompatible]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFTypedExpander, *flex.testFlexAWSExpanderIncompatible]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpander](), reflect.TypeFor[*testFlexAWSExpanderIncompatible]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpander](), reflect.TypeFor[testFlexAWSExpanderIncompatible]()),
				infoSourceImplementsFlexTypedExpander("", reflect.TypeFor[testFlexTFTypedExpander](), "", reflect.TypeFor[*testFlexAWSExpanderIncompatible]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"top level expands to nil": {
			Source: testFlexTFTypedExpanderToNil{
				Field1: types.StringValue("value1"),
			},
			Target: &testFlexAWSExpander{},
			expectedDiags: diag.Diagnostics{
				diagExpandsToNil(reflect.TypeFor[testFlexTFTypedExpanderToNil]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFTypedExpanderToNil, *flex.testFlexAWSExpander]"),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderToNil](), reflect.TypeFor[*testFlexAWSExpander]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderToNil](), reflect.TypeFor[testFlexAWSExpander]()),
				infoSourceImplementsFlexTypedExpander("", reflect.TypeFor[testFlexTFTypedExpanderToNil](), "", reflect.TypeFor[*testFlexAWSExpander]()),
			},
		},
		"single list Source and single struct Target": {
			Source: testFlexTFTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSExpanderSingleStruct{},
			WantTarget: &testFlexAWSExpanderSingleStruct{
				Field1: testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[testFlexAWSExpanderSingleStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[testFlexAWSExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"single set Source and single struct Target": {
			Source: testFlexTFSetNestedObject[testFlexTFTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSExpanderSingleStruct{},
			WantTarget: &testFlexAWSExpanderSingleStruct{
				Field1: testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFTypedExpander]](), reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				infoConverting(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFTypedExpander]](), reflect.TypeFor[testFlexAWSExpanderSingleStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[testFlexAWSExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"single list Source and single *struct Target": {
			Source: testFlexTFTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSExpanderSinglePtr{},
			WantTarget: &testFlexAWSExpanderSinglePtr{
				Field1: &testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[testFlexAWSExpanderSinglePtr]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"single set Source and single *struct Target": {
			Source: testFlexTFTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &testFlexAWSExpanderSinglePtr{},
			WantTarget: &testFlexAWSExpanderSinglePtr{
				Field1: &testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[testFlexAWSExpanderSinglePtr]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"empty list Source and empty struct Target": {
			Source: testFlexTFTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{}),
			},
			Target: &testFlexAWSExpanderStructSlice{},
			WantTarget: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[testFlexAWSExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), 0, "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
			},
		},
		"non-empty list Source and non-empty struct Target": {
			Source: testFlexTFTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &testFlexAWSExpanderStructSlice{},
			WantTarget: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[testFlexAWSExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), 2, "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1[0]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1[0]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
				infoSourceImplementsFlexTypedExpander("Field1[1]", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1[1]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1[1]", reflect.TypeFor[types.String](), "Field1[1]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"empty list Source and empty *struct Target": {
			Source: testFlexTFTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{}),
			},
			Target: &testFlexAWSExpanderPtrSlice{},
			WantTarget: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[testFlexAWSExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), 0, "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
			},
		},
		"non-empty list Source and non-empty *struct Target": {
			Source: testFlexTFTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &testFlexAWSExpanderPtrSlice{},
			WantTarget: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[testFlexAWSExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), 2, "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1[0]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1[0]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
				infoSourceImplementsFlexTypedExpander("Field1[1]", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1[1]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1[1]", reflect.TypeFor[types.String](), "Field1[1]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"empty set Source and empty struct Target": {
			Source: testFlexTFTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{}),
			},
			Target: &testFlexAWSExpanderStructSlice{},
			WantTarget: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[testFlexAWSExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), 0, "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
			},
		},
		"non-empty set Source and non-empty struct Target": {
			Source: testFlexTFTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &testFlexAWSExpanderStructSlice{},
			WantTarget: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[testFlexAWSExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), 2, "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1[0]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1[0]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
				infoSourceImplementsFlexTypedExpander("Field1[1]", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1[1]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1[1]", reflect.TypeFor[types.String](), "Field1[1]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"empty set Source and empty *struct Target": {
			Source: testFlexTFTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{}),
			},
			Target: &testFlexAWSExpanderPtrSlice{},
			WantTarget: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[testFlexAWSExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), 0, "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
			},
		},
		"non-empty set Source and non-empty *struct Target": {
			Source: testFlexTFTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &testFlexAWSExpanderPtrSlice{},
			WantTarget: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[testFlexAWSExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), 2, "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1[0]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1[0]", reflect.TypeFor[types.String](), "Field1[0]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
				infoSourceImplementsFlexTypedExpander("Field1[1]", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1[1]", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1[1]", reflect.TypeFor[types.String](), "Field1[1]", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"object value Source and struct Target": {
			Source: testFlexTFTypedExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &testFlexTFTypedExpander{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &testFlexAWSExpanderSingleStruct{},
			WantTarget: &testFlexAWSExpanderSingleStruct{
				Field1: testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderObjectValue](), reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderObjectValue](), reflect.TypeFor[testFlexAWSExpanderSingleStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderObjectValue](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[testFlexAWSExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
		"object value Source and *struct Target": {
			Source: testFlexTFTypedExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &testFlexTFTypedExpander{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &testFlexAWSExpanderSinglePtr{},
			WantTarget: &testFlexAWSExpanderSinglePtr{
				Field1: &testFlexAWSExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[testFlexTFTypedExpanderObjectValue](), reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				infoConverting(reflect.TypeFor[testFlexTFTypedExpanderObjectValue](), reflect.TypeFor[testFlexAWSExpanderSinglePtr]()),
				traceMatchedFields("Field1", reflect.TypeFor[testFlexTFTypedExpanderObjectValue](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1", reflect.TypeFor[*testFlexTFTypedExpander](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
				// StringValueFromFramework in testFlexTFTypedExpander.Expand()
				infoExpandingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				infoConvertingWithPath("", reflect.TypeFor[types.String](), "", reflect.TypeFor[string]()), // TODO: fix path
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

type autoFlexTestCase struct {
	ContextFn func(context.Context) context.Context
	Options   []AutoFlexOptionsFunc
	// TestName         string
	Source           any
	Target           any
	expectedDiags    diag.Diagnostics
	expectedLogLines []map[string]any
	WantTarget       any
	WantDiff         bool
}

type autoFlexTestCases map[string]autoFlexTestCase

func runAutoExpandTestCases(t *testing.T, testCases autoFlexTestCases) {
	t.Helper()

	for testName, testCase := range testCases {
		testCase := testCase
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			if testCase.ContextFn != nil {
				ctx = testCase.ContextFn(ctx)
			}

			var buf bytes.Buffer
			ctx = tflogtest.RootLogger(ctx, &buf)

			ctx = registerTestingLogger(ctx)

			diags := Expand(ctx, testCase.Source, testCase.Target, testCase.Options...)

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
				if diff := cmp.Diff(testCase.Target, testCase.WantTarget); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}
