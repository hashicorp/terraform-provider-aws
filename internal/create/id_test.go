// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package create

import (
	"fmt"
	"math/rand" // nosemgrep: go.lang.security.audit.crypto.math_random.math-random-used -- Deterministic PRNG required for VCR test reproducibility
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
)

func TestUniqueId(t *testing.T) {
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
			expectedRegexp: regexache.MustCompile(fmt.Sprintf("^terraform-[[:xdigit:]]{%d}$", sdkid.UniqueIDSuffixLength)),
		},
		{
			testName: "go-vcr enabled (1)",
			source:   rand.NewSource(fixedSeed1),
			expected: "terraform-0000000000088b5ac78f2059ed",
		},
		{
			testName: "go-vcr enabled (2)",
			source:   rand.NewSource(fixedSeed2),
			expected: "terraform-000000000045df53545befa95c",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			if testCase.source != nil {
				ctx = vcr.NewContext(ctx, testCase.source)
			}

			got := UniqueId(ctx)

			// Standard (regexp match)
			if testCase.expectedRegexp != nil && !testCase.expectedRegexp.MatchString(got) {
				t.Errorf("UniqueId = %v, does not match regexp %s", got, testCase.expectedRegexp)
			}
			// Go-VCR enabled (exact match)
			if testCase.expected != "" && testCase.expected != got {
				t.Errorf("UniqueId = %v, does not match %s", got, testCase.expected)
			}
		})
	}
}
