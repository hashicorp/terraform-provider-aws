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

type TestFlex00 struct{}

type TestFlexTF01 struct {
	Field1 types.String `tfsdk:"field1"`
}

type TestFlexTF02 struct {
	Field1 types.Int64 `tfsdk:"field1"`
}

// All primitive types.
type TestFlexTF03 struct {
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

// List/Set/Map of primitive types.
type TestFlexTF04 struct {
	Field1 types.List `tfsdk:"field1"`
	Field2 types.List `tfsdk:"field2"`
	Field3 types.Set  `tfsdk:"field3"`
	Field4 types.Set  `tfsdk:"field4"`
	Field5 types.Map  `tfsdk:"field5"`
	Field6 types.Map  `tfsdk:"field6"`
}

type TestFlexTF05 struct {
	Field1 fwtypes.ListNestedObjectValueOf[TestFlexTF01] `tfsdk:"field1"`
}

type TestFlexTF06 struct {
	Field1 fwtypes.SetNestedObjectValueOf[TestFlexTF01] `tfsdk:"field1"`
}

type TestFlexAWS01 struct {
	Field1 string
}

type TestFlexAWS02 struct {
	Field1 *string
}

type TestFlexAWS03 struct {
	Field1 int64
}

type TestFlexAWS04 struct {
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

type TestFlexAWS05 struct {
	Field1 []string
	Field2 []*string
	Field3 []string
	Field4 []*string
	Field5 map[string]string
	Field6 map[string]*string
}

type TestFlexAWS06 struct {
	Field1 *TestFlexAWS01
}

type TestFlexAWS07 struct {
	Field1 []*TestFlexAWS01
}

type TestFlexAWS08 struct {
	Field1 []TestFlexAWS01
}

func TestGenericExpand(t *testing.T) {
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
			Source:     &TestFlexTF01{Field1: types.StringValue("a")},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
		},
		{
			TestName: "does not implement attr.Value Source",
			Source:   &TestFlexAWS01{Field1: "a"},
			Target:   &TestFlexAWS01{},
			WantErr:  true,
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
			WantErr:    true,
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
			TestName:   "single list Source and *struct Target",
			Source:     &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfPtr(ctx, &TestFlexTF01{Field1: types.StringValue("a")})},
			Target:     &TestFlexAWS06{},
			WantTarget: &TestFlexAWS06{Field1: &TestFlexAWS01{Field1: "a"}},
		},
		{
			TestName:   "single set Source and *struct Target",
			Source:     &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfPtr(ctx, &TestFlexTF01{Field1: types.StringValue("a")})},
			Target:     &TestFlexAWS06{},
			WantTarget: &TestFlexAWS06{Field1: &TestFlexAWS01{Field1: "a"}},
		},
		{
			TestName:   "empty list Source and empty []struct Target",
			Source:     &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []TestFlexTF01{})},
			Target:     &TestFlexAWS08{},
			WantTarget: &TestFlexAWS08{Field1: []TestFlexAWS01{}},
		},
		{
			TestName: "non-empty list Source and non-empty []struct Target",
			Source: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []TestFlexTF01{
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
			Source:     &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfSlice(ctx, []*TestFlexTF01{})},
			Target:     &TestFlexAWS07{},
			WantTarget: &TestFlexAWS07{Field1: []*TestFlexAWS01{}},
		},
		{
			TestName: "non-empty list Source and non-empty []*struct Target",
			Source: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfSlice(ctx, []*TestFlexTF01{
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
			Source:     &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []TestFlexTF01{})},
			Target:     &TestFlexAWS08{},
			WantTarget: &TestFlexAWS08{Field1: []TestFlexAWS01{}},
		},
		{
			TestName: "non-empty list Source and non-empty []struct Target",
			Source: &TestFlexTF05{Field1: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []TestFlexTF01{
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
			Source:     &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfSlice(ctx, []*TestFlexTF01{})},
			Target:     &TestFlexAWS07{},
			WantTarget: &TestFlexAWS07{Field1: []*TestFlexAWS01{}},
		},
		{
			TestName: "non-empty set Source and non-empty []*struct Target",
			Source: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfSlice(ctx, []*TestFlexTF01{
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
			Source: &TestFlexTF06{Field1: fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, []TestFlexTF01{
				{Field1: types.StringValue("a")},
				{Field1: types.StringValue("b")},
			})},
			Target: &TestFlexAWS08{},
			WantTarget: &TestFlexAWS08{Field1: []TestFlexAWS01{
				{Field1: "a"},
				{Field1: "b"},
			}},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			err := Expand(ctx, testCase.Source, testCase.Target)
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
