// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/internal/vcr"
)

// ConfigureWrapper composes test-only behavior into the AWS provider's
// Plugin SDKv2 ConfigureContextFunc lifecycle.
//
// A wrapper receives the next ConfigureContextFunc in the chain and
// returns a replacement that may transform the context, the result, or
// the diagnostics either before or after invoking next. Wrappers must
// invoke next exactly once.
//
// Implementations live alongside the test mechanism they wire in: the
// API-call recorder exposes [APICallRecorderWrapper], and a future
// OpenTelemetry hook would expose its own. VCR is composed automatically
// by [ProtoV5ProviderFactoriesWithWrappers] when the VCR_MODE / VCR_PATH
// environment variables are set; tests do not opt in to it explicitly.
type ConfigureWrapper func(next schema.ConfigureContextFunc) schema.ConfigureContextFunc

// chainConfigureWrappers composes wrappers right-to-left so the first
// wrapper in the slice runs outermost: its before-call code runs first
// and its after-call code runs last. Nil wrappers are skipped.
func chainConfigureWrappers(inner schema.ConfigureContextFunc, wrappers ...ConfigureWrapper) schema.ConfigureContextFunc {
	for i := len(wrappers) - 1; i >= 0; i-- {
		if w := wrappers[i]; w != nil {
			inner = w(inner)
		}
	}
	return inner
}

// ProtoV5ProviderFactoriesWithWrappers returns Plugin Protocol v5
// provider factories for the AWS provider, with the given
// ConfigureWrappers applied to each factory's ConfigureContextFunc.
//
// The first wrapper runs outermost; ordering matters when wrappers
// communicate via context or shared state.
//
// VCR record/replay is composed transparently: when [vcr.IsEnabled] is
// true at construction time, a VCR wrapper is prepended to the chain
// and the test is marked so [Test] / [ParallelTest] do not double-wrap.
// Tests therefore behave correctly under VCR without any code change.
//
// Example:
//
//	rec := apicall.NewRecorder()
//	factories := acctest.ProtoV5ProviderFactoriesWithWrappers(ctx, t,
//	    acctest.APICallRecorderWrapper(rec),
//	)
func ProtoV5ProviderFactoriesWithWrappers(
	ctx context.Context,
	t *testing.T,
	wrappers ...ConfigureWrapper,
) map[string]func() (tfprotov5.ProviderServer, error) {
	t.Helper()

	// Snapshot at construction so the factory closure and the auto-wrap
	// marker agree on whether VCR was enabled for this test.
	vcrEnabled := vcr.IsEnabled()
	if vcrEnabled {
		disableVCRAutoWrap(t)
	}

	return map[string]func() (tfprotov5.ProviderServer, error){
		ProviderName: func() (tfprotov5.ProviderServer, error) {
			providerServerFactory, primary, err := provider.ProtoV5ProviderServerFactory(ctx)
			if err != nil {
				return nil, err
			}

			chain := wrappers
			if vcrEnabled {
				// VCR runs outermost so its HTTP-client swap and provider-identity
				// cache happen around any inner wrapper.
				chain = append([]ConfigureWrapper{vcrConfigureWrapper(primary, t)}, wrappers...)
			}
			primary.ConfigureContextFunc = chainConfigureWrappers(primary.ConfigureContextFunc, chain...)

			return providerServerFactory(), nil
		},
	}
}
