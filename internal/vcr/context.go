// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package vcr

import (
	"context"
	"math/rand" // nosemgrep: go.lang.security.audit.crypto.math_random.math-random-used -- Deterministic PRNG required for VCR test reproducibility
)

type contextKeyType int

var contextKey contextKeyType

// NewContext returns a new context with the provided randomness source stored
//
// This is used to provide deterministic randomness for VCR test recording and replay.
func NewContext(ctx context.Context, source rand.Source) context.Context {
	return context.WithValue(ctx, contextKey, source)
}

// FromContext extracts the randomness source from the context, if present.
func FromContext(ctx context.Context) (rand.Source, bool) {
	source, ok := ctx.Value(contextKey).(rand.Source)
	return source, ok
}
