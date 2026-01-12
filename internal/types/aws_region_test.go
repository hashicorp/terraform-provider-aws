// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"testing"

	"github.com/YakDriver/regexache"
)

func TestCanonicalRegionPatterns(t *testing.T) {
	t.Parallel()

	// Test regions
	testRegions := []struct {
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

	patterns := map[string]string{
		"Canonical":           CanonicalRegionPatternNoAnchors,
		"Lambda functionName": CanonicalRegionPatternNoAnchors,    // Now uses canonical pattern
		"Lambda permission":   `[a-z]{2,4}-(?:[a-z]+-){1,2}\d{1}`, // Already worked
	}

	for patternName, pattern := range patterns {
		patternName, pattern := patternName, pattern
		t.Run(patternName, func(t *testing.T) {
			t.Parallel()
			regex := regexache.MustCompile("^" + pattern + "$")

			for _, test := range testRegions {
				result := regex.MatchString(test.region)
				if result != test.valid {
					t.Errorf("Pattern %s: region %s expected %t, got %t",
						patternName, test.region, test.valid, result)
				}
			}
		})
	}
}
