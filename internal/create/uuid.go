// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package create

import (
	"context"
	"math/rand" // nosemgrep: go.lang.security.audit.crypto.math_random.math-random-used -- Deterministic PRNG required for VCR test reproducibility

	guuid "github.com/google/uuid"
	huuid "github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
)

// Helper for a resource to generate a UUID
//
// If context contains a VCR randomness source, it uses that for deterministic UUID
// generation. Otherwise, it falls back to the go-uuid library (generates a random UUID)
func UUID(ctx context.Context) string {
	if s, ok := vcr.FromContext(ctx); ok && s != nil {
		data := make([]byte, 16)
		rng := rand.New(s)
		rng.Read(data) //nolint:errcheck // rand.Rand.Read always returns a nil error
		return guuid.NewSHA1(guuid.NameSpaceOID, data).String()
	}

	uuid, _ := huuid.GenerateUUID() //nolint:errcheck // crypto/rand.Reader.Read returns nil on Go 1.24+ (or the program crashes)
	return uuid
}
