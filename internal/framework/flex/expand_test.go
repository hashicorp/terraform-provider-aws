// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type ATestExpand struct{}

type BTestExpand struct {
	Name types.String `tfsdk:"name"`
}

type CTestExpand struct {
	Name string
}

type DTestExpand struct {
	Name *string
}

type ETestExpand struct {
	Name types.Int64
}

type FTestExpand struct {
	Name int64
}

type GTestExpand struct {
	Name *int64
}

type HTestExpand struct {
	Name int32
}

type ITestExpand struct {
	Name *int32
}

type JTestExpand struct {
	Name types.Float64
}

type KTestExpand struct {
	Name float64
}

type LTestExpand struct {
	Name *float64
}

type MTestExpand struct {
	Name float32
}

type NTestExpand struct {
	Name *float32
}

type OTestExpand struct {
	Name types.Bool
}

type PTestExpand struct {
	Name bool
}

type QTestExpand struct {
	Name *bool
}

type RTestExpand struct {
	Names types.Set
}

type STestExpand struct {
	Names []string
}

type TTestExpand struct {
	Names []*string
}

type UTestExpand struct {
	Names types.List
}

type VTestExpand struct {
	Name fwtypes.Duration
}

type WTestExpand struct {
	Names types.Map
}

type XTestExpand struct {
	Names map[string]string
}

type YTestExpand struct {
	Names map[string]*string
}

type AATestExpand struct {
	Data fwtypes.ListNestedObjectValueOf[BTestExpand]
}

