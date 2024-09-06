// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diff_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwdiff "github.com/hashicorp/terraform-provider-aws/internal/framework/diff"
)

type testResourceData1 struct {
	Name   types.String
	Number types.Int64
}

type testResourceData2 struct {
	Name types.String
}

func TestCalculate(t *testing.T) {
	testCases := map[string]struct {
		plan                      any
		state                     any
		expectedIgnoredFieldNames []string
		expectedChange            bool
		expectedErr               bool
	}{
		"no change": {
			plan:  testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(1)},
			state: testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(1)},
			expectedIgnoredFieldNames: []string{
				"Name",
				"Number",
			},
			expectedChange: false,
			expectedErr:    false,
		},
		"different struct types": {
			plan:        testResourceData2{Name: types.StringValue("test")},
			state:       testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(1)},
			expectedErr: true,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			results, diags := fwdiff.Calculate(context.Background(), test.plan, test.state)

			if diff := cmp.Diff(diags.HasError(), test.expectedErr); diff != "" {
				t.Fatalf("unexpected diff (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(results.HasChanges(), test.expectedChange); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(results.IgnoredFieldNames(), test.expectedIgnoredFieldNames); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}

			if len(results.FlexIgnoredFieldNames()) != len(test.expectedIgnoredFieldNames) {
				t.Errorf("unexpected length of FlexIgnoredFieldNames. got: %d, want: %d", len(results.FlexIgnoredFieldNames()), len(test.expectedIgnoredFieldNames))
			}
		})
	}

}
