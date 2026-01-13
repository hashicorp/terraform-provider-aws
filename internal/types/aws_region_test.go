// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"testing"

	"github.com/YakDriver/regexache"
)

func TestCanonicalRegionPatterns(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		region string
		valid  bool
	}{
		// Standard regions
		{"us-east-1", true},
		{"eu-west-1", true},
		{"ap-southeast-1", true},

		// Government regions
		{"us-gov-west-1", true},
		{"us-gov-east-1", true},

		// ISO regions
		{"us-iso-east-1", true},
		{"us-isob-east-1", true},

		// ESC regions
		{"eusc-de-east-1", true},

		// Invalid regions
		{"invalid-region", false},
		{"us-east", false},
		{"1-east-1", false},
		{"us-east-1a", false},
	}

	regex := regexache.MustCompile(CanonicalRegionPattern)

	for _, tc := range testCases {
		t.Run(tc.region, func(t *testing.T) {
			result := regex.MatchString(tc.region)
			if result != tc.valid {
				t.Errorf("region %s: expected %t, got %t", tc.region, tc.valid, result)
			}
		})
	}
}
