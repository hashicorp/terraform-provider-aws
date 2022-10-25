package provider

import (
	"regexp"
	"testing"
)

func TestValidAssumeRoleDuration(t *testing.T) {
	testCases := []struct {
		val         interface{}
		expectedErr *regexp.Regexp
	}{
		{
			val:         "",
			expectedErr: regexp.MustCompile(`cannot be parsed as a duration`),
		},
		{
			val:         "1",
			expectedErr: regexp.MustCompile(`cannot be parsed as a duration`),
		},
		{
			val:         "10m",
			expectedErr: regexp.MustCompile(`must be between 15 minutes \(15m\) and 12 hours \(12h\)`),
		},
		{
			val:         "12h30m",
			expectedErr: regexp.MustCompile(`must be between 15 minutes \(15m\) and 12 hours \(12h\)`),
		},
		{

			val: "15m",
		},
		{
			val: "1h10m10s",
		},
		{

			val: "12h",
		},
	}

	matchErr := func(errs []error, r *regexp.Regexp) bool {
		// err must match one provided
		for _, err := range errs {
			if r.MatchString(err.Error()) {
				return true
			}
		}

		return false
	}

	for i, tc := range testCases {
		_, errs := validAssumeRoleDuration(tc.val, "test_property")

		if len(errs) == 0 && tc.expectedErr == nil {
			continue
		}

		if len(errs) != 0 && tc.expectedErr == nil {
			t.Fatalf("expected test case %d to produce no errors, got %v", i, errs)
		}

		if !matchErr(errs, tc.expectedErr) {
			t.Fatalf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, errs)
		}
	}
}
