// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func TestStringUTF8LenBetween(t *testing.T) {
	t.Parallel()

	t.Run("valid ASCII length", func(t *testing.T) {
		t.Parallel()

		diags := StringUTF8LenBetween(2, 4)("test", cty.Path{})

		if len(diags) != 0 {
			t.Fatalf("expected no diagnostics, got %#v", diags)
		}
	})

	t.Run("valid multibyte UTF-8 length", func(t *testing.T) {
		t.Parallel()

		diags := StringUTF8LenBetween(2, 4)("あい", cty.Path{})

		if len(diags) != 0 {
			t.Fatalf("expected no diagnostics, got %#v", diags)
		}
	})

	t.Run("too short", func(t *testing.T) {
		t.Parallel()

		diags := StringUTF8LenBetween(2, 4)("a", cty.Path{})

		if len(diags) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diags)
		}
		if diags[0].Severity != diag.Error {
			t.Fatalf("expected severity %q, got %q", diag.Error, diags[0].Severity)
		}
		if got, want := diags[0].Summary, "Invalid character length"; got != want {
			t.Fatalf("expected summary %q, got %q", want, got)
		}
		if got, want := diags[0].Detail, "expected length of pattern to be between 2 and 4 UTF-8 characters, got 1"; got != want {
			t.Fatalf("expected detail %q, got %q", want, got)
		}
	})

	t.Run("too long with multibyte UTF-8 characters", func(t *testing.T) {
		t.Parallel()

		diags := StringUTF8LenBetween(1, 2)("あいう", cty.Path{})

		if len(diags) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diags)
		}
		if diags[0].Severity != diag.Error {
			t.Fatalf("expected severity %q, got %q", diag.Error, diags[0].Severity)
		}
		if got, want := diags[0].Summary, "Invalid character length"; got != want {
			t.Fatalf("expected summary %q, got %q", want, got)
		}
		if got, want := diags[0].Detail, "expected length of pattern to be between 1 and 2 UTF-8 characters, got 3"; got != want {
			t.Fatalf("expected detail %q, got %q", want, got)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()

		diags := StringUTF8LenBetween(1, 2)(123, cty.Path{})

		if len(diags) != 1 {
			t.Fatalf("expected 1 diagnostic, got %#v", diags)
		}
		if diags[0].Severity != diag.Error {
			t.Fatalf("expected severity %q, got %q", diag.Error, diags[0].Severity)
		}
		if got, want := diags[0].Summary, "expected a string, got int"; got != want {
			t.Fatalf("expected summary %q, got %q", want, got)
		}
		if got, want := diags[0].Detail, "expected a string, got int"; got != want {
			t.Fatalf("expected detail %q, got %q", want, got)
		}
	})
}
