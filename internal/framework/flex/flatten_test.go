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

// TODO Simplify by having more fields in structs.

type ATestFlatten struct{}

type BTestFlatten struct {
	Name string
}

type CTestFlatten struct {
	Name *string
}

type DTestFlatten struct {
	Name types.String `tfsdk:"name"`
}

type ETestFlatten struct {
	Name int64
}

type FTestFlatten struct {
	Name *int64
}

type GTestFlatten struct {
	Name types.Int64
}

type HTestFlatten struct {
	Name int32
}

type ITestFlatten struct {
	Name *int32
}

type JTestFlatten struct {
	Name float64
}

type KTestFlatten struct {
	Name *float64
}

type LTestFlatten struct {
	Name types.Float64
}

type MTestFlatten struct {
	Name float32
}

type NTestFlatten struct {
	Name *float32
}

type OTestFlatten struct {
	Name bool
}

type PTestFlatten struct {
	Name *bool
}

type QTestFlatten struct {
	Name types.Bool
}

type RTestFlatten struct {
	Names []string
}

type STestFlatten struct {
	Names []*string
}

type TTestFlatten struct {
	Names types.Set
}

type UTestFlatten struct {
	Names types.List
}

type VTestFlatten struct {
	Name fwtypes.Duration
}

type WTestFlatten struct {
	Names map[string]string
}

type XTestFlatten struct {
	Names map[string]*string
}

type YTestFlatten struct {
	Names types.Map
}

type AATestFlatten struct {
	Data *BTestFlatten
}

type BBTestFlatten struct {
	Data fwtypes.ListNestedObjectValueOf[DTestFlatten]
}

type CCTestFlatten struct {
	Data []BTestFlatten
}

type DDTestFlatten struct {
	Data []*BTestFlatten
}

type EETestFlatten struct {
	Data fwtypes.SetNestedObjectValueOf[DTestFlatten]
}

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
			Source:     &TestFlexAWS01{Name: "a"},
			Target:     &TestFlex00{},
			WantTarget: &TestFlex00{},
		},
		{
			TestName: "does not implement attr.Value Target",
			Source:   &TestFlexAWS01{Name: "a"},
			Target:   &TestFlexAWS01{},
			WantErr:  true,
		},
		{
			TestName:   "single empty string Source and single string Target",
			Source:     &TestFlexAWS01{},
			Target:     &TestFlexTF01{},
			WantTarget: &TestFlexTF01{Name: types.StringValue("")},
		},
		{
			TestName:   "single string Source and single string Target",
			Source:     &TestFlexAWS01{Name: "a"},
			Target:     &TestFlexTF01{},
			WantTarget: &TestFlexTF01{Name: types.StringValue("a")},
		},
		{
			TestName:   "single nil *string Source and single string Target",
			Source:     &TestFlexAWS02{},
			Target:     &TestFlexTF01{},
			WantTarget: &TestFlexTF01{Name: types.StringNull()},
		},
		{
			TestName:   "single *string Source and single string Target",
			Source:     &TestFlexAWS02{Name: aws.String("a")},
			Target:     &TestFlexTF01{},
			WantTarget: &TestFlexTF01{Name: types.StringValue("a")},
		},
		{
			TestName:   "single string Source and single int64 Target",
			Source:     &TestFlexAWS01{Name: "a"},
			Target:     &TestFlexTF02{},
			WantTarget: &TestFlexTF02{},
			WantErr:    true,
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
			TestName:   "single *struct Source and single list Target",
			Source:     &AATestFlatten{Data: &BTestFlatten{Name: "a"}},
			Target:     &BBTestFlatten{},
			WantTarget: &BBTestFlatten{Data: fwtypes.NewListNestedObjectValueOfPtr(ctx, &DTestFlatten{Name: types.StringValue("a")})},
		},
		{
			TestName:   "empty []struct and empty list Target",
			Source:     &CCTestFlatten{Data: []BTestFlatten{}},
			Target:     &BBTestFlatten{},
			WantTarget: &BBTestFlatten{Data: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []DTestFlatten{})},
		},
		{
			TestName: "non-empty []struct and non-empty list Target",
			Source: &CCTestFlatten{Data: []BTestFlatten{
				{Name: "a"},
				{Name: "b"},
			}},
			Target: &BBTestFlatten{},
			WantTarget: &BBTestFlatten{Data: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []DTestFlatten{
				{Name: types.StringValue("a")},
				{Name: types.StringValue("b")},
			})},
		},
		{
			TestName:   "empty []*struct and empty list Target",
			Source:     &DDTestFlatten{Data: []*BTestFlatten{}},
			Target:     &BBTestFlatten{},
			WantTarget: &BBTestFlatten{Data: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []DTestFlatten{})},
		},
		{
			TestName: "non-empty []*struct and non-empty list Target",
			Source: &DDTestFlatten{Data: []*BTestFlatten{
				{Name: "a"},
				{Name: "b"},
			}},
			Target: &BBTestFlatten{},
			WantTarget: &BBTestFlatten{Data: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []DTestFlatten{
				{Name: types.StringValue("a")},
				{Name: types.StringValue("b")},
			})},
		},
		{
			TestName: "non-empty []*struct and non-empty set Target",
			Source: &DDTestFlatten{Data: []*BTestFlatten{
				{Name: "a"},
				{Name: "b"},
			}},
			Target: &EETestFlatten{},
			WantTarget: &EETestFlatten{Data: fwtypes.NewSetNestedObjectValueOfValueSlice(ctx, []DTestFlatten{
				{Name: types.StringValue("a")},
				{Name: types.StringValue("b")},
			})},
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
