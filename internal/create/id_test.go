// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package create

import (
	mathrand "math/rand" // nosemgrep: go.lang.security.audit.crypto.math_random.math-random-used -- Deterministic PRNG required for VCR test reproducibility
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
)

// Ported from Terraform Plugin SDK V2
// https://github.com/hashicorp/terraform-plugin-sdk/blob/main/helper/id/id_test.go
func TestUniqueId(t *testing.T) {
	t.Parallel()

	var allB62 = regexache.MustCompile(`^[a-zA-Z0-9]+$`)

	ctx := t.Context()
	iterations := 10000
	ids := make(map[string]struct{})
	for range iterations {
		id := UniqueId(ctx)

		if _, ok := ids[id]; ok {
			t.Fatalf("Got duplicated id! %s", id)
		}

		if !strings.HasPrefix(id, UniqueIDPrefix) {
			t.Fatalf("Unique ID didn't have terraform- prefix! %s", id)
		}

		random := strings.TrimPrefix(id, UniqueIDPrefix)

		if len(random) != UniqueIDSuffixLength {
			t.Fatalf("UniqueId is out of sync with UniqueIDSuffixLength, post-prefix part has wrong length! %s", random)
		}

		if !allB62.MatchString(random) {
			t.Fatalf("Random part not all base62! %s", random)
		}

		ids[id] = struct{}{}
	}
}

// TestUniqueId_VCR verifies deterministic ID generation when go-vcr is enabled
func TestUniqueId_VCR(t *testing.T) {
	t.Parallel()

	const (
		fixedSeed1 int64 = 12345678
		fixedSeed2 int64 = 23456789
	)

	testCases := []struct {
		testName string
		seed     int64
		expected string
	}{
		{
			testName: "go-vcr enabled (1)",
			seed:     fixedSeed1,
			expected: "terraform-0000000000088b5ac78f2059ed",
		},
		{
			testName: "go-vcr enabled (2)",
			seed:     fixedSeed2,
			expected: "terraform-000000000045df53545befa95c",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			ctx = vcr.NewContext(ctx, mathrand.NewSource(testCase.seed))
			uniqueId := UniqueId(ctx)

			if testCase.expected != uniqueId {
				t.Errorf("UniqueId = %v, does not match %s", uniqueId, testCase.expected)
			}

			// test with a new source and the same seed to confirm it gives the same results
			ctx = vcr.NewContext(ctx, mathrand.NewSource(testCase.seed))
			uniqueId2 := UniqueId(ctx)
			if testCase.expected != uniqueId2 {
				t.Errorf("UniqueId (repeat) = %v, does not match %s", uniqueId2, testCase.expected)
			}
		})
	}
}
