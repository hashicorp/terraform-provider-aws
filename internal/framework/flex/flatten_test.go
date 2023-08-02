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
)

type ATestFlatten struct{}

type BTestFlatten struct {
	Name string
}

type CTestFlatten struct {
	Name *string
}

type DTestFlatten struct {
	Name types.String
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
			Source:   ATestFlatten{},
			Target:   0,
			WantErr:  true,
		},
		{
			TestName: "non-struct Source",
			Source:   testString,
			Target:   &ATestFlatten{},
			WantErr:  true,
		},
		{
			TestName: "non-struct Target",
			Source:   ATestFlatten{},
			Target:   &testString,
			WantErr:  true,
		},
		{
			TestName:   "empty struct Source and Target",
			Source:     ATestFlatten{},
			Target:     &ATestFlatten{},
			WantTarget: &ATestFlatten{},
		},
		{
			TestName:   "empty struct pointer Source and Target",
			Source:     &ATestFlatten{},
			Target:     &ATestFlatten{},
			WantTarget: &ATestFlatten{},
		},
		{
			TestName:   "single string struct pointer Source and empty Target",
			Source:     &BTestFlatten{Name: "a"},
			Target:     &ATestFlatten{},
			WantTarget: &ATestFlatten{},
		},
		{
			TestName: "does not implement attr.Value Target",
			Source:   &BTestFlatten{Name: "a"},
			Target:   &CTestFlatten{},
			WantErr:  true,
		},
		{
			TestName:   "single empty string Source and single string Target",
			Source:     &BTestFlatten{},
			Target:     &DTestFlatten{},
			WantTarget: &DTestFlatten{Name: types.StringValue("")},
		},
		{
			TestName:   "single string Source and single string Target",
			Source:     &BTestFlatten{Name: "a"},
			Target:     &DTestFlatten{},
			WantTarget: &DTestFlatten{Name: types.StringValue("a")},
		},
		{
			TestName:   "single nil *string Source and single string Target",
			Source:     &CTestFlatten{},
			Target:     &DTestFlatten{},
			WantTarget: &DTestFlatten{Name: types.StringNull()},
		},
		{
			TestName:   "single *string Source and single string Target",
			Source:     &CTestFlatten{Name: aws.String("a")},
			Target:     &DTestFlatten{},
			WantTarget: &DTestFlatten{Name: types.StringValue("a")},
		},
		{
			TestName: "single string Source and single int64 Target",
			Source:   &BTestFlatten{Name: "a"},
			Target:   &GTestFlatten{},
			WantErr:  true,
		},
		{
			TestName:   "single int64 Source and single int64 Target",
			Source:     &ETestFlatten{Name: 42},
			Target:     &GTestFlatten{},
			WantTarget: &GTestFlatten{Name: types.Int64Value(42)},
		},
		{
			TestName:   "single *int64 Source and single int64 Target",
			Source:     &FTestFlatten{Name: aws.Int64(42)},
			Target:     &GTestFlatten{},
			WantTarget: &GTestFlatten{Name: types.Int64Value(42)},
		},
		{
			TestName:   "single int32 Source and single int64 Target",
			Source:     &HTestFlatten{Name: 42},
			Target:     &GTestFlatten{},
			WantTarget: &GTestFlatten{Name: types.Int64Value(42)},
		},
		{
			TestName:   "single *int32 Source and single int64 Target",
			Source:     &ITestFlatten{Name: aws.Int32(42)},
			Target:     &GTestFlatten{},
			WantTarget: &GTestFlatten{Name: types.Int64Value(42)},
		},
		{
			TestName: "single *int32 Source and single string Target",
			Source:   &ITestFlatten{Name: aws.Int32(42)},
			Target:   &DTestFlatten{},
			WantErr:  true,
		},
		{
			TestName:   "single float64 Source and single float64 Target",
			Source:     &JTestFlatten{Name: 4.2},
			Target:     &LTestFlatten{},
			WantTarget: &LTestFlatten{Name: types.Float64Value(4.2)},
		},
		{
			TestName:   "single *float64 Source and single float64 Target",
			Source:     &KTestFlatten{Name: aws.Float64(4.2)},
			Target:     &LTestFlatten{},
			WantTarget: &LTestFlatten{Name: types.Float64Value(4.2)},
		},
		{
			TestName:   "single float32 Source and single float64 Target",
			Source:     &MTestFlatten{Name: 4},
			Target:     &LTestFlatten{},
			WantTarget: &LTestFlatten{Name: types.Float64Value(4)},
		},
		{
			TestName:   "single *float32 Source and single float64 Target",
			Source:     &NTestFlatten{Name: aws.Float32(4)},
			Target:     &LTestFlatten{},
			WantTarget: &LTestFlatten{Name: types.Float64Value(4)},
		},
		{
			TestName:   "single nil *float32 Source and single float64 Target",
			Source:     &NTestFlatten{},
			Target:     &LTestFlatten{},
			WantTarget: &LTestFlatten{Name: types.Float64Null()},
		},
		{
			TestName:   "single bool Source and single bool Target",
			Source:     &OTestFlatten{Name: true},
			Target:     &QTestFlatten{},
			WantTarget: &QTestFlatten{Name: types.BoolValue(true)},
		},
		{
			TestName:   "single *bool Source and single bool Target",
			Source:     &PTestFlatten{Name: aws.Bool(true)},
			Target:     &QTestFlatten{},
			WantTarget: &QTestFlatten{Name: types.BoolValue(true)},
		},
		{
			TestName:   "single string slice Source and single set Target",
			Source:     &RTestFlatten{Names: []string{"a"}},
			Target:     &TTestFlatten{},
			WantTarget: &TTestFlatten{Names: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")})},
		},
		{
			TestName:   "single nil string slice Source and single set Target",
			Source:     &RTestFlatten{},
			Target:     &TTestFlatten{},
			WantTarget: &TTestFlatten{Names: types.SetNull(types.StringType)},
		},
		{
			TestName:   "single *string slice Source and single set Target",
			Source:     &STestFlatten{Names: aws.StringSlice([]string{"a"})},
			Target:     &TTestFlatten{},
			WantTarget: &TTestFlatten{Names: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("a")})},
		},
		{
			TestName:   "single string slice Source and single list Target",
			Source:     &RTestFlatten{Names: []string{"a"}},
			Target:     &UTestFlatten{},
			WantTarget: &UTestFlatten{Names: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")})},
		},
		{
			TestName:   "single *string slice Source and single list Target",
			Source:     &STestFlatten{Names: aws.StringSlice([]string{"a"})},
			Target:     &UTestFlatten{},
			WantTarget: &UTestFlatten{Names: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")})},
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
