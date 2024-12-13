// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package nullable

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type testCase struct {
	val             interface{}
	f               schema.SchemaValidateFunc
	expectedErr     *regexp.Regexp
	expectedWarning *regexp.Regexp
}

func runValidationTestCases(t *testing.T, cases []testCase) {
	t.Helper()

	matchErr := func(errs []error, r *regexp.Regexp) bool {
		// err must match one provided
		for _, err := range errs {
			if r.MatchString(err.Error()) {
				return true
			}
		}

		return false
	}

	matchWarning := func(warnings []string, r *regexp.Regexp) bool {
		// warning must match one provided
		for _, warning := range warnings {
			if r.MatchString(warning) {
				return true
			}
		}

		return false
	}

	for i, tc := range cases {
		warnings, errs := tc.f(tc.val, "test_property")

		if tc.expectedErr == nil && len(errs) != 0 {
			t.Errorf("expected test case %d to produce no errors, got %v", i, errs)
		}
		if tc.expectedErr != nil && !matchErr(errs, tc.expectedErr) {
			t.Errorf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, errs)
		}

		if tc.expectedWarning == nil && len(warnings) != 0 {
			t.Errorf("expected test case %d to produce no warnings, got %v", i, warnings)
		}
		if tc.expectedWarning != nil && !matchWarning(warnings, tc.expectedWarning) {
			t.Errorf("expected test case %d to produce warning matching \"%s\", got %v", i, tc.expectedWarning, warnings)
		}
	}
}
