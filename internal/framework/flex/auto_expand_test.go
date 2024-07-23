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
		},
		{
			TestName: "typed nil Source",
			Source:   typedNilSource,
			Target:   &TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Cannot expand nil source"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[*flex.TestFlex00, *flex.TestFlex00]"),
			},
		},
		{
			TestName: "nil Target",
			Source:   TestFlex00{},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "Target cannot be nil"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.TestFlex00, <nil>]"),
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
		},
		{
			TestName: "non-pointer Target",
			Source:   TestFlex00{},
			Target:   0,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "target (int): int, want pointer"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.TestFlex00, int]"),
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
		},
		{
			TestName: "non-struct Target",
			Source:   TestFlex00{},
			Target:   &testString,
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic("AutoFlEx", "does not implement attr.Value: struct"),
				diag.NewErrorDiagnostic("AutoFlEx", "Expand[flex.TestFlex00, *string]"),
			},
		},
		{
			TestName:   "types.String to string",
			Source:     types.StringValue("a"),
			Target:     &testString,
			WantTarget: &testStringResult,
		},
		{
			TestName:   "empty struct Source and Target",
			Source:     TestFlex00{},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
		},
		{
			TestName:   "empty struct pointer Source and Target",
			Source:     &TestFlex00{},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
		},
		{
			TestName:   "single string struct pointer Source and empty Target",
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
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
		},
		{
			TestName:   "single string Source and single string Target",
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{Field1: "a"},
		},
		{
			TestName:   "single string Source and single *string Target",
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlexAWS02{},
			WantTarget: &TestFlexAWS02{Field1: aws.String("a")},
		},
		{
			TestName:   "single string Source and single int64 Target",
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlexAWS03{},
			WantTarget: &TestFlexAWS03{},
			expectedLogLines: []map[string]any{
				{
					"@level":   "info",
					"@module":  "provider",
					"@message": "AutoFlex Expand; incompatible types",
					"from":     map[string]any{},
					"to":       float64(reflect.Int64),
				},
			},
		},
		{
			TestName: "primtive types Source and primtive types Target",
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
		},
		{
			TestName: "List/Set/Map of primitive types Source and slice/map of primtive types Target",
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
		},
		{
			ContextFn: func(ctx context.Context) context.Context { return context.WithValue(ctx, ResourcePrefix, "Intent") },
			TestName:  "resource name prefix",
			Source: &TestFlexTF16{
				Name: types.StringValue("Ovodoghen"),
			},
			Target: &TestFlexAWS18{},
			WantTarget: &TestFlexAWS18{
				IntentName: aws.String("Ovodoghen"),
			},
		},
		{
			TestName:   "single ARN Source and single string Target",
			Source:     &TestFlexTF17{Field1: fwtypes.ARNValue(testARN)},
			Target:     &TestFlexAWS01{},
			WantTarget: &TestFlexAWS01{Field1: testARN},
		},
		{
			TestName:   "single ARN Source and single *string Target",
			Source:     &TestFlexTF17{Field1: fwtypes.ARNValue(testARN)},
			Target:     &TestFlexAWS02{},
			WantTarget: &TestFlexAWS02{Field1: aws.String(testARN)},
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
		},
		{
			TestName: "string Source to interface Target",
			Source:   &TestFlexTF20{Field1: fwtypes.SmithyJSONValue(`{"field1": "a"}`, newTestJSONDocument)},
			Target:   &TestFlexAWS19{},
			WantTarget: &TestFlexAWS19{
				Field1: &testJSONDocument{
					Value: map[string]any{
						"field1": "a",
					},
				},
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
			TestName:   "single list Source and *struct Target",
			Source:     &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &TestFlexTF01{Field1: types.StringValue("a")})},
			Target:     &TestFlexAWS06{},
			WantTarget: &TestFlexAWS06{Field1: &TestFlexAWS01{Field1: "a"}},
		},
		{
			TestName:   "single set Source and *struct Target",
			Source:     &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &TestFlexTF01{Field1: types.StringValue("a")})},
			Target:     &TestFlexAWS06{},
			WantTarget: &TestFlexAWS06{Field1: &TestFlexAWS01{Field1: "a"}},
		},
		{
			TestName:   "empty list Source and empty []struct Target",
			Source:     &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{})},
			Target:     &TestFlexAWS08{},
			WantTarget: &TestFlexAWS08{Field1: []TestFlexAWS01{}},
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
		},
		{
			TestName:   "empty list Source and empty []*struct Target",
			Source:     &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*TestFlexTF01{})},
			Target:     &TestFlexAWS07{},
			WantTarget: &TestFlexAWS07{Field1: []*TestFlexAWS01{}},
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
		},
		{
			TestName:   "empty list Source and empty []struct Target",
			Source:     &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []TestFlexTF01{})},
			Target:     &TestFlexAWS08{},
			WantTarget: &TestFlexAWS08{Field1: []TestFlexAWS01{}},
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
		},
		{
			TestName:   "empty set Source and empty []*struct Target",
			Source:     &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*TestFlexTF01{})},
			Target:     &TestFlexAWS07{},
			WantTarget: &TestFlexAWS07{Field1: []*TestFlexAWS01{}},
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
		},
		{
			TestName:   "single nested block nil",
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfNull[tf01](ctx)},
			Target:     &aws02{},
			WantTarget: &aws02{},
		},
		{
			TestName:   "single nested block value",
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{Field1: types.StringValue("a"), Field2: types.Int64Value(1)})},
			Target:     &aws03{},
			WantTarget: &aws03{Field1: aws01{Field1: aws.String("a"), Field2: 1}},
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
			Target:     &aws03{},
			WantTarget: &aws03{Field1: &aws02{Field1: &aws01{Field1: true, Field2: []string{"a", "b"}}}},
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
		},
		{
			TestName:   "empty value",
			Source:     fwtypes.StringEnumNull[TestEnum](),
			Target:     &testEnum,
			WantTarget: &testEnum,
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
		},
		{
			TestName:   "empty value []int64",
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
		},
		{
			TestName:   "null value []int64",
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
		},
		{
			TestName: "valid value []*int64",
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{aws.Int64(1), aws.Int64(-1)},
		},
		{
			TestName:   "empty value []*int64",
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
		},
		{
			TestName:   "null value []*int64",
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
		},
		{
			TestName: "valid value []int32",
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int32{},
			WantTarget: &[]int32{1, -1},
		},
		{
			TestName:   "empty value []int32",
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
		},
		{
			TestName:   "null value []int32",
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
		},
		{
			TestName: "valid value []*int32",
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{aws.Int32(1), aws.Int32(-1)},
		},
		{
			TestName:   "empty value []*int32",
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
		},
		{
			TestName:   "null value []*int32",
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
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
		},
		{
			TestName:   "empty value []int64",
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
		},
		{
			TestName:   "null value []int64",
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
		},
		{
			TestName: "valid value []*int64",
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{aws.Int64(1), aws.Int64(-1)},
		},
		{
			TestName:   "empty value []*int64",
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
		},
		{
			TestName:   "null value []*int64",
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
		},
		{
			TestName: "valid value []int32",
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int32{},
			WantTarget: &[]int32{1, -1},
		},
		{
			TestName:   "empty value []int32",
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
		},
		{
			TestName:   "null value []int32",
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
		},
		{
			TestName: "valid value []*int32",
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{aws.Int32(1), aws.Int32(-1)},
		},
		{
			TestName:   "empty value []*int32",
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
		},
		{
			TestName:   "null value []*int32",
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
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
		},
		{
			TestName:   "empty value",
			Source:     types.ListValueMust(types.StringType, []attr.Value{}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
		},
		{
			TestName:   "null value",
			Source:     types.ListNull(types.StringType),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
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
		},
		{
			TestName:   "empty value",
			Source:     types.SetValueMust(types.StringType, []attr.Value{}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
		},
		{
			TestName:   "null value",
			Source:     types.SetNull(types.StringType),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
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
		},
		{
			TestName:   "single nested empty value",
			Source:     &tf01{Field1: types.Int64Value(1), Field2: fwtypes.StringEnumNull[TestEnum]()},
			Target:     &aws01{},
			WantTarget: &aws01{Field1: 1, Field2: ""},
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
			TestName:   "single nested valid value",
			Source:     &tf02{Field1: types.Int64Value(1), Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{Field2: fwtypes.StringEnumValue(TestEnumList)})},
			Target:     &aws01{},
			WantTarget: &aws01{Field1: 1, Field2: &aws02{Field2: TestEnumList}},
		},
		{
			TestName:   "single nested empty value",
			Source:     &tf02{Field1: types.Int64Value(1), Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{Field2: fwtypes.StringEnumNull[TestEnum]()})},
			Target:     &aws01{},
			WantTarget: &aws01{Field1: 1, Field2: &aws02{Field2: ""}},
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
		},
		{
			TestName: "single list Source and single interface Target",
			Source: testFlexTFInterfaceListNestedObject{
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
		},
		{
			TestName: "single list non-Expander Source and single interface Target",
			Source: testFlexTFInterfaceListNestedObjectNonFlexer{
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
				{
					"@level":   "info",
					"@module":  "provider",
					"@message": "AutoFlex Expand; incompatible types",
					"from":     map[string]any{},
					"to":       float64(reflect.Interface),
				},
			},
		},
		{
			TestName: "single set Source and single interface Target",
			Source: testFlexTFInterfaceSetNestedObject{
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
		},
		{
			TestName: "empty list Source and empty interface Target",
			Source: testFlexTFInterfaceListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{},
			},
		},
		{
			TestName: "non-empty list Source and non-empty interface Target",
			Source: testFlexTFInterfaceListNestedObject{
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
		},
		{
			TestName: "empty set Source and empty interface Target",
			Source: testFlexTFInterfaceSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []testFlexTFInterfaceFlexer{}),
			},
			Target: &testFlexAWSInterfaceSlice{},
			WantTarget: &testFlexAWSInterfaceSlice{
				Field1: []testFlexAWSInterfaceInterface{},
			},
		},
		{
			TestName: "non-empty set Source and non-empty interface Target",
			Source: testFlexTFInterfaceListNestedObject{
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
		},
		{
			TestName: "object value Source and struct Target",
			Source: testFlexTFInterfaceObjectValue{
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
		},
		{
			TestName: "top level string Target",
			Source: testFlexTFExpanderToString{
				Field1: types.StringValue("value1"),
			},
			Target:     aws.String(""),
			WantTarget: aws.String("value1"),
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
