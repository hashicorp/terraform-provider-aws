// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"testing"
)

func TestIsNone(t *testing.T) {
	t.Parallel()

	none := None[int]()
	some := Some("yes")

	if got, want := none.IsNone(), true; got != want {
		t.Errorf("none.IsNone = %v, want = %v", got, want)
	}

	if got, want := some.IsNone(), false; got != want {
		t.Errorf("some.IsNone = %v, want = %v", got, want)
	}
}

func TestIsSome(t *testing.T) {
	t.Parallel()

	none := None[int]()
	some := Some("yes")

	if got, want := none.IsSome(), false; got != want {
		t.Errorf("none.IsSome = %v, want = %v", got, want)
	}

	if got, want := some.IsSome(), true; got != want {
		t.Errorf("some.IsSome = %v, want = %v", got, want)
	}
}

func TestUnwrapOr(t *testing.T) {
	t.Parallel()

	none := None[int]()
	some := Some("yes")

	if got, want := none.UnwrapOr(42), 42; got != want {
		t.Errorf("none.UnwrapOr = %v, want = %v", got, want)
	}

	if got, want := some.UnwrapOr("no"), "yes"; got != want {
		t.Errorf("some.UnwrapOr = %v, want = %v", got, want)
	}
}

func TestUnwrapOrDefault(t *testing.T) {
	t.Parallel()

	none := None[int]()
	some := Some("yes")

	if got, want := none.UnwrapOrDefault(), 0; got != want {
		t.Errorf("none.UnwrapOrDefault = %v, want = %v", got, want)
	}

	if got, want := some.UnwrapOrDefault(), "yes"; got != want {
		t.Errorf("some.UnwrapOrDefault = %v, want = %v", got, want)
	}
}

func TestUnwrapOrElse(t *testing.T) {
	t.Parallel()

	none := None[int]()
	some := Some("yes")

	if got, want := none.UnwrapOrElse(func() int { return 42 }), 42; got != want {
		t.Errorf("none.UnwrapOrElse = %v, want = %v", got, want)
	}

	if got, want := some.UnwrapOrElse(func() string { return "no" }), "yes"; got != want {
		t.Errorf("some.UnwrapOrElse = %v, want = %v", got, want)
	}
}
