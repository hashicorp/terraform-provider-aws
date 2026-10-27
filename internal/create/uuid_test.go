// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package create

import (
	"math/rand" // nosemgrep: go.lang.security.audit.crypto.math_random.math-random-used -- Deterministic PRNG required for VCR test reproducibility
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
)

var (
	uuidRegexp = regexache.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

func TestUUID(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	uuid1, uuid2 := UUID(ctx), UUID(ctx)
	if uuid1 == uuid2 {
		t.Fatal("UUIDs not unique")
	}

	if !uuidRegexp.MatchString(uuid1) {
		t.Errorf("UUID = %v, does not match regexp %s", uuid1, uuidRegexp)
	}
	if !uuidRegexp.MatchString(uuid2) {
		t.Errorf("UUID = %v, does not match regexp %s", uuid2, uuidRegexp)
	}
}

func TestUUID_VCR(t *testing.T) {
	t.Parallel()

	const (
		fixedSeed1 int64 = 12345678
		fixedSeed2 int64 = 23456789
	)
	testCases := []struct {
		testName       string
		source         rand.Source
		expectedRegexp *regexp.Regexp
		expected       string
	}{
		{
			testName:       "standard",
			expectedRegexp: uuidRegexp,
		},
		{
			testName: "go-vcr enabled (1)",
			source:   rand.NewSource(fixedSeed1),
			expected: "794d015e-63cd-51e4-b6e9-6151aebfb22e",
		},
		{
			testName: "go-vcr enabled (2)",
			source:   rand.NewSource(fixedSeed2),
			expected: "42011c4d-eb10-561e-8d46-6955b3ca6e8f",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			if testCase.source != nil {
				ctx = vcr.NewContext(ctx, testCase.source)
			}

			got := UUID(ctx)

			// Standard (regexp match)
			if testCase.expectedRegexp != nil && !testCase.expectedRegexp.MatchString(got) {
				t.Errorf("UUID = %v, does not match regexp %s", got, testCase.expectedRegexp)
			}
			// Go-VCR enabled (exact match)
			if testCase.expected != "" && testCase.expected != got {
				t.Errorf("UUID = %v, does not match %s", got, testCase.expected)
			}
		})
	}
}