type BBTestExpand struct {
	Data *CTestExpand
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
			Source:   ATestExpand{},
			Target:   0,
			WantErr:  true,
		},
		{
			TestName: "non-struct Source",
			Source:   testString,
			Target:   &ATestExpand{},
			WantErr:  true,
		},
		{
			TestName: "non-struct Target",
			Source:   ATestExpand{},
			Target:   &testString,
			WantErr:  true,
		},
		{
			TestName:   "empty struct Source and Target",
			Source:     ATestExpand{},
			Target:     &ATestExpand{},
			WantTarget: &ATestExpand{},
		},
		{
			TestName:   "empty struct pointer Source and Target",
			Source:     &ATestExpand{},
			Target:     &ATestExpand{},
			WantTarget: &ATestExpand{},
		},
		{
			TestName:   "single string struct pointer Source and empty Target",
			Source:     &BTestExpand{Name: types.StringValue("a")},
			Target:     &ATestExpand{},
			WantTarget: &ATestExpand{},
		},
		{
			TestName: "does not implement attr.Value Source",
			Source:   &CTestExpand{Name: "a"},
			Target:   &CTestExpand{},
			WantErr:  true,
		},
		{
			TestName:   "single string Source and single string Target",
			Source:     &BTestExpand{Name: types.StringValue("a")},
			Target:     &CTestExpand{},
			WantTarget: &CTestExpand{Name: "a"},
		},
		{
			TestName:   "single string Source and single *string Target",
			Source:     &BTestExpand{Name: types.StringValue("a")},
			Target:     &DTestExpand{},
			WantTarget: &DTestExpand{Name: aws.String("a")},
		},
		{
			TestName: "single string Source and single int64 Target",
			Source:   &BTestExpand{Name: types.StringValue("a")},
			Target:   &FTestExpand{},
			WantErr:  true,
		},
		{
			TestName:   "single int64 Source and single int64 Target",
			Source:     &ETestExpand{Name: types.Int64Value(42)},
			Target:     &FTestExpand{},
			WantTarget: &FTestExpand{Name: 42},
		},
		{
			TestName:   "single int64 Source and single *int64 Target",
			Source:     &ETestExpand{Name: types.Int64Value(42)},
			Target:     &GTestExpand{},
			WantTarget: &GTestExpand{Name: aws.Int64(42)},
		},
		{
			TestName:   "single int64 Source and single int32 Target",
			Source:     &ETestExpand{Name: types.Int64Value(42)},
			Target:     &HTestExpand{},
			WantTarget: &HTestExpand{Name: 42},
		},
		{
			TestName:   "single int64 Source and single *int32 Target",
			Source:     &ETestExpand{Name: types.Int64Value(42)},
			Target:     &ITestExpand{},
			WantTarget: &ITestExpand{Name: aws.Int32(42)},
		},
		{
			TestName: "single int64 Source and single float64 Target",
			Source:   &ETestExpand{Name: types.Int64Value(42)},
			Target:   &KTestExpand{},
			WantErr:  true,
		},
		{
			TestName:   "single float64 Source and single float64 Target",
			Source:     &JTestExpand{Name: types.Float64Value(4.2)},
			Target:     &KTestExpand{},
			WantTarget: &KTestExpand{Name: 4.2},
		},
		{
			TestName:   "single float64 Source and single *float64 Target",
			Source:     &JTestExpand{Name: types.Float64Value(4.2)},
			Target:     &LTestExpand{},
			WantTarget: &LTestExpand{Name: aws.Float64(4.2)},
		},
		{
			TestName:   "single float64 Source and single float32 Target",
			Source:     &JTestExpand{Name: types.Float64Value(4.2)},
			Target:     &MTestExpand{},
			WantTarget: &MTestExpand{Name: 4.2},
		},
		{
			TestName:   "single float64 Source and single *float32 Target",
			Source:     &JTestExpand{Name: types.Float64Value(4.2)},
			Target:     &NTestExpand{},
			WantTarget: &NTestExpand{Name: aws.Float32(4.2)},
		},
		{
			TestName: "single float64 Source and single bool Target",
			Source:   &JTestExpand{Name: types.Float64Value(4.2)},
			Target:   &PTestExpand{},
			WantErr:  true,
		},
		{
			TestName:   "single bool Source and single bool Target",
			Source:     &OTestExpand{Name: types.BoolValue(true)},
			Target:     &PTestExpand{},
			WantTarget: &PTestExpand{Name: true},
		},
		{
			TestName:   "single bool Source and single *bool Target",
			Source:     &OTestExpand{Name: types.BoolValue(true)},
			Target:     &QTestExpand{},
			WantTarget: &QTestExpand{Name: aws.Bool(true)},
		},
		{
			TestName:   "single set Source and single string slice Target",
			Source:     &RTestExpand{Names: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")})},
			Target:     &STestExpand{},
			WantTarget: &STestExpand{Names: []string{"a"}},
		},
		{
			TestName:   "single set Source and single *string slice Target",
			Source:     &RTestExpand{Names: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")})},
			Target:     &TTestExpand{},
			WantTarget: &TTestExpand{Names: aws.StringSlice([]string{"a"})},
		},
		{
			TestName:   "single list Source and single string slice Target",
			Source:     &UTestExpand{Names: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")})},
			Target:     &STestExpand{},
			WantTarget: &STestExpand{Names: []string{"a"}},
		},
		{
			TestName:   "single list Source and single *string slice Target",
			Source:     &UTestExpand{Names: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")})},
			Target:     &TTestExpand{},
			WantTarget: &TTestExpand{Names: aws.StringSlice([]string{"a"})},
		},
		{
			TestName:   "single Duration Source and single string Target",
			Source:     &VTestExpand{Name: fwtypes.DurationValue(10 * time.Minute)},
			Target:     &CTestExpand{},
			WantTarget: &CTestExpand{Name: "10m0s"},
		},
		{
			TestName:   "single Duration Source and single *string Target",
			Source:     &VTestExpand{Name: fwtypes.DurationValue(10 * time.Minute)},
			Target:     &DTestExpand{},
			WantTarget: &DTestExpand{Name: aws.String("10m0s")},
		},
		{
			TestName:   "single map Source and single map[string]string slice Target",
			Source:     &WTestExpand{Names: types.MapValueMust(types.StringType, map[string]attr.Value{"A": types.StringValue("a")})},
			Target:     &XTestExpand{},
			WantTarget: &XTestExpand{Names: map[string]string{"A": "a"}},
		},
		{
			TestName:   "single map Source and single map[string]*string slice Target",
			Source:     &WTestExpand{Names: types.MapValueMust(types.StringType, map[string]attr.Value{"A": types.StringValue("a")})},
			Target:     &YTestExpand{},
			WantTarget: &YTestExpand{Names: aws.StringMap(map[string]string{"A": "a"})},
		},
		/*
			{
				TestName: "single list Source and single *struct Target",
				Source:   &AATestExpand{Data: fwtypes.NewListNestedObjectValueOfPtr(ctx, &BTestExpand{Name: types.StringValue("a")})},
				Target:   &BBTestExpand{},
				// WantTarget: &BBTestExpand{Data: &CTestExpand{Name: "a"}},
				WantTarget: &BBTestExpand{},
			},
		*/
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
