// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestGenericFlatten(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testString := "test"
	testCases := []struct {
		TestName   string
		Source     any
		Target     any
		WantErr    bool
		WantTarget any
	}{
		{
			TestName: "nil Source and Target",
			WantErr:  true,
		},
		{
			TestName: "non-pointer Target",
			Source:   TestFlex00{},
			Target:   0,
			WantErr:  true,
		},
		{
			TestName: "non-struct Source",
			Source:   testString,
			Target:   &TestFlex00{},
			WantErr:  true,
		},
		{
			TestName: "non-struct Target",
			Source:   TestFlex00{},
			Target:   &testString,
			WantErr:  true,
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
			Source:     &TestFlexAWS01{Field1: "a"},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
		},
		{
			TestName: "does not implement attr.Value Target",
			Source:   &TestFlexAWS01{Field1: "a"},
			Target:   &TestFlexAWS01{},
			WantErr:  true,
		},
		{
			TestName:   "single empty string Source and single string Target",
			Source:     &TestFlexAWS01{},
			Target:     &TestFlexTF01{},
			WantTarget: &TestFlexTF01{Field1: types.StringValue("")},
		},
		{
			TestName:   "single string Source and single string Target",
			Source:     &TestFlexAWS01{Field1: "a"},
			Target:     &TestFlexTF01{},
			WantTarget: &TestFlexTF01{Field1: types.StringValue("a")},
		},
		{
			TestName:   "single nil *string Source and single string Target",
			Source:     &TestFlexAWS02{},
			Target:     &TestFlexTF01{},
			WantTarget: &TestFlexTF01{Field1: types.StringNull()},
		},
		{
			TestName:   "single *string Source and single string Target",
			Source:     &TestFlexAWS02{Field1: aws.String("a")},
			Target:     &TestFlexTF01{},
			WantTarget: &TestFlexTF01{Field1: types.StringValue("a")},
		},
		{
			TestName:   "single string Source and single int64 Target",
			Source:     &TestFlexAWS01{Field1: "a"},
			Target:     &TestFlexTF02{},
			WantTarget: &TestFlexTF02{},
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
				Field1: types.StringValue("field1"),
				Field2: types.StringValue("field2"),
				Field3: types.Int64Value(3),
				Field4: types.Int64Value(-4),
				Field5: types.Int64Value(5),
				Field6: types.Int64Value(-6),
				// float32 -> float64 precision problems.
				Field7:  types.Float64Value(float64(float32(7.7))),
				Field8:  types.Float64Value(float64(float32(-8.8))),
				Field9:  types.Float64Value(9.99),
				Field10: types.Float64Value(-10.101),
				Field11: types.BoolValue(true),
				Field12: types.BoolValue(false),
			},
		},
		{
			TestName: "zero value slice/map of primtive primtive types Source and List/Set/Map of primtive types Target",
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
		},
		{
			TestName: "slice/map of primtive primtive types Source and List/Set/Map of primtive types Target",
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
		},
		{
			TestName:   "nil *struct Source and single list Target",
			Source:     &TestFlexAWS06{},
			Target:     &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx)},
		},
		{
			TestName:   "*struct Source and single list Target",
			Source:     &TestFlexAWS06{Field1: &TestFlexAWS01{Field1: "a"}},
			Target:     &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfPtr(ctx, &TestFlexTF01{Field1: types.StringValue("a")})},
		},
		{
			TestName:   "*struct Source and single set Target",
			Source:     &TestFlexAWS06{Field1: &TestFlexAWS01{Field1: "a"}},
			Target:     &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfPtr(ctx, &TestFlexTF01{Field1: types.StringValue("a")})},
		},
		{
			TestName:   "nil []struct and null list Target",
			Source:     &TestFlexAWS08{},
			Target:     &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx)},
		},
		{
			TestName:   "nil []struct and null set Target",
			Source:     &TestFlexAWS08{},
			Target:     &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfNull[TestFlexTF01](ctx)},
		},
		{
			TestName:   "empty []struct and empty list Target",
			Source:     &TestFlexAWS08{Field1: []TestFlexAWS01{}},
			Target:     &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []TestFlexTF01{})},
		},
		{
			TestName:   "empty []struct and empty struct Target",
			Source:     &TestFlexAWS08{Field1: []TestFlexAWS01{}},
			Target:     &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, []TestFlexTF01{})},
		},
		{
			TestName: "non-empty []struct and non-empty list Target",
			Source: &TestFlexAWS08{Field1: []TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
			Target: &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
		},
		{
			TestName: "non-empty []struct and non-empty set Target",
			Source: &TestFlexAWS08{Field1: []TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
			Target: &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, []TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
		},
		{
			TestName:   "nil []*struct and null list Target",
			Source:     &TestFlexAWS07{},
			Target:     &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfNull[TestFlexTF01](ctx)},
		},
		{
			TestName:   "nil []*struct and null set Target",
			Source:     &TestFlexAWS07{},
			Target:     &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfNull[TestFlexTF01](ctx)},
		},
		{
			TestName:   "empty []*struct and empty list Target",
			Source:     &TestFlexAWS07{Field1: []*TestFlexAWS01{}},
			Target:     &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfSlice(ctx, []*TestFlexTF01{})},
		},
		{
			TestName:   "empty []*struct and empty set Target",
			Source:     &TestFlexAWS07{Field1: []*TestFlexAWS01{}},
			Target:     &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfSlice(ctx, []*TestFlexTF01{})},
		},
		{
			TestName: "non-empty []*struct and non-empty list Target",
			Source: &TestFlexAWS07{Field1: []*TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
			Target: &TestFlexTF05{},
			WantTarget: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfSlice(ctx, []*TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
		},
		{
			TestName: "non-empty []*struct and non-empty set Target",
			Source: &TestFlexAWS07{Field1: []*TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
			Target: &TestFlexTF06{},
			WantTarget: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfSlice(ctx, []*TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
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
				Field2: fwtypes.NewListNestedObjectValueOfPtr(ctx, &TestFlexTF05{
					Field1: fwtypes.NewListNestedObjectValueOfPtr(ctx, &TestFlexTF01{
						Field1: types.StringValue("n"),
					}),
				}),
				Field3: types.MapValueMust(types.StringType, map[string]attr.Value{
					"X": types.StringValue("x"),
					"Y": types.StringValue("y"),
				}),
				Field4: fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, []TestFlexTF02{
					{Field1: types.Int64Value(100)},
					{Field1: types.Int64Value(2000)},
					{Field1: types.Int64Value(30000)},
				}),
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			err := Flatten(ctx, testCase.Source, testCase.Target)
			gotErr := err != nil

			if gotErr != testCase.WantErr {
				t.Errorf("gotErr = %v, wantErr = %v", gotErr, testCase.WantErr)
			}

			if gotErr {
				if !testCase.WantErr {
					t.Errorf("err = %q", err)
				}
			} else if diff := cmp.Diff(testCase.Target, testCase.WantTarget); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
