package flex

import (
	"context"
	"testing"

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
