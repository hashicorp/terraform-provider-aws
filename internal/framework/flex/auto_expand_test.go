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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
		{
			TestName: "nil Source",
			Target:   &TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Cannot expand nil source"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[<nil>, *flex.TestFlex00]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(nil, reflect.TypeFor[*TestFlex00]()),
			},
		},
		{
			TestName: "typed nil Source",
			Source:   typedNilSource,
			Target:   &TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Cannot expand nil source"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[*flex.TestFlex00, *flex.TestFlex00]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlex00](), reflect.TypeFor[*TestFlex00]()),
			},
		},
		{
			TestName: "nil Target",
			Source:   TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Target cannot be nil"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.TestFlex00, <nil>]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[TestFlex00](), nil),
			},
		},
		{
			TestName: "typed nil Target",
			Source:   TestFlex00{},
			Target:   typedNilTarget,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Target cannot be nil"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.TestFlex00, *flex.TestFlex00]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[TestFlex00](), reflect.TypeFor[*TestFlex00]()),
			},
		},
		{
			TestName: "non-pointer Target",
			Source:   TestFlex00{},
			Target:   0,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "target (int): int, want pointer"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.TestFlex00, int]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[TestFlex00](), reflect.TypeFor[int]()),
			},
		},
		{
			TestName: "non-struct Source",
			Source:   testString,
			Target:   &TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "does not implement attr.Value: string"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[string, *flex.TestFlex00]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[string](), reflect.TypeFor[*TestFlex00]()),
				convertingLogLine(reflect.TypeFor[string](), reflect.TypeFor[TestFlex00]()),
			},
		},
		{
			TestName: "non-struct Target",
			Source:   TestFlex00{},
			Target:   &testString,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "does not implement attr.Value: struct"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.TestFlex00, *string]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[TestFlex00](), reflect.TypeFor[*string]()),
				convertingLogLine(reflect.TypeFor[TestFlex00](), reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "types.String to string",
			Source:     types.StringValue("a"),
			Target:     &testString,
			WantTarget: &testStringResult,
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[types.String](), reflect.TypeFor[*string]()),
				convertingLogLine(reflect.TypeFor[types.String](), reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "empty struct Source and Target",
			Source:     TestFlex00{},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[TestFlex00](), reflect.TypeFor[*TestFlex00]()),
				// convertingLogLine(reflect.TypeFor[TestFlex00](), reflect.TypeFor[TestFlex00]()),
			},
		},
		{
			TestName:   "empty struct pointer Source and Target",
			Source:     &TestFlex00{},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlex00](), reflect.TypeFor[*TestFlex00]()),
				// convertingLogLine(reflect.TypeFor[TestFlex00](), reflect.TypeFor[TestFlex00]()),
			},
		},
		{
			TestName:   "single string struct pointer Source and empty Target",
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF01](), reflect.TypeFor[*TestFlex00]()),
				// convertingLogLine(reflect.TypeFor[TestFlexTF01](), reflect.TypeFor[TestFlex00]()),
				noCorrespondingFieldLogLine(reflect.TypeFor[*TestFlexTF01](), "Field1", reflect.TypeFor[*TestFlex00]()),
			},
		},
		{
			TestName: "does not implement attr.Value Source",
			Source:   &TestFlexAWS01{Field1: "a"},
			Target:   &TestFlexAWS01{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "does not implement attr.Value: string"),
				diag.NewErrorDiagnostic("AutoFlEx", "convert (Field1)"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[*flex.TestFlexAWS01, *flex.TestFlexAWS01]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexAWS01](), reflect.TypeFor[*TestFlexAWS01]()),
				// convertingLogLine(reflect.TypeFor[TestFlexAWS01](), reflect.TypeFor[TestFlexAWS01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexAWS01](), "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "single string Source and single string Target",
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{Field1: "a"},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF01](), reflect.TypeFor[*TestFlexAWS01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF01](), "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "single string Source and single *string Target",
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlexAWS02{},
			WantTarget: &TestFlexAWS02{Field1: aws.String("a")},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF01](), reflect.TypeFor[*TestFlexAWS02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF01](), "Field1", reflect.TypeFor[*TestFlexAWS02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
			},
		},
		{
			TestName:   "single string Source and single int64 Target",
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlexAWS03{},
			WantTarget: &TestFlexAWS03{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF01](), reflect.TypeFor[*TestFlexAWS03]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF01](), "Field1", reflect.TypeFor[*TestFlexAWS03]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[int64]()),
				{
					"@level":             "info",
					"@module":            "provider.autoflex",
					"@message":           "AutoFlex Expand; incompatible types",
					"from":               map[string]any{},
					"to":                 float64(reflect.Int64),
					logAttrKeySourcePath: "Field1",
					logAttrKeyTargetPath: "Field1",
				},
			},
		},
		{
			TestName: "primitive types Source and primtive types Target",
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF03](), reflect.TypeFor[*TestFlexAWS04]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF03](), "Field1", reflect.TypeFor[*TestFlexAWS04]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*TestFlexTF03](), "Field2", reflect.TypeFor[*TestFlexAWS04]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[types.String](), "Field2", reflect.TypeFor[*string]()),
				matchedFieldsLogLine("Field3", reflect.TypeFor[*TestFlexTF03](), "Field3", reflect.TypeFor[*TestFlexAWS04]()),
				convertingWithPathLogLine("Field3", reflect.TypeFor[types.Int64](), "Field3", reflect.TypeFor[int32]()),
				matchedFieldsLogLine("Field4", reflect.TypeFor[*TestFlexTF03](), "Field4", reflect.TypeFor[*TestFlexAWS04]()),
				convertingWithPathLogLine("Field4", reflect.TypeFor[types.Int64](), "Field4", reflect.TypeFor[*int32]()),
				matchedFieldsLogLine("Field5", reflect.TypeFor[*TestFlexTF03](), "Field5", reflect.TypeFor[*TestFlexAWS04]()),
				convertingWithPathLogLine("Field5", reflect.TypeFor[types.Int64](), "Field5", reflect.TypeFor[int64]()),
				matchedFieldsLogLine("Field6", reflect.TypeFor[*TestFlexTF03](), "Field6", reflect.TypeFor[*TestFlexAWS04]()),
				convertingWithPathLogLine("Field6", reflect.TypeFor[types.Int64](), "Field6", reflect.TypeFor[*int64]()),
				matchedFieldsLogLine("Field7", reflect.TypeFor[*TestFlexTF03](), "Field7", reflect.TypeFor[*TestFlexAWS04]()),
				convertingWithPathLogLine("Field7", reflect.TypeFor[types.Float64](), "Field7", reflect.TypeFor[float32]()),
				matchedFieldsLogLine("Field8", reflect.TypeFor[*TestFlexTF03](), "Field8", reflect.TypeFor[*TestFlexAWS04]()),
				convertingWithPathLogLine("Field8", reflect.TypeFor[types.Float64](), "Field8", reflect.TypeFor[*float32]()),
				matchedFieldsLogLine("Field9", reflect.TypeFor[*TestFlexTF03](), "Field9", reflect.TypeFor[*TestFlexAWS04]()),
				convertingWithPathLogLine("Field9", reflect.TypeFor[types.Float64](), "Field9", reflect.TypeFor[float64]()),
				matchedFieldsLogLine("Field10", reflect.TypeFor[*TestFlexTF03](), "Field10", reflect.TypeFor[*TestFlexAWS04]()),
				convertingWithPathLogLine("Field10", reflect.TypeFor[types.Float64](), "Field10", reflect.TypeFor[*float64]()),
				matchedFieldsLogLine("Field11", reflect.TypeFor[*TestFlexTF03](), "Field11", reflect.TypeFor[*TestFlexAWS04]()),
				convertingWithPathLogLine("Field11", reflect.TypeFor[types.Bool](), "Field11", reflect.TypeFor[bool]()),
				matchedFieldsLogLine("Field12", reflect.TypeFor[*TestFlexTF03](), "Field12", reflect.TypeFor[*TestFlexAWS04]()),
				convertingWithPathLogLine("Field12", reflect.TypeFor[types.Bool](), "Field12", reflect.TypeFor[*bool]()),
			},
		},
		{
			TestName: "Collection of primitive types Source and slice or map of primtive types Target",
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF04](), reflect.TypeFor[*TestFlexAWS05]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF04](), "Field1", reflect.TypeFor[*TestFlexAWS05]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.List](), "Field1", reflect.TypeFor[[]string]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*TestFlexTF04](), "Field2", reflect.TypeFor[*TestFlexAWS05]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[types.List](), "Field2", reflect.TypeFor[[]*string]()),
				matchedFieldsLogLine("Field3", reflect.TypeFor[*TestFlexTF04](), "Field3", reflect.TypeFor[*TestFlexAWS05]()),
				convertingWithPathLogLine("Field3", reflect.TypeFor[types.Set](), "Field3", reflect.TypeFor[[]string]()),
				matchedFieldsLogLine("Field4", reflect.TypeFor[*TestFlexTF04](), "Field4", reflect.TypeFor[*TestFlexAWS05]()),
				convertingWithPathLogLine("Field4", reflect.TypeFor[types.Set](), "Field4", reflect.TypeFor[[]*string]()),
				matchedFieldsLogLine("Field5", reflect.TypeFor[*TestFlexTF04](), "Field5", reflect.TypeFor[*TestFlexAWS05]()),
				convertingWithPathLogLine("Field5", reflect.TypeFor[types.Map](), "Field5", reflect.TypeFor[map[string]string]()),
				matchedFieldsLogLine("Field6", reflect.TypeFor[*TestFlexTF04](), "Field6", reflect.TypeFor[*TestFlexAWS05]()),
				convertingWithPathLogLine("Field6", reflect.TypeFor[types.Map](), "Field6", reflect.TypeFor[map[string]*string]()),
			},
		},
		{
			TestName: "plural field names",
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF09](), reflect.TypeFor[*TestFlexAWS11]()),
				matchedFieldsLogLine("City", reflect.TypeFor[*TestFlexTF09](), "Cities", reflect.TypeFor[*TestFlexAWS11]()),
				convertingWithPathLogLine("City", reflect.TypeFor[types.List](), "Cities", reflect.TypeFor[[]*string]()),
				matchedFieldsLogLine("Coach", reflect.TypeFor[*TestFlexTF09](), "Coaches", reflect.TypeFor[*TestFlexAWS11]()),
				convertingWithPathLogLine("Coach", reflect.TypeFor[types.List](), "Coaches", reflect.TypeFor[[]*string]()),
				matchedFieldsLogLine("Tomato", reflect.TypeFor[*TestFlexTF09](), "Tomatoes", reflect.TypeFor[*TestFlexAWS11]()),
				convertingWithPathLogLine("Tomato", reflect.TypeFor[types.List](), "Tomatoes", reflect.TypeFor[[]*string]()),
				matchedFieldsLogLine("Vertex", reflect.TypeFor[*TestFlexTF09](), "Vertices", reflect.TypeFor[*TestFlexAWS11]()),
				convertingWithPathLogLine("Vertex", reflect.TypeFor[types.List](), "Vertices", reflect.TypeFor[[]*string]()),
				matchedFieldsLogLine("Criterion", reflect.TypeFor[*TestFlexTF09](), "Criteria", reflect.TypeFor[*TestFlexAWS11]()),
				convertingWithPathLogLine("Criterion", reflect.TypeFor[types.List](), "Criteria", reflect.TypeFor[[]*string]()),
				matchedFieldsLogLine("Datum", reflect.TypeFor[*TestFlexTF09](), "Data", reflect.TypeFor[*TestFlexAWS11]()),
				convertingWithPathLogLine("Datum", reflect.TypeFor[types.List](), "Data", reflect.TypeFor[[]*string]()),
				matchedFieldsLogLine("Hive", reflect.TypeFor[*TestFlexTF09](), "Hives", reflect.TypeFor[*TestFlexAWS11]()),
				convertingWithPathLogLine("Hive", reflect.TypeFor[types.List](), "Hives", reflect.TypeFor[[]*string]()),
			},
		},
		{
			TestName: "capitalization field names",
			Source: &TestFlexTF10{
				FieldURL: types.StringValue("h"),
			},
			Target: &TestFlexAWS12{},
			WantTarget: &TestFlexAWS12{
				FieldUrl: aws.String("h"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF10](), reflect.TypeFor[*TestFlexAWS12]()),
				matchedFieldsLogLine("FieldURL", reflect.TypeFor[*TestFlexTF10](), "FieldUrl", reflect.TypeFor[*TestFlexAWS12]()),
				convertingWithPathLogLine("FieldURL", reflect.TypeFor[types.String](), "FieldUrl", reflect.TypeFor[*string]()),
			},
		},
		{
			TestName:  "resource name prefix",
			ContextFn: func(ctx context.Context) context.Context { return context.WithValue(ctx, ResourcePrefix, "Intent") },
			Source: &TestFlexTF16{
				Name: types.StringValue("Ovodoghen"),
			},
			Target: &TestFlexAWS18{},
			WantTarget: &TestFlexAWS18{
				IntentName: aws.String("Ovodoghen"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF16](), reflect.TypeFor[*TestFlexAWS18]()),
				matchedFieldsLogLine("Name", reflect.TypeFor[*TestFlexTF16](), "IntentName", reflect.TypeFor[*TestFlexAWS18]()),
				convertingWithPathLogLine("Name", reflect.TypeFor[types.String](), "IntentName", reflect.TypeFor[*string]()),
			},
		},
		{
			TestName:   "single ARN Source and single string Target",
			Source:     &TestFlexTF17{Field1: fwtypes.ARNValue(testARN)},
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{Field1: testARN},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF17](), reflect.TypeFor[*TestFlexAWS01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF17](), "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ARN](), "Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "single ARN Source and single *string Target",
			Source:     &TestFlexTF17{Field1: fwtypes.ARNValue(testARN)},
			Target:     &TestFlexAWS02{},
			WantTarget: &TestFlexAWS02{Field1: aws.String(testARN)},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF17](), reflect.TypeFor[*TestFlexAWS02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF17](), "Field1", reflect.TypeFor[*TestFlexAWS02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ARN](), "Field1", reflect.TypeFor[*string]()),
			},
		},
		{
			TestName: "timestamp pointer",
			Source: &TestFlexTimeTF01{
				CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
			},
			Target: &TestFlexTimeAWS01{},
			WantTarget: &TestFlexTimeAWS01{
				CreationDateTime: &testTimeTime,
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTimeTF01](), reflect.TypeFor[*TestFlexTimeAWS01]()),
				matchedFieldsLogLine("CreationDateTime", reflect.TypeFor[*TestFlexTimeTF01](), "CreationDateTime", reflect.TypeFor[*TestFlexTimeAWS01]()),
				convertingWithPathLogLine("CreationDateTime", reflect.TypeFor[timetypes.RFC3339](), "CreationDateTime", reflect.TypeFor[*time.Time]()),
			},
		},
		{
			TestName: "timestamp",
			Source: &TestFlexTimeTF01{
				CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
			},
			Target: &TestFlexTimeAWS02{},
			WantTarget: &TestFlexTimeAWS02{
				CreationDateTime: testTimeTime,
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTimeTF01](), reflect.TypeFor[*TestFlexTimeAWS02]()),
				matchedFieldsLogLine("CreationDateTime", reflect.TypeFor[*TestFlexTimeTF01](), "CreationDateTime", reflect.TypeFor[*TestFlexTimeAWS02]()),
				convertingWithPathLogLine("CreationDateTime", reflect.TypeFor[timetypes.RFC3339](), "CreationDateTime", reflect.TypeFor[time.Time]()),
			},
		},
		{
			TestName: "JSONValue Source to json interface Target",
			Source:   &TestFlexTF20{Field1: fwtypes.SmithyJSONValue(`{"field1": "a"}`, newTestJSONDocument)},
			Target:   &TestFlexAWS19{},
			WantTarget: &TestFlexAWS19{
				Field1: &testJSONDocument{
					Value: map[string]any{
						"field1": "a",
					},
				},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF20](), reflect.TypeFor[*TestFlexAWS19]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF20](), "Field1", reflect.TypeFor[*TestFlexAWS19]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SmithyJSON[smithyjson.JSONStringer]](), "Field1", reflect.TypeFor[smithyjson.JSONStringer]()),
			},
		},
	}

	runAutoExpandTestCases(t, testCases)
}

