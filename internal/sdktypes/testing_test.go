// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdktypes

import (
	"regexp"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type testCase struct {
	val             interface{}
	f               schema.SchemaValidateDiagFunc
	expectedSummary *regexp.Regexp
	expectedDetail  *regexp.Regexp
}

func runTestCases(t *testing.T, cases map[string]testCase) {
	t.Helper()

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.f(tc.val, cty.Path{cty.GetAttrStep{Name: "test_property"}})

			if !diags.HasError() && tc.expectedSummary == nil && tc.expectedDetail == nil {
				return
			}

			if diags.HasError() && tc.expectedSummary == nil && tc.expectedDetail == nil {
				t.Fatalf("expected no errors, got %v", diags)
			}

			for _, d := range diags {
				if d.Severity != diag.Error {
					continue
				}
				if tc.expectedSummary != nil && !tc.expectedSummary.MatchString(d.Summary) {
					t.Errorf("expected error with summary matching \"%s\", got %#v", tc.expectedSummary, diags)
				}
				if tc.expectedDetail != nil && !tc.expectedDetail.MatchString(d.Detail) {
					t.Errorf("expected error with detail matching \"%s\", got %#v", tc.expectedDetail, diags)
				}
			}
		})
	}
}
