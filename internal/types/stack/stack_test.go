// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package stack

import (
	"testing"
)

func TestStack(t *testing.T) {
	t.Parallel()

	s := New[int]()

	if got, expected := s.Len(), 0; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	if got, expected := s.Peek().IsNone(), true; got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}

	if got, expected := s.Pop().IsSome(), false; got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}

	s.Push(1)

	if got, expected := s.Len(), 1; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	if got, expected := s.Peek().MustUnwrap(), 1; got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}

	if got, expected := s.Pop().MustUnwrap(), 1; got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}

	if got, expected := s.Len(), 0; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	s.Push(2)
	s.Push(3)

	if got, expected := s.Len(), 2; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	if got, expected := s.Peek().MustUnwrap(), 3; got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}

	if got, expected := s.Pop().MustUnwrap(), 3; got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}

	if got, expected := s.Peek().MustUnwrap(), 2; got != expected {
		t.Fatalf("incorrect value. Expected: %v, got: %v", expected, got)
	}
}
