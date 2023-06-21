package flex

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ATest struct{}

type BTest struct {
	Name types.String
}

type CTest struct {
	Name string
}

type DTest struct {
	Name *string
}

type ETest struct {
	Name types.Int64
}

type FTest struct {
	Name int64
}

type GTest struct {
	Name *int64
}

type HTest struct {
	Name int32
}

type ITest struct {
	Name *int32
}

type JTest struct {
	Name types.Float64
}

type KTest struct {
	Name float64
}

type LTest struct {
	Name *float64
}

type MTest struct {
	Name float32
}

type NTest struct {
	Name *float32
}

type OTest struct {
	Name types.Bool
}

type PTest struct {
	Name bool
}

type QTest struct {
	Name *bool
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
			Source:   ATest{},
			Target:   0,
			WantErr:  true,
		},
		{
			TestName: "non-struct Source",
			Source:   testString,
			Target:   &ATest{},
			WantErr:  true,
		},
		{
			TestName: "non-struct Target",
			Source:   ATest{},
			Target:   &testString,
			WantErr:  true,
		},
		{
			TestName:   "empty struct Source and Target",
			Source:     ATest{},
			Target:     &ATest{},
			WantTarget: &ATest{},
		},
		{
			TestName:   "empty struct pointer Source and Target",
			Source:     &ATest{},
			Target:     &ATest{},
			WantTarget: &ATest{},
		},
		{
			TestName:   "single string struct pointer Source and empty Target",
			Source:     &BTest{Name: types.StringValue("a")},
			Target:     &ATest{},
			WantTarget: &ATest{},
		},
		{
			TestName: "does not implement attr.Value Source",
			Source:   &CTest{Name: "a"},
			Target:   &CTest{},
			WantErr:  true,
		},
		{
			TestName:   "single string Source and single string Target",
			Source:     &BTest{Name: types.StringValue("a")},
			Target:     &CTest{},
			WantTarget: &CTest{Name: "a"},
		},
		{
			TestName:   "single string Source and single *string Target",
			Source:     &BTest{Name: types.StringValue("a")},
			Target:     &DTest{},
			WantTarget: &DTest{Name: aws.String("a")},
		},
		{
			TestName: "single string Source and single int64 Target",
			Source:   &BTest{Name: types.StringValue("a")},
			Target:   &FTest{},
			WantErr:  true,
		},
		{
			TestName:   "single int64 Source and single int64 Target",
			Source:     &ETest{Name: types.Int64Value(42)},
			Target:     &FTest{},
			WantTarget: &FTest{Name: 42},
		},
		{
			TestName:   "single int64 Source and single *int64 Target",
			Source:     &ETest{Name: types.Int64Value(42)},
			Target:     &GTest{},
			WantTarget: &GTest{Name: aws.Int64(42)},
		},
		{
			TestName:   "single int64 Source and single int32 Target",
			Source:     &ETest{Name: types.Int64Value(42)},
			Target:     &HTest{},
			WantTarget: &HTest{Name: 42},
		},
		{
			TestName:   "single int64 Source and single *int32 Target",
			Source:     &ETest{Name: types.Int64Value(42)},
			Target:     &ITest{},
			WantTarget: &ITest{Name: aws.Int32(42)},
		},
		{
			TestName: "single int64 Source and single float64 Target",
			Source:   &ETest{Name: types.Int64Value(42)},
			Target:   &KTest{},
			WantErr:  true,
		},
		{
			TestName:   "single float64 Source and single float64 Target",
			Source:     &JTest{Name: types.Float64Value(4.2)},
			Target:     &KTest{},
			WantTarget: &KTest{Name: 4.2},
		},
		{
			TestName:   "single float64 Source and single *float64 Target",
			Source:     &JTest{Name: types.Float64Value(4.2)},
			Target:     &LTest{},
			WantTarget: &LTest{Name: aws.Float64(4.2)},
		},
		{
			TestName:   "single float64 Source and single float32 Target",
			Source:     &JTest{Name: types.Float64Value(4.2)},
			Target:     &MTest{},
			WantTarget: &MTest{Name: 4.2},
		},
		{
			TestName:   "single float64 Source and single *float32 Target",
			Source:     &JTest{Name: types.Float64Value(4.2)},
			Target:     &NTest{},
			WantTarget: &NTest{Name: aws.Float32(4.2)},
		},
		{
			TestName: "single float64 Source and single bool Target",
			Source:   &JTest{Name: types.Float64Value(4.2)},
			Target:   &PTest{},
			WantErr:  true,
		},
		{
			TestName:   "single bool Source and single bool Target",
			Source:     &OTest{Name: types.BoolValue(true)},
			Target:     &PTest{},
			WantTarget: &PTest{Name: true},
		},
		{
			TestName:   "single bool Source and single *bool Target",
			Source:     &OTest{Name: types.BoolValue(true)},
			Target:     &QTest{},
			WantTarget: &QTest{Name: aws.Bool(true)},
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
