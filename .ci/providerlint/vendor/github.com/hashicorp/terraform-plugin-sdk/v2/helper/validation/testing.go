// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/go-cty/cty"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type testCase struct {
	val         interface{}
	f           schema.SchemaValidateFunc
	expectedErr *regexp.Regexp
}

func runTestCases(t *testing.T, cases []testCase) {
	t.Helper()

	for i, tc := range cases {
		t.Run(fmt.Sprintf("TestCase_%d", i), func(t *testing.T) {
			_, errs := tc.f(tc.val, "test_property")

			if len(errs) == 0 && tc.expectedErr == nil {
				return
			}

			if len(errs) != 0 && tc.expectedErr == nil {
				t.Fatalf("expected test case %d to produce no errors, got %v", i, errs)
			}

			if !matchAnyError(errs, tc.expectedErr) {
				t.Fatalf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, errs)
			}
		})
	}
}

type diagTestCase struct {
	val                 interface{}
	f                   schema.SchemaValidateDiagFunc
	expectedDiagSummary *regexp.Regexp
}

func runDiagTestCases(t *testing.T, cases []diagTestCase) {
	t.Helper()

	for i, tc := range cases {
		t.Run(fmt.Sprintf("TestCase_%d", i), func(t *testing.T) {
			diags := tc.f(tc.val, cty.GetAttrPath("test_property"))

			if len(diags) == 0 && tc.expectedDiagSummary == nil {
				return
			}

			if len(diags) != 0 && tc.expectedDiagSummary == nil {
				t.Fatalf("expected test case %d to produce no diagnostics, got %v", i, diags)
			}

			if !matchAnyDiagSummary(diags, tc.expectedDiagSummary) {
				t.Fatalf("expected test case %d to produce diagnostic summary matching \"%s\", got %v", i, tc.expectedDiagSummary, diags)
			}
		})
	}
}

func matchAnyError(errs []error, r *regexp.Regexp) bool {
	// err must match one provided
	for _, err := range errs {
		if r.MatchString(err.Error()) {
			return true
		}
	}
	return false
}

func matchAnyDiagSummary(ds diag.Diagnostics, r *regexp.Regexp) bool {
	for _, d := range ds {
		if r.MatchString(d.Summary) {
			return true
		}
	}
	return false
}
