// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package create

import (
	"context"
	"fmt"
	"math/rand" // nosemgrep: go.lang.security.audit.crypto.math_random.math-random-used -- Deterministic PRNG required for VCR test reproducibility

	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
)

// Helper for a resource to generate a unique identifier with default prefix
//
// This is a VCR-aware variant of the Plugin SDK V2 id.UniqueId function.
// If context contains a VCR randomness source, it uses that for deterministic ID
// generation. Otherwise, it falls back to id.UniqueId.
func UniqueId(ctx context.Context) string {
	return prefixedUniqueId(ctx, sdkid.UniqueIdPrefix)
}

// prefixedUniqueId generates a unique ID with the given prefix
//
// This is a VCR-aware variant of the Plugin SDK V2 id.PrefixUniqueId function.
// If context contains a VCR randomness source, it uses that for deterministic ID
// generation. Otherwise, it falls back to id.PrefixedUniqueId.
func prefixedUniqueId(ctx context.Context, prefix string) string {
	if s, ok := vcr.FromContext(ctx); ok && s != nil {
		rng := rand.New(s)
		// Pad the generated int64 to match the length of the id.PrefixUniqueId (26 characters)
		return fmt.Sprintf("%s%026x", prefix, rng.Int63())
	}
	return sdkid.PrefixedUniqueId(prefix)
}
