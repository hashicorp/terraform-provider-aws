// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestChainConfigureWrappers verifies that wrappers compose so that the
// first wrapper in the slice runs outermost: its before-call code runs
// first and its after-call code runs last.
func TestChainConfigureWrappers(t *testing.T) {
	t.Parallel()

	var trace []string

	wrap := func(label string) ConfigureWrapper {
		return func(next schema.ConfigureContextFunc) schema.ConfigureContextFunc {
			return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
				trace = append(trace, "before:"+label)
				v, ds := next(ctx, d)
				trace = append(trace, "after:"+label)
				return v, ds
			}
		}
	}

	inner := func(_ context.Context, _ *schema.ResourceData) (any, diag.Diagnostics) {
		trace = append(trace, "inner")
		return "result", nil
	}

	chained := chainConfigureWrappers(inner, wrap("A"), wrap("B"), wrap("C"))

	v, diags := chained(context.Background(), nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if v != "result" {
		t.Errorf("returned %v, want %q", v, "result")
	}

	want := []string{
		"before:A", "before:B", "before:C",
		"inner",
		"after:C", "after:B", "after:A",
	}
	if !slices.Equal(trace, want) {
		t.Errorf("trace = %v, want %v", trace, want)
	}
}

// TestChainConfigureWrappers_NilEntriesSkipped verifies that nil wrappers
// in the slice are silently skipped.
func TestChainConfigureWrappers_NilEntriesSkipped(t *testing.T) {
	t.Parallel()

	var calls int
	wrap := func(next schema.ConfigureContextFunc) schema.ConfigureContextFunc {
		return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
			calls++
			return next(ctx, d)
		}
	}

	inner := func(_ context.Context, _ *schema.ResourceData) (any, diag.Diagnostics) {
		return nil, nil
	}

	chained := chainConfigureWrappers(inner, nil, wrap, nil, wrap, nil)
	if _, diags := chained(context.Background(), nil); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if calls != 2 {
		t.Errorf("wrap called %d times, want 2", calls)
	}
}

// TestChainConfigureWrappers_NoWrappers verifies that an empty wrapper
// list returns inner unchanged.
func TestChainConfigureWrappers_NoWrappers(t *testing.T) {
	t.Parallel()

	called := false
	inner := func(_ context.Context, _ *schema.ResourceData) (any, diag.Diagnostics) {
		called = true
		return "x", nil
	}

	chained := chainConfigureWrappers(inner)
	v, diags := chained(context.Background(), nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !called {
		t.Error("inner was not invoked")
	}
	if v != "x" {
		t.Errorf("returned %v, want %q", v, "x")
	}
}
