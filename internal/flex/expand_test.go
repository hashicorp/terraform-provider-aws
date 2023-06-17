package flex

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
			Source:   struct{}{},
			Target:   0,
			WantErr:  true,
		},
		{
			TestName: "non-struct Source",
			Source:   testString,
			Target:   &struct{}{},
			WantErr:  true,
		},
		{
			TestName: "non-struct Target",
			Source:   struct{}{},
			Target:   &testString,
			WantErr:  true,
		},
		{
			TestName:   "empty struct Source and Target",
			Source:     struct{}{},
			Target:     &struct{}{},
			WantTarget: &struct{}{},
		},
		{
			TestName:   "empty struct pointer Source and Target",
			Source:     &struct{}{},
			Target:     &struct{}{},
			WantTarget: &struct{}{},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			gotErr := expand(ctx, testCase.Source, testCase.Target) != nil

			if gotErr != testCase.WantErr {
				t.Errorf("gotErr = %v, wantErr = %v", gotErr, testCase.WantErr)
			}

			if !gotErr {
				if diff := cmp.Diff(testCase.Target, testCase.WantTarget); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}