func TestExpandGeneric(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		{
			TestName: "single list Source and *struct Target",
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF05](), reflect.TypeFor[*TestFlexAWS06]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF05](), "Field1", reflect.TypeFor[*TestFlexAWS06]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1.Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName: "single set Source and *struct Target",
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF06](), reflect.TypeFor[*TestFlexAWS06]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF06](), "Field1", reflect.TypeFor[*TestFlexAWS06]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1.Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "empty list Source and empty []struct Target",
			Source:     &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{})},
			Target:     &TestFlexAWS08{},
			WantTarget: &TestFlexAWS08{Field1: []TestFlexAWS01{}},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF05](), reflect.TypeFor[*TestFlexAWS08]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF05](), "Field1", reflect.TypeFor[*TestFlexAWS08]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]TestFlexAWS01]()),
			},
		},
		{
			TestName: "non-empty list Source and non-empty []struct Target",
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF05](), reflect.TypeFor[*TestFlexAWS08]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF05](), "Field1", reflect.TypeFor[*TestFlexAWS08]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1[0].Field1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("Field1[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1[1].Field1", reflect.TypeFor[types.String](), "Field1[1].Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "empty list Source and empty []*struct Target",
			Source:     &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*TestFlexTF01{})},
			Target:     &TestFlexAWS07{},
			WantTarget: &TestFlexAWS07{Field1: []*TestFlexAWS01{}},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF05](), reflect.TypeFor[*TestFlexAWS07]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF05](), "Field1", reflect.TypeFor[*TestFlexAWS07]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]*TestFlexAWS01]()),
			},
		},
		{
			TestName: "non-empty list Source and non-empty []*struct Target",
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF05](), reflect.TypeFor[*TestFlexAWS07]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF05](), "Field1", reflect.TypeFor[*TestFlexAWS07]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]*TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1[0].Field1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("Field1[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1[1].Field1", reflect.TypeFor[types.String](), "Field1[1].Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "empty set Source and empty []*struct Target",
			Source:     &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*TestFlexTF01{})},
			Target:     &TestFlexAWS07{},
			WantTarget: &TestFlexAWS07{Field1: []*TestFlexAWS01{}},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF06](), reflect.TypeFor[*TestFlexAWS07]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF06](), "Field1", reflect.TypeFor[*TestFlexAWS07]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]*TestFlexAWS01]()),
			},
		},
		{
			TestName: "non-empty set Source and non-empty []*struct Target",
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF06](), reflect.TypeFor[*TestFlexAWS07]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF06](), "Field1", reflect.TypeFor[*TestFlexAWS07]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]*TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1[0].Field1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("Field1[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1[1].Field1", reflect.TypeFor[types.String](), "Field1[1].Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName: "non-empty set Source and non-empty []struct Target",
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF06](), reflect.TypeFor[*TestFlexAWS08]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF06](), "Field1", reflect.TypeFor[*TestFlexAWS08]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[[]TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1[0].Field1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("Field1[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field1[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field1[1].Field1", reflect.TypeFor[types.String](), "Field1[1].Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName: "complex Source and complex Target",
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
				Field2: &TestFlexAWS06{Field1: &TestFlexAWS01{Field1: "n"}},
				Field3: aws.StringMap(map[string]string{"X": "x", "Y": "y"}),
				Field4: []TestFlexAWS03{{Field1: 100}, {Field1: 2000}, {Field1: 30000}},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*TestFlexTF07](), reflect.TypeFor[*TestFlexAWS09]()),

				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF07](), "Field1", reflect.TypeFor[*TestFlexAWS09]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*TestFlexTF07](), "Field2", reflect.TypeFor[*TestFlexAWS09]()),

				convertingWithPathLogLine("Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF05]](), "Field2", reflect.TypeFor[*TestFlexAWS06]()),
				matchedFieldsWithPathLogLine("Field2[0]", "Field1", reflect.TypeFor[*TestFlexTF05](), "Field2", "Field1", reflect.TypeFor[*TestFlexAWS06]()),
				convertingWithPathLogLine("Field2[0].Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field2.Field1", reflect.TypeFor[*TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("Field2[0].Field1[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "Field2.Field1", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("Field2[0].Field1[0].Field1", reflect.TypeFor[types.String](), "Field2.Field1.Field1", reflect.TypeFor[string]()),

				matchedFieldsLogLine("Field3", reflect.TypeFor[*TestFlexTF07](), "Field3", reflect.TypeFor[*TestFlexAWS09]()),
				convertingWithPathLogLine("Field3", reflect.TypeFor[types.Map](), "Field3", reflect.TypeFor[map[string]*string]()),

				matchedFieldsLogLine("Field4", reflect.TypeFor[*TestFlexTF07](), "Field4", reflect.TypeFor[*TestFlexAWS09]()),
				convertingWithPathLogLine("Field4", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF02]](), "Field4", reflect.TypeFor[[]TestFlexAWS03]()),
				matchedFieldsWithPathLogLine("Field4[0]", "Field1", reflect.TypeFor[*TestFlexTF02](), "Field4[0]", "Field1", reflect.TypeFor[*TestFlexAWS03]()),
				convertingWithPathLogLine("Field4[0].Field1", reflect.TypeFor[types.Int64](), "Field4[0].Field1", reflect.TypeFor[int64]()),
				matchedFieldsWithPathLogLine("Field4[1]", "Field1", reflect.TypeFor[*TestFlexTF02](), "Field4[1]", "Field1", reflect.TypeFor[*TestFlexAWS03]()),
				convertingWithPathLogLine("Field4[1].Field1", reflect.TypeFor[types.Int64](), "Field4[1].Field1", reflect.TypeFor[int64]()),
				matchedFieldsWithPathLogLine("Field4[2]", "Field1", reflect.TypeFor[*TestFlexTF02](), "Field4[2]", "Field1", reflect.TypeFor[*TestFlexAWS03]()),
				convertingWithPathLogLine("Field4[2].Field1", reflect.TypeFor[types.Int64](), "Field4[2].Field1", reflect.TypeFor[int64]()),
			},
		},
		{
			TestName: "map string",
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF11](), reflect.TypeFor[*TestFlexAWS13]()),
				matchedFieldsLogLine("FieldInner", reflect.TypeFor[*TestFlexTF11](), "FieldInner", reflect.TypeFor[*TestFlexAWS13]()),
				convertingWithPathLogLine("FieldInner", reflect.TypeFor[fwtypes.MapValueOf[basetypes.StringValue]](), "FieldInner", reflect.TypeFor[map[string]string]()),
			},
		},
		{
			TestName: "map of map of string",
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF21](), reflect.TypeFor[*TestFlexAWS21]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF21](), "Field1", reflect.TypeFor[*TestFlexAWS21]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[basetypes.StringValue]]](), "Field1", reflect.TypeFor[map[string]map[string]string]()),
			},
		},
		{
			TestName: "map of map of string pointer",
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF21](), reflect.TypeFor[*TestFlexAWS22]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*TestFlexTF21](), "Field1", reflect.TypeFor[*TestFlexAWS22]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[basetypes.StringValue]]](), "Field1", reflect.TypeFor[map[string]map[string]*string]()),
			},
		},
		{
			TestName: "nested string map",
			Source: &TestFlexTF14{
				FieldOuter: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &TestFlexTF11{
					FieldInner: fwtypes.NewMapValueOfMust[basetypes.StringValue](ctx, map[string]attr.Value{
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
				expandingLogLine(reflect.TypeFor[*TestFlexTF14](), reflect.TypeFor[*TestFlexAWS16]()),
				matchedFieldsLogLine("FieldOuter", reflect.TypeFor[*TestFlexTF14](), "FieldOuter", reflect.TypeFor[*TestFlexAWS16]()),
				convertingWithPathLogLine("FieldOuter", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF11]](), "FieldOuter", reflect.TypeFor[TestFlexAWS13]()),
				matchedFieldsWithPathLogLine("FieldOuter[0]", "FieldInner", reflect.TypeFor[*TestFlexTF11](), "FieldOuter", "FieldInner", reflect.TypeFor[*TestFlexAWS13]()),
				convertingWithPathLogLine("FieldOuter[0].FieldInner", reflect.TypeFor[fwtypes.MapValueOf[basetypes.StringValue]](), "FieldOuter.FieldInner", reflect.TypeFor[map[string]string]()),
			},
		},
		{
			TestName: "map block key list",
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
				expandingLogLine(reflect.TypeFor[*TestFlexMapBlockKeyTF01](), reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				matchedFieldsLogLine("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF01](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				convertingWithPathLogLine("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF02]](), "MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02]()),

				mapBlockKeyFieldLogLine("MapBlock[0]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock"),
				matchedFieldsWithPathLogLine("MapBlock[0]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock.Attr1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("MapBlock[0]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock.Attr2", reflect.TypeFor[string]()),

				mapBlockKeyFieldLogLine("MapBlock[1]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock"),
				matchedFieldsWithPathLogLine("MapBlock[1]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock.Attr1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("MapBlock[1]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock.Attr2", reflect.TypeFor[string]()),
			},
		},
		{
			TestName: "map block key set",
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
				expandingLogLine(reflect.TypeFor[*TestFlexMapBlockKeyTF03](), reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				matchedFieldsLogLine("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF03](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				convertingWithPathLogLine("MapBlock", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexMapBlockKeyTF02]](), "MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02]()),

				mapBlockKeyFieldLogLine("MapBlock[0]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock"),
				matchedFieldsWithPathLogLine("MapBlock[0]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock.Attr1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("MapBlock[0]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock.Attr2", reflect.TypeFor[string]()),

				mapBlockKeyFieldLogLine("MapBlock[1]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock"),
				matchedFieldsWithPathLogLine("MapBlock[1]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock.Attr1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("MapBlock[1]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock.Attr2", reflect.TypeFor[string]()),
			},
		},
		{
			TestName: "map block key ptr source",
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
				expandingLogLine(reflect.TypeFor[*TestFlexMapBlockKeyTF01](), reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				matchedFieldsLogLine("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF01](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				convertingWithPathLogLine("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF02]](), "MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02]()),

				mapBlockKeyFieldLogLine("MapBlock[0]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock"),
				matchedFieldsWithPathLogLine("MapBlock[0]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock.Attr1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("MapBlock[0]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock.Attr2", reflect.TypeFor[string]()),

				mapBlockKeyFieldLogLine("MapBlock[1]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock"),
				matchedFieldsWithPathLogLine("MapBlock[1]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock.Attr1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("MapBlock[1]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock.Attr2", reflect.TypeFor[string]()),
			},
		},
		{
			TestName: "map block key ptr both",
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
				expandingLogLine(reflect.TypeFor[*TestFlexMapBlockKeyTF01](), reflect.TypeFor[*TestFlexMapBlockKeyAWS03]()),
				matchedFieldsLogLine("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF01](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS03]()),
				convertingWithPathLogLine("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF02]](), "MapBlock", reflect.TypeFor[map[string]*TestFlexMapBlockKeyAWS02]()),

				mapBlockKeyFieldLogLine("MapBlock[0]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock"),
				matchedFieldsWithPathLogLine("MapBlock[0]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock.Attr1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("MapBlock[0]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock.Attr2", reflect.TypeFor[string]()),

				mapBlockKeyFieldLogLine("MapBlock[1]", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock"),
				matchedFieldsWithPathLogLine("MapBlock[1]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock.Attr1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("MapBlock[1]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF02](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock.Attr2", reflect.TypeFor[string]()),
			},
		},
		{
			TestName: "map block enum key",
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
				expandingLogLine(reflect.TypeFor[*TestFlexMapBlockKeyTF04](), reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				matchedFieldsLogLine("MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyTF04](), "MapBlock", reflect.TypeFor[*TestFlexMapBlockKeyAWS01]()),
				convertingWithPathLogLine("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF05]](), "MapBlock", reflect.TypeFor[map[string]TestFlexMapBlockKeyAWS02]()),

				mapBlockKeyFieldLogLine("MapBlock[0]", reflect.TypeFor[*TestFlexMapBlockKeyTF05](), "MapBlock"),
				matchedFieldsWithPathLogLine("MapBlock[0]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF05](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock.Attr1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("MapBlock[0]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF05](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock.Attr2", reflect.TypeFor[string]()),

				mapBlockKeyFieldLogLine("MapBlock[1]", reflect.TypeFor[*TestFlexMapBlockKeyTF05](), "MapBlock"),
				matchedFieldsWithPathLogLine("MapBlock[1]", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyTF05](), "MapBlock", "Attr1", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock.Attr1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("MapBlock[1]", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyTF05](), "MapBlock", "Attr2", reflect.TypeFor[*TestFlexMapBlockKeyAWS02]()),
				convertingWithPathLogLine("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock.Attr2", reflect.TypeFor[string]()),
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
		{
			TestName:   "single nested block pointer",
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{Field1: types.StringValue("a"), Field2: types.Int64Value(1)})},
			Target:     &aws02{},
			WantTarget: &aws02{Field1: &aws01{Field1: aws.String("a"), Field2: 1}},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*tf02](), "Field1", reflect.TypeFor[*aws02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1", reflect.TypeFor[*aws01]()),
				matchedFieldsWithPathLogLine("Field1", "Field1", reflect.TypeFor[*tf01](), "Field1", "Field1", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1.Field1", reflect.TypeFor[types.String](), "Field1.Field1", reflect.TypeFor[*string]()),
				matchedFieldsWithPathLogLine("Field1", "Field2", reflect.TypeFor[*tf01](), "Field1", "Field2", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1.Field2", reflect.TypeFor[types.Int64](), "Field1.Field2", reflect.TypeFor[int64]()),
			},
		},
		{
			TestName:   "single nested block nil",
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfNull[tf01](ctx)},
			Target:     &aws02{},
			WantTarget: &aws02{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws02]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*tf02](), "Field1", reflect.TypeFor[*aws02]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1", reflect.TypeFor[*aws01]()),
			},
		},
		{
			TestName:   "single nested block value",
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{Field1: types.StringValue("a"), Field2: types.Int64Value(1)})},
			Target:     &aws03{},
			WantTarget: &aws03{Field1: aws01{Field1: aws.String("a"), Field2: 1}},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws03]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*tf02](), "Field1", reflect.TypeFor[*aws03]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1", reflect.TypeFor[aws01]()),
				matchedFieldsWithPathLogLine("Field1", "Field1", reflect.TypeFor[*tf01](), "Field1", "Field1", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1.Field1", reflect.TypeFor[types.String](), "Field1.Field1", reflect.TypeFor[*string]()),
				matchedFieldsWithPathLogLine("Field1", "Field2", reflect.TypeFor[*tf01](), "Field1", "Field2", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1.Field2", reflect.TypeFor[types.Int64](), "Field1.Field2", reflect.TypeFor[int64]()),
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
		{
			TestName: "single nested block pointer",
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
				expandingLogLine(reflect.TypeFor[*tf03](), reflect.TypeFor[*aws03]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*tf03](), "Field1", reflect.TypeFor[*aws03]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf02]](), "Field1", reflect.TypeFor[*aws02]()),
				matchedFieldsWithPathLogLine("Field1", "Field1", reflect.TypeFor[*tf02](), "Field1", "Field1", reflect.TypeFor[*aws02]()),
				convertingWithPathLogLine("Field1.Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1.Field1", reflect.TypeFor[*aws01]()),
				matchedFieldsWithPathLogLine("Field1.Field1", "Field1", reflect.TypeFor[*tf01](), "Field1.Field1", "Field1", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1.Field1.Field1", reflect.TypeFor[types.Bool](), "Field1.Field1.Field1", reflect.TypeFor[bool]()),
				matchedFieldsWithPathLogLine("Field1.Field1", "Field2", reflect.TypeFor[*tf01](), "Field1.Field1", "Field2", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1.Field1.Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]](), "Field1.Field1.Field2", reflect.TypeFor[[]string]()),
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
		{
			TestName:   "valid value",
			Source:     fwtypes.StringEnumValue(TestEnumList),
			Target:     &testEnum,
			WantTarget: &testEnumList,
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), reflect.TypeFor[*TestEnum]()),
				convertingLogLine(reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), reflect.TypeFor[TestEnum]()),
			},
		},
		{
			TestName:   "empty value",
			Source:     fwtypes.StringEnumNull[TestEnum](),
			Target:     &testEnum,
			WantTarget: &testEnum,
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), reflect.TypeFor[*TestEnum]()),
				convertingLogLine(reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), reflect.TypeFor[TestEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandListOfInt64(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		{
			TestName: "valid value []int64",
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int64{},
			WantTarget: &[]int64{1, -1},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]int64]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]int64]()),
			},
		},
		{
			TestName:   "empty value []int64",
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]int64]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]int64]()),
			},
		},
		{
			TestName:   "null value []int64",
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]int64]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]int64]()),
			},
		},
		{
			TestName: "valid value []*int64",
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{aws.Int64(1), aws.Int64(-1)},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]*int64]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]*int64]()),
			},
		},
		{
			TestName:   "empty value []*int64",
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]*int64]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]*int64]()),
			},
		},
		{
			TestName:   "null value []*int64",
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]*int64]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]*int64]()),
			},
		},
		{
			TestName: "valid value []int32",
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int32{},
			WantTarget: &[]int32{1, -1},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]int32]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]int32]()),
			},
		},
		{
			TestName:   "empty value []int32",
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]int32]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]int32]()),
			},
		},
		{
			TestName:   "null value []int32",
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]int32]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]int32]()),
			},
		},
		{
			TestName: "valid value []*int32",
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{aws.Int32(1), aws.Int32(-1)},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]*int32]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]*int32]()),
			},
		},
		{
			TestName:   "empty value []*int32",
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]*int32]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]*int32]()),
			},
		},
		{
			TestName:   "null value []*int32",
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]*int32]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]*int32]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandSetOfInt64(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		{
			TestName: "valid value []int64",
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int64{},
			WantTarget: &[]int64{1, -1},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]int64]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]int64]()),
			},
		},
		{
			TestName:   "empty value []int64",
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]int64]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]int64]()),
			},
		},
		{
			TestName:   "null value []int64",
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]int64]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]int64]()),
			},
		},
		{
			TestName: "valid value []*int64",
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{aws.Int64(1), aws.Int64(-1)},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]*int64]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]*int64]()),
			},
		},
		{
			TestName:   "empty value []*int64",
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]*int64]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]*int64]()),
			},
		},
		{
			TestName:   "null value []*int64",
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]*int64]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]*int64]()),
			},
		},
		{
			TestName: "valid value []int32",
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int32{},
			WantTarget: &[]int32{1, -1},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]int32]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]int32]()),
			},
		},
		{
			TestName:   "empty value []int32",
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]int32]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]int32]()),
			},
		},
		{
			TestName:   "null value []int32",
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]int32]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]int32]()),
			},
		},
		{
			TestName: "valid value []*int32",
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{aws.Int32(1), aws.Int32(-1)},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]*int32]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]*int32]()),
			},
		},
		{
			TestName:   "empty value []*int32",
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]*int32]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]*int32]()),
			},
		},
		{
			TestName:   "null value []*int32",
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]*int32]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]*int32]()),
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
		{
			TestName: "valid value",
			Source: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue(string(testEnumFoo)),
				types.StringValue(string(testEnumBar)),
			}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{testEnumFoo, testEnumBar},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]testEnum]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]testEnum]()),
			},
		},
		{
			TestName:   "empty value",
			Source:     types.ListValueMust(types.StringType, []attr.Value{}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]testEnum]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]testEnum]()),
			},
		},
		{
			TestName:   "null value",
			Source:     types.ListNull(types.StringType),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[*[]testEnum]()),
				convertingLogLine(reflect.TypeFor[basetypes.ListValue](), reflect.TypeFor[[]testEnum]()),
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
		{
			TestName: "valid value",
			Source: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue(string(testEnumFoo)),
				types.StringValue(string(testEnumBar)),
			}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{testEnumFoo, testEnumBar},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]testEnum]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]testEnum]()),
			},
		},
		{
			TestName:   "empty value",
			Source:     types.SetValueMust(types.StringType, []attr.Value{}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]testEnum]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]testEnum]()),
			},
		},
		{
			TestName:   "null value",
			Source:     types.SetNull(types.StringType),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[*[]testEnum]()),
				convertingLogLine(reflect.TypeFor[basetypes.SetValue](), reflect.TypeFor[[]testEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandListOfNestedObject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		{
			TestName: "valid value to []struct",
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
				expandingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("[0].Field1", reflect.TypeFor[types.String](), "[0].Field1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("[1].Field1", reflect.TypeFor[types.String](), "[1].Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "empty value to []struct",
			Source:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{}),
			Target:     &[]TestFlexAWS01{},
			WantTarget: &[]TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]TestFlexAWS01]()),
			},
		},
		{
			TestName:   "null value to []struct",
			Source:     fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx),
			Target:     &[]TestFlexAWS01{},
			WantTarget: &[]TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]TestFlexAWS01]()),
			},
		},

		{
			TestName: "valid value to []*struct",
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
				expandingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]*TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]*TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("[0].Field1", reflect.TypeFor[types.String](), "[0].Field1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("[1].Field1", reflect.TypeFor[types.String](), "[1].Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "empty value to []*struct",
			Source:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{}),
			Target:     &[]*TestFlexAWS01{},
			WantTarget: &[]*TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]*TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]*TestFlexAWS01]()),
			},
		},
		{
			TestName:   "null value to []*struct",
			Source:     fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx),
			Target:     &[]*TestFlexAWS01{},
			WantTarget: &[]*TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]*TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]*TestFlexAWS01]()),
			},
		},

		{
			TestName: "single list value to single struct",
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
				expandingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("[0].Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "empty list value to single struct",
			Source:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{}),
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[TestFlexAWS01]()),
			},
		},
		{
			TestName:   "null value to single struct",
			Source:     fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx),
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[TestFlexAWS01]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandSetOfNestedObject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		{
			TestName: "valid value to []struct",
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
				expandingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("[0].Field1", reflect.TypeFor[types.String](), "[0].Field1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("[1].Field1", reflect.TypeFor[types.String](), "[1].Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "empty value to []struct",
			Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{}),
			Target:     &[]TestFlexAWS01{},
			WantTarget: &[]TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]TestFlexAWS01]()),
			},
		},
		{
			TestName:   "null value to []struct",
			Source:     fwtypes.NewSetNestedObjectValueOfNull[TestFlexTF01](ctx),
			Target:     &[]TestFlexAWS01{},
			WantTarget: &[]TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]TestFlexAWS01]()),
			},
		},

		{
			TestName: "valid value to []*struct",
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
				expandingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]*TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]*TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[0]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("[0].Field1", reflect.TypeFor[types.String](), "[0].Field1", reflect.TypeFor[string]()),
				matchedFieldsWithPathLogLine("[1]", "Field1", reflect.TypeFor[*TestFlexTF01](), "[1]", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("[1].Field1", reflect.TypeFor[types.String](), "[1].Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "empty value to []*struct",
			Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{}),
			Target:     &[]*TestFlexAWS01{},
			WantTarget: &[]*TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]*TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]*TestFlexAWS01]()),
			},
		},
		{
			TestName:   "null value to []*struct",
			Source:     fwtypes.NewSetNestedObjectValueOfNull[TestFlexTF01](ctx),
			Target:     &[]*TestFlexAWS01{},
			WantTarget: &[]*TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*[]*TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[[]*TestFlexAWS01]()),
			},
		},

		{
			TestName: "single set value to single struct",
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
				expandingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[TestFlexAWS01]()),
				matchedFieldsWithPathLogLine("[0]", "Field1", reflect.TypeFor[*TestFlexTF01](), "", "Field1", reflect.TypeFor[*TestFlexAWS01]()),
				convertingWithPathLogLine("[0].Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
			},
		},
		{
			TestName:   "empty set value to single struct",
			Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{}),
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[TestFlexAWS01]()),
			},
		},
		{
			TestName:   "null value to single struct",
			Source:     fwtypes.NewSetNestedObjectValueOfNull[TestFlexTF01](ctx),
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[*TestFlexAWS01]()),
				convertingLogLine(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[TestFlexTF01]](), reflect.TypeFor[TestFlexAWS01]()),
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
		{
			TestName:   "single nested valid value",
			Source:     &tf01{Field1: types.Int64Value(1), Field2: fwtypes.StringEnumValue(TestEnumList)},
			Target:     &aws01{},
			WantTarget: &aws01{Field1: 1, Field2: TestEnumList},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*tf01](), "Field1", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*tf01](), "Field2", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), "Field2", reflect.TypeFor[TestEnum]()),
			},
		},
		{
			TestName:   "single nested empty value",
			Source:     &tf01{Field1: types.Int64Value(1), Field2: fwtypes.StringEnumNull[TestEnum]()},
			Target:     &aws01{},
			WantTarget: &aws01{Field1: 1, Field2: ""},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*tf01](), "Field1", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*tf01](), "Field2", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), "Field2", reflect.TypeFor[TestEnum]()),
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
		{
			TestName: "single nested valid value",
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
				expandingLogLine(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*tf02](), "Field1", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*tf02](), "Field2", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]](), "Field2", reflect.TypeFor[*aws02]()),
				matchedFieldsWithPathLogLine("Field2[0]", "Field2", reflect.TypeFor[*tf01](), "Field2", "Field2", reflect.TypeFor[*aws02]()),
				convertingWithPathLogLine("Field2[0].Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), "Field2.Field2", reflect.TypeFor[TestEnum]()),
			},
		},
		{
			TestName: "single nested empty value",
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
				expandingLogLine(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*tf02](), "Field1", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				matchedFieldsLogLine("Field2", reflect.TypeFor[*tf02](), "Field2", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]](), "Field2", reflect.TypeFor[*aws02]()),
				matchedFieldsWithPathLogLine("Field2[0]", "Field2", reflect.TypeFor[*tf01](), "Field2", "Field2", reflect.TypeFor[*aws02]()),
				convertingWithPathLogLine("Field2[0].Field2", reflect.TypeFor[fwtypes.StringEnum[TestEnum]](), "Field2.Field2", reflect.TypeFor[TestEnum]()),
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
		{
			TestName:   "empty source with tags",
			Source:     &tf01{},
			Target:     &aws01{},
			WantTarget: &aws01{},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*tf01](), "Field1", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				ignoredFieldLogLine(reflect.TypeFor[*tf01](), "Tags"),
			},
		},
		{
			TestName: "ignore tags by default",
			Source: &tf01{
				Field1: types.BoolValue(true),
				Tags: fwtypes.NewMapValueOfMust[types.String](
					ctx,
					map[string]attr.Value{
						"foo": types.StringValue("bar"),
					},
				),
			},
			Target:     &aws01{},
			WantTarget: &aws01{Field1: true},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*tf01](), "Field1", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				ignoredFieldLogLine(reflect.TypeFor[*tf01](), "Tags"),
			},
		},
		{
			TestName: "include tags with option override",
			Options: []AutoFlexOptionsFunc{
				func(opts *AutoFlexOptions) {
					opts.SetIgnoredFields([]string{})
				},
			},
			Source: &tf01{
				Field1: types.BoolValue(true),
				Tags: fwtypes.NewMapValueOfMust[types.String](
					ctx,
					map[string]attr.Value{
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
				expandingLogLine(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[*tf01](), "Field1", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				matchedFieldsLogLine("Tags", reflect.TypeFor[*tf01](), "Tags", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Tags", reflect.TypeFor[fwtypes.MapValueOf[basetypes.StringValue]](), "Tags", reflect.TypeFor[map[string]string]()),
			},
		},
		{
			TestName: "ignore custom field",
			Options: []AutoFlexOptionsFunc{
				func(opts *AutoFlexOptions) {
					opts.SetIgnoredFields([]string{"Field1"})
				},
			},
			Source: &tf01{
				Field1: types.BoolValue(true),
				Tags: fwtypes.NewMapValueOfMust[types.String](
					ctx,
					map[string]attr.Value{
						"foo": types.StringValue("bar"),
					},
				),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Tags: map[string]string{"foo": "bar"},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				ignoredFieldLogLine(reflect.TypeFor[*tf01](), "Field1"),
				matchedFieldsLogLine("Tags", reflect.TypeFor[*tf01](), "Tags", reflect.TypeFor[*aws01]()),
				convertingWithPathLogLine("Tags", reflect.TypeFor[fwtypes.MapValueOf[basetypes.StringValue]](), "Tags", reflect.TypeFor[map[string]string]()),
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
		{
			TestName: "top level",
			Source: testFlexTFInterfaceFlexer{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			WantTarget: testFlexAWSInterfaceInterfacePtr(&testFlexAWSInterfaceInterfaceImpl{
				AWSField: "value1",
			}),
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFInterfaceFlexer](), reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				convertingLogLine(reflect.TypeFor[testFlexTFInterfaceFlexer](), reflect.TypeFor[testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "top level return value does not implement target interface",
			Source: testFlexTFInterfaceIncompatibleExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			expectedDiags: diag.Diagnostics{
				diagExpandedTypeDoesNotImplement(reflect.TypeFor[*testFlexAWSInterfaceIncompatibleImpl](), reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFInterfaceIncompatibleExpander, *flex.testFlexAWSInterfaceInterface]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFInterfaceIncompatibleExpander](), reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				convertingLogLine(reflect.TypeFor[testFlexTFInterfaceIncompatibleExpander](), reflect.TypeFor[testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "single list Source and single interface Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "single list non-Expander Source and single interface Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFListNestedObject[TestFlexTF01]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFListNestedObject[TestFlexTF01]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				{
					"@level":             "info",
					"@module":            "provider.autoflex",
					"@message":           "AutoFlex Expand; incompatible types",
					"from":               map[string]any{},
					"to":                 float64(reflect.Interface),
					logAttrKeySourcePath: "Field1",
					logAttrKeyTargetPath: "Field1",
				},
			},
		},
		{
			TestName: "single set Source and single interface Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "empty list Source and empty interface Target",
			Source: testFlexTFListNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "non-empty list Source and non-empty interface Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "empty set Source and empty interface Target",
			Source: testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "non-empty set Source and non-empty interface Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "object value Source and struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFObjectValue[testFlexTFInterfaceFlexer]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFObjectValue[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFInterfaceFlexer]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
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
		{
			TestName: "top level struct Target",
			Source: testFlexTFFlexer{
				Field1: types.StringValue("value1"),
			},
			Target: &testFlexAWSExpander{},
			WantTarget: &testFlexAWSExpander{
				AWSField: "value1",
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFFlexer](), reflect.TypeFor[*testFlexAWSExpander]()),
				// convertingLogLine(reflect.TypeFor[testFlexTFFlexer](), reflect.TypeFor[testFlexAWSExpander]()),
			},
		},
		{
			TestName: "top level string Target",
			Source: testFlexTFExpanderToString{
				Field1: types.StringValue("value1"),
			},
			Target:     aws.String(""),
			WantTarget: aws.String("value1"),
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderToString](), reflect.TypeFor[*string]()),
				convertingLogLine(reflect.TypeFor[testFlexTFExpanderToString](), reflect.TypeFor[string]()),
			},
		},
		{
			TestName: "top level incompatible struct Target",
			Source: testFlexTFFlexer{
				Field1: types.StringValue("value1"),
			},
			Target: &testFlexAWSExpanderIncompatible{},
			expectedDiags: diag.Diagnostics{
				diagCannotBeAssigned(reflect.TypeFor[testFlexAWSExpander](), reflect.TypeFor[testFlexAWSExpanderIncompatible]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFFlexer, *flex.testFlexAWSExpanderIncompatible]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFFlexer](), reflect.TypeFor[*testFlexAWSExpanderIncompatible]()),
				// convertingLogLine(reflect.TypeFor[testFlexTFFlexer](), reflect.TypeFor[testFlexAWSExpanderIncompatible]()),
			},
		},
		{
			TestName: "top level expands to nil",
			Source: testFlexTFExpanderToNil{
				Field1: types.StringValue("value1"),
			},
			Target: &testFlexAWSExpander{},
			expectedDiags: diag.Diagnostics{
				diagExpandsToNil(reflect.TypeFor[testFlexTFExpanderToNil]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFExpanderToNil, *flex.testFlexAWSExpander]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderToNil](), reflect.TypeFor[*testFlexAWSExpander]()),
				// convertingLogLine(reflect.TypeFor[testFlexTFExpanderToNil](), reflect.TypeFor[testFlexAWSExpander]()),
			},
		},
		{
			TestName: "top level incompatible non-struct Target",
			Source: testFlexTFExpanderToString{
				Field1: types.StringValue("value1"),
			},
			Target: aws.Int64(0),
			expectedDiags: diag.Diagnostics{
				diagCannotBeAssigned(reflect.TypeFor[string](), reflect.TypeFor[int64]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFExpanderToString, *int64]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderToString](), reflect.TypeFor[*int64]()),
				convertingLogLine(reflect.TypeFor[testFlexTFExpanderToString](), reflect.TypeFor[int64]()),
			},
		},
		{
			TestName: "single list Source and single struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[testFlexAWSExpander]()),
			},
		},
		{
			TestName: "single set Source and single struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[testFlexAWSExpander]()),
			},
		},
		{
			TestName: "single list Source and single *struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
			},
		},
		{
			TestName: "single set Source and single *struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
			},
		},
		{
			TestName: "empty list Source and empty struct Target",
			Source: testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{}),
			},
			Target: &testFlexAWSExpanderStructSlice{},
			WantTarget: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
			},
		},
		{
			TestName: "non-empty list Source and non-empty struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
			},
		},
		{
			TestName: "empty list Source and empty *struct Target",
			Source: testFlexTFExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{}),
			},
			Target: &testFlexAWSExpanderPtrSlice{},
			WantTarget: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
			},
		},
		{
			TestName: "non-empty list Source and non-empty *struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
			},
		},
		{
			TestName: "empty set Source and empty struct Target",
			Source: testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{}),
			},
			Target: &testFlexAWSExpanderStructSlice{},
			WantTarget: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
			},
		},
		{
			TestName: "non-empty set Source and non-empty struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
			},
		},
		{
			TestName: "empty set Source and empty *struct Target",
			Source: testFlexTFExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFFlexer{}),
			},
			Target: &testFlexAWSExpanderPtrSlice{},
			WantTarget: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
			},
		},
		{
			TestName: "non-empty set Source and non-empty *struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
			},
		},
		{
			TestName: "object value Source and struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderObjectValue](), reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderObjectValue](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[testFlexAWSExpander]()),
			},
		},
		{
			TestName: "object value Source and *struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFExpanderObjectValue](), reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFExpanderObjectValue](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFFlexer]](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
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
		{
			TestName: "top level",
			Source: testFlexTFInterfaceTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			WantTarget: testFlexAWSInterfaceInterfacePtr(&testFlexAWSInterfaceInterfaceImpl{
				AWSField: "value1",
			}),
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFInterfaceTypedExpander](), reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				convertingLogLine(reflect.TypeFor[testFlexTFInterfaceTypedExpander](), reflect.TypeFor[testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "top level return value does not implement target interface",
			Source: testFlexTFInterfaceIncompatibleTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			expectedDiags: diag.Diagnostics{
				diagExpandedTypeDoesNotImplement(reflect.TypeFor[*testFlexAWSInterfaceIncompatibleImpl](), reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFInterfaceIncompatibleTypedExpander, *flex.testFlexAWSInterfaceInterface]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFInterfaceIncompatibleTypedExpander](), reflect.TypeFor[*testFlexAWSInterfaceInterface]()),
				convertingLogLine(reflect.TypeFor[testFlexTFInterfaceIncompatibleTypedExpander](), reflect.TypeFor[testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "single list Source and single interface Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "single list non-Expander Source and single interface Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFListNestedObject[TestFlexTF01]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFListNestedObject[TestFlexTF01]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[TestFlexTF01]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
				{
					"@level":             "info",
					"@module":            "provider.autoflex",
					"@message":           "AutoFlex Expand; incompatible types",
					"from":               map[string]any{},
					"to":                 float64(reflect.Interface),
					logAttrKeySourcePath: "Field1",
					logAttrKeyTargetPath: "Field1",
				},
			},
		},
		{
			TestName: "single set Source and single interface Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "empty list Source and empty interface Target",
			Source: testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceTypedExpander{}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "non-empty list Source and non-empty interface Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFListNestedObject[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "empty set Source and empty interface Target",
			Source: testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceTypedExpander{}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "non-empty set Source and non-empty interface Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSInterfaceInterface]()),
			},
		},
		{
			TestName: "object value Source and struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFObjectValue[testFlexTFInterfaceTypedExpander]](), reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFObjectValue[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSInterfaceSingle]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFInterfaceTypedExpander]](), "Field1", reflect.TypeFor[testFlexAWSInterfaceInterface]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandTypedExpander(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		{
			TestName: "top level struct Target",
			Source: testFlexTFTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &testFlexAWSExpander{},
			WantTarget: &testFlexAWSExpander{
				AWSField: "value1",
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpander](), reflect.TypeFor[*testFlexAWSExpander]()),
				// convertingLogLine(reflect.TypeFor[testFlexTFTypedExpander](), reflect.TypeFor[testFlexAWSExpander]()),
			},
		},
		{
			TestName: "top level incompatible struct Target",
			Source: testFlexTFTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &testFlexAWSExpanderIncompatible{},
			expectedDiags: diag.Diagnostics{
				diagCannotBeAssigned(reflect.TypeFor[testFlexAWSExpander](), reflect.TypeFor[testFlexAWSExpanderIncompatible]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFTypedExpander, *flex.testFlexAWSExpanderIncompatible]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpander](), reflect.TypeFor[*testFlexAWSExpanderIncompatible]()),
				// convertingLogLine(reflect.TypeFor[testFlexTFTypedExpander](), reflect.TypeFor[testFlexAWSExpander]()),
			},
		},
		{
			TestName: "top level expands to nil",
			Source: testFlexTFTypedExpanderToNil{
				Field1: types.StringValue("value1"),
			},
			Target: &testFlexAWSExpander{},
			expectedDiags: diag.Diagnostics{
				diagExpandsToNil(reflect.TypeFor[testFlexTFTypedExpanderToNil]()),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.testFlexTFTypedExpanderToNil, *flex.testFlexAWSExpander]"),
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderToNil](), reflect.TypeFor[*testFlexAWSExpander]()),
				// convertingLogLine(reflect.TypeFor[testFlexTFTypedExpander](), reflect.TypeFor[testFlexAWSExpander]()),
			},
		},
		{
			TestName: "single list Source and single struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[testFlexAWSExpander]()),
			},
		},
		{
			TestName: "single set Source and single struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFTypedExpander]](), reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFSetNestedObject[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[testFlexAWSExpander]()),
			},
		},
		{
			TestName: "single list Source and single *struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
			},
		},
		{
			TestName: "single set Source and single *struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
			},
		},
		{
			TestName: "empty list Source and empty struct Target",
			Source: testFlexTFTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{}),
			},
			Target: &testFlexAWSExpanderStructSlice{},
			WantTarget: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
			},
		},
		{
			TestName: "non-empty list Source and non-empty struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
			},
		},
		{
			TestName: "empty list Source and empty *struct Target",
			Source: testFlexTFTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{}),
			},
			Target: &testFlexAWSExpanderPtrSlice{},
			WantTarget: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
			},
		},
		{
			TestName: "non-empty list Source and non-empty *struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
			},
		},
		{
			TestName: "empty set Source and empty struct Target",
			Source: testFlexTFTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{}),
			},
			Target: &testFlexAWSExpanderStructSlice{},
			WantTarget: &testFlexAWSExpanderStructSlice{
				Field1: []testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
			},
		},
		{
			TestName: "non-empty set Source and non-empty struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderStructSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]testFlexAWSExpander]()),
			},
		},
		{
			TestName: "empty set Source and empty *struct Target",
			Source: testFlexTFTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFTypedExpander{}),
			},
			Target: &testFlexAWSExpanderPtrSlice{},
			WantTarget: &testFlexAWSExpanderPtrSlice{
				Field1: []*testFlexAWSExpander{},
			},
			expectedLogLines: []map[string]any{
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
			},
		},
		{
			TestName: "non-empty set Source and non-empty *struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*testFlexAWSExpanderPtrSlice]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[[]*testFlexAWSExpander]()),
			},
		},
		{
			TestName: "object value Source and struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderObjectValue](), reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderObjectValue](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSingleStruct]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[testFlexAWSExpander]()),
			},
		},
		{
			TestName: "object value Source and *struct Target",
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
				expandingLogLine(reflect.TypeFor[testFlexTFTypedExpanderObjectValue](), reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				matchedFieldsLogLine("Field1", reflect.TypeFor[testFlexTFTypedExpanderObjectValue](), "Field1", reflect.TypeFor[*testFlexAWSExpanderSinglePtr]()),
				convertingWithPathLogLine("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[testFlexTFTypedExpander]](), "Field1", reflect.TypeFor[*testFlexAWSExpander]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

type autoFlexTestCase struct {
	ContextFn        func(context.Context) context.Context
	Options          []AutoFlexOptionsFunc
	TestName         string
	Source           any
	Target           any
	expectedDiags    diag.Diagnostics
	expectedLogLines []map[string]any
	WantTarget       any
	WantDiff         bool
}

type autoFlexTestCases []autoFlexTestCase

func runAutoExpandTestCases(t *testing.T, testCases autoFlexTestCases) {
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
