// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

type testResourceData1 struct {
	Name   types.String
	Number types.Int64
	Age    types.Int64
}

type testResourceData2 struct {
	Name types.String
}

type testResourceData3 struct {
	testResourceData2
	Number types.Int64
	Age    types.Int64
}

func TestCalculate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		plan                      any
		state                     any
		expectedIgnoredFieldNames []string
		expectedChange            bool
		expectErr                 bool
	}{
		"no change": {
			plan:  testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(1), Age: types.Int64Value(100)},
			state: testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(1), Age: types.Int64Value(100)},
			expectedIgnoredFieldNames: []string{
				"Name",
				"Number",
				"Age",
			},
			expectedChange: false,
			expectErr:      false,
		},
		"different struct types": {
			plan:      testResourceData2{Name: types.StringValue("test")},
			state:     testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(1)},
			expectErr: true,
		},
		"has change plan": {
			plan:  testResourceData1{Name: types.StringValue("testChanged"), Number: types.Int64Value(1)},
			state: testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(1)},
			expectedIgnoredFieldNames: []string{
				"Number",
				"Age",
			},
			expectedChange: true,
			expectErr:      false,
		},
		"has change state": {
			plan:  testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(1)},
			state: testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(2)},
			expectedIgnoredFieldNames: []string{
				"Name",
				"Age",
			},
			expectedChange: true,
			expectErr:      false,
		},
		"has multiple changes": {
			plan:  testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(1), Age: types.Int64Value(100)},
			state: testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(2), Age: types.Int64Value(200)},
			expectedIgnoredFieldNames: []string{
				"Name",
			},
			expectedChange: true,
			expectErr:      false,
		},

		"embedded no change": {
			plan:  testResourceData3{testResourceData2: testResourceData2{Name: types.StringValue("test")}, Number: types.Int64Value(1), Age: types.Int64Value(100)},
			state: testResourceData3{testResourceData2: testResourceData2{Name: types.StringValue("test")}, Number: types.Int64Value(1), Age: types.Int64Value(100)},
			expectedIgnoredFieldNames: []string{
				"Name",
				"Number",
				"Age",
			},
			expectedChange: false,
			expectErr:      false,
		},
		"embedded different struct types": {
			plan:      testResourceData3{testResourceData2: testResourceData2{Name: types.StringValue("test")}, Number: types.Int64Value(1)},
			state:     testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(1)},
			expectErr: true,
		},
		"embedded has multiple changes": {
			plan:  testResourceData3{testResourceData2: testResourceData2{Name: types.StringValue("test")}, Number: types.Int64Value(1), Age: types.Int64Value(100)},
			state: testResourceData3{testResourceData2: testResourceData2{Name: types.StringValue("test")}, Number: types.Int64Value(2), Age: types.Int64Value(200)},
			expectedIgnoredFieldNames: []string{
				"Name",
			},
			expectedChange: true,
			expectErr:      false,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			results, diags := fwflex.Diff(context.Background(), test.plan, test.state)

			if diff := cmp.Diff(diags.HasError(), test.expectErr); diff != "" {
				t.Fatalf("unexpected diff (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(results.HasChanges(), test.expectedChange); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(results.IgnoredFieldNames(), test.expectedIgnoredFieldNames); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}

			if len(results.IgnoredFieldNamesOpts()) != len(test.expectedIgnoredFieldNames) {
				t.Errorf("unexpected length of FlexIgnoredFieldNames. got: %d, want: %d", len(results.IgnoredFieldNamesOpts()), len(test.expectedIgnoredFieldNames))
			}
		})
	}
}

func TestWithException(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		plan                      any
		state                     any
		withException             []fwflex.ChangeOption
		expectedIgnoredFieldNames []string
	}{
		"ignore changed field": {
			plan:          testResourceData1{Name: types.StringValue("test2"), Number: types.Int64Value(1), Age: types.Int64Value(100)},
			state:         testResourceData1{Name: types.StringValue("test"), Number: types.Int64Value(1), Age: types.Int64Value(100)},
			withException: []fwflex.ChangeOption{fwflex.WithIgnoredField("Name")},
			expectedIgnoredFieldNames: []string{
				"Name",
				"Number",
				"Age",
			},
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			results, _ := fwflex.Diff(context.Background(), test.plan, test.state, test.withException...)

			if diff := cmp.Diff(results.IgnoredFieldNames(), test.expectedIgnoredFieldNames); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}

			if len(results.IgnoredFieldNamesOpts()) != len(test.expectedIgnoredFieldNames) {
				t.Errorf("unexpected length of FlexIgnoredFieldNames. got: %d, want: %d", len(results.IgnoredFieldNamesOpts()), len(test.expectedIgnoredFieldNames))
			}
		})
	}
}
