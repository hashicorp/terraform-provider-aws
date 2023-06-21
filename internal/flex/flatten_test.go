package flex

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
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
