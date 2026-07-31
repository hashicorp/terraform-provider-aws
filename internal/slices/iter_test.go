// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package slices

import (
	"testing"
)

// Copied and adapted from stdlib slices package
func TestBackwardValues(t *testing.T) {
	t.Parallel()

	for size := range 10 {
		var s []int
		for i := range size {
			s = append(s, i)
		}
		ev := size - 1
		cnt := 0
		for v := range BackwardValues(s) {
			if v != ev {
				t.Errorf("at iteration %d got  %d want %d", cnt, v, ev)
			}
			ev--
			cnt++
		}
		if cnt != size {
			t.Errorf("read %d values expected %d", cnt, size)
		}
	}
}
