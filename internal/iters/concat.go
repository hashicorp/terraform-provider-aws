// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iters

import (
	"iter"
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
