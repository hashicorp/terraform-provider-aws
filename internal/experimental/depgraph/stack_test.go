// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package depgraph

import (
	"testing"
)

func TestStack(t *testing.T) {
	t.Parallel()

	s := newStack()

	if got, expected := s.len(), 0; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	if got, expected := s.peek(), interface{}(nil); got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}

	if got, expected := s.pop(), interface{}(nil); got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}

	s.push(1)

	if got, expected := s.len(), 1; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	if got, expected := s.peek(), interface{}(1); got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}

	if got, expected := s.pop(), interface{}(1); got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}

	if got, expected := s.len(), 0; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	s.push(2)
	s.push(3)

	if got, expected := s.len(), 2; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	if got, expected := s.peek(), interface{}(3); got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}

	if got, expected := s.pop(), interface{}(3); got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}

	if got, expected := s.peek(), interface{}(2); got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}
}
