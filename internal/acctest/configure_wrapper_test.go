// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"reflect"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

// TestProtoV5ProviderFactoriesWithWrappers_NoVCR verifies that with VCR
// environment variables unset, the auto-wrap marker is not set so the
// existing acctest.Test/ParallelTest auto-wrap path remains in charge of
// VCR for legacy callers using bare factories.
func TestProtoV5ProviderFactoriesWithWrappers_NoVCR(t *testing.T) {
	t.Setenv("VCR_MODE", "")
	t.Setenv("VCR_PATH", "")

	factories := ProtoV5ProviderFactoriesWithWrappers(context.Background(), t)
	if factories == nil {
		t.Fatal("expected non-nil factories")
	}
	if isVCRAutoWrapDisabled(t) {
		t.Error("VCR auto-wrap should not be disabled when VCR is not enabled")
	}
}

// TestProtoV5ProviderFactoriesWithWrappers_VCREnabled_TransparentlyMarksTest
// verifies that setting the VCR env vars causes the factory builder to
// transparently disable acctest.Test's auto-wrap, so VCR composes with
// any inner wrappers (apicall, future OTEL) instead of replacing them.
func TestProtoV5ProviderFactoriesWithWrappers_VCREnabled_TransparentlyMarksTest(t *testing.T) {
	t.Setenv("VCR_MODE", "RECORD_ONLY")
	t.Setenv("VCR_PATH", t.TempDir())

	factories := ProtoV5ProviderFactoriesWithWrappers(context.Background(), t)
	if factories == nil {
		t.Fatal("expected non-nil factories")
	}
	if !isVCRAutoWrapDisabled(t) {
		t.Error("VCR auto-wrap should be disabled when factories include VCR transparently")
	}
}

// TestDisableVCRAutoWrap_Idempotent verifies that repeated calls leave
// the marker set without panicking.
func TestDisableVCRAutoWrap_Idempotent(t *testing.T) {
	t.Parallel()

	disableVCRAutoWrap(t)
	disableVCRAutoWrap(t)
	disableVCRAutoWrap(t)
	if !isVCRAutoWrapDisabled(t) {
		t.Error("expected marker to be set after repeated calls")
	}
}

// TestVCRTestCase_ShortCircuitsWhenAutoWrapDisabled verifies that
// vcrTestCase is a no-op when [disableVCRAutoWrap] has been called for
// the test, so factories that already include VCR via the wrapper API
// are not double-wrapped.
func TestVCRTestCase_ShortCircuitsWhenAutoWrapDisabled(t *testing.T) {
	t.Setenv("VCR_MODE", "RECORD_ONLY")
	t.Setenv("VCR_PATH", t.TempDir())

	disableVCRAutoWrap(t)

	// Sentinel factory we can recognize after vcrTestCase returns.
	original := map[string]func() (tfprotov5.ProviderServer, error){
		ProviderName: func() (tfprotov5.ProviderServer, error) { return nil, nil },
	}
	tc := resource.TestCase{ProtoV5ProviderFactories: original}

	if !vcrTestCase(context.Background(), t, &tc) {
		t.Fatal("vcrTestCase returned false; want true so the test still runs")
	}

	if reflect.ValueOf(tc.ProtoV5ProviderFactories).Pointer() !=
		reflect.ValueOf(original).Pointer() {
		t.Error("vcrTestCase replaced ProtoV5ProviderFactories despite the auto-wrap-disabled marker")
	}
}
