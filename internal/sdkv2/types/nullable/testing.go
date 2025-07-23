// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package nullable

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type testCase struct {
	val         any
	f           schema.SchemaValidateFunc
	expectedErr *regexp.Regexp
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

	for i, tc := range cases {
		_, errs := tc.f(tc.val, "test_property")

		if tc.expectedErr == nil && len(errs) != 0 {
			t.Errorf("expected test case %d to produce no errors, got %v", i, errs)
		}
		if tc.expectedErr != nil && !matchErr(errs, tc.expectedErr) {
			t.Errorf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, errs)
		}
	}
}
