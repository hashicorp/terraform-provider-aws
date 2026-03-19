// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"iter"

	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// Concat returns an iterator over the concatenation of the sequences.
func Concat[V any](seqs ...iter.Seq[V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, seq := range seqs {
			for e := range seq {
				if !yield(e) {
					return
				}
			}
		}
	}
}

// Concat returns an iterator over the concatenation of the values.
// The first non-nil error in seq is returned.
// If seq is empty, the result is nil.
func ConcatValuesWithError[E any](seq iter.Seq2[[]E, error]) iter.Seq2[E, error] {
	return func(yield func(E, error) bool) {
		for s, err := range seq {
			if err != nil {
				yield(inttypes.Zero[E](), err)
				return
			}

			for _, e := range s {
				if !yield(e, nil) {
					return
				}
			}
		}
	}
}
