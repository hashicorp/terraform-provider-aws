// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
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
// OpenTelemetry hook would expose its own. VCR has its own auto-wrap
// path inside [Test] / [ParallelTest] and is not a ConfigureWrapper
// today; it could be migrated when its provider-identity caching is
// reworked.
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
// communicate via context or shared state. With no wrappers the result
// is equivalent to building factories directly from the provider.
//
// Example:
//
//	rec := apicall.NewRecorder()
//	factories := acctest.ProtoV5ProviderFactoriesWithWrappers(ctx,
//	    acctest.APICallRecorderWrapper(rec),
//	)
func ProtoV5ProviderFactoriesWithWrappers(
	ctx context.Context,
	wrappers ...ConfigureWrapper,
) map[string]func() (tfprotov5.ProviderServer, error) {
	return map[string]func() (tfprotov5.ProviderServer, error){
		ProviderName: func() (tfprotov5.ProviderServer, error) {
			providerServerFactory, primary, err := provider.ProtoV5ProviderServerFactory(ctx)
			if err != nil {
				return nil, err
			}
			primary.ConfigureContextFunc = chainConfigureWrappers(primary.ConfigureContextFunc, wrappers...)
			return providerServerFactory(), nil
		},
	}
}
