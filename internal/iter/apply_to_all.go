// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"iter"
)

// ApplyToAll returns a new iterator yielding the results of applying the function `f` to each element of the original iterator `seq`.
func ApplyToAll[E1, E2 any](seq iter.Seq[E1], f func(E1) E2) iter.Seq[E2] {
	return func(yield func(E2) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}
