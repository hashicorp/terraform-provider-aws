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
