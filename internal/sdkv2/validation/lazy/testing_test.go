// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lazyvalidation

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type testCase struct {
	val         any
	f           schema.SchemaValidateFunc
	expectedErr *regexp.Regexp
}

func runTestCases(t *testing.T, cases []testCase) {
	t.Helper()

	for i, tc := range cases {
		t.Run(fmt.Sprintf("TestCase_%d", i), func(t *testing.T) {
			t.Parallel()

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

func matchAnyError(errs []error, r *regexp.Regexp) bool {
	// err must match one provided
	for _, err := range errs {
		if r.MatchString(err.Error()) {
			return true
		}
	}
	return false
}

type diagTestCase struct {
	val                any
	f                  schema.SchemaValidateDiagFunc
	expectedDiagDetail *regexp.Regexp
}

func runDiagTestCases(t *testing.T, cases []diagTestCase) {
	t.Helper()

	for i, tc := range cases {
		t.Run(fmt.Sprintf("TestCase_%d", i), func(t *testing.T) {
			t.Parallel()

			diags := tc.f(tc.val, cty.GetAttrPath("test_property"))

			if len(diags) == 0 && tc.expectedDiagDetail == nil {
				return
			}

			if len(diags) != 0 && tc.expectedDiagDetail == nil {
				t.Fatalf("expected test case %d to produce no diagnostics, got %v", i, diags)
			}

			if !matchAnyDiagDetail(diags, tc.expectedDiagDetail) {
				t.Fatalf("expected test case %d to produce diagnostic detail matching \"%s\", got %v", i, tc.expectedDiagDetail, diags)
			}
		})
	}
}

func matchAnyDiagDetail(ds diag.Diagnostics, r *regexp.Regexp) bool {
	for _, d := range ds {
		if r.MatchString(d.Detail) {
			return true
		}
	}
	return false
}
