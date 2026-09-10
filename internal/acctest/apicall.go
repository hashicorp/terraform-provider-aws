// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Test helpers for asserting which AWS SDK for Go v2 API operations a
// resource invokes. Backed by internal/conns/apicall.

package acctest

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/conns/apicall"
)

// APICallRecorderWrapper returns a [ConfigureWrapper] that attaches rec
// to the *conns.AWSClient produced by Configure, so subsequent AWS SDK
// operations record into rec. The recorder is the integration point with
// the apicall middleware: see internal/conns/apicall for the model.
//
// Compose with [ProtoV5ProviderFactoriesWithWrappers] alongside other
// wrappers, or use [ProtoV5ProviderFactoriesWithCallRecorder] for the
// common case of a single fresh recorder per test.
func APICallRecorderWrapper(rec *apicall.Recorder) ConfigureWrapper {
	return func(next schema.ConfigureContextFunc) schema.ConfigureContextFunc {
		return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
			v, ds := next(ctx, d)
			if c, ok := v.(*conns.AWSClient); ok && c != nil {
				c.SetCallRecorder(rec)
			}
			return v, ds
		}
	}
}

// ProtoV5ProviderFactoriesWithCallRecorder returns ProtoV5 provider
// factories that attach a fresh API call recorder to the *conns.AWSClient
// produced by each Configure. The same recorder is used across all factory
// invocations in the test, so plan, apply, and refresh all record to it.
//
// Use Mark/CallsSince/ContainsSince to scope assertions to a window.
//
// Example:
//
//	factories, rec := acctest.ProtoV5ProviderFactoriesWithCallRecorder(ctx, t)
//	var step2Mark apicall.Cursor
//	resource.Test(t, resource.TestCase{
//	    ProtoV5ProviderFactories: factories,
//	    Steps: []resource.TestStep{
//	        { Config: configWithSettings(rName) },
//	        {
//	            PreConfig: func() { step2Mark = rec.Mark() },
//	            Config:    configWithoutSettings(rName),
//	            Check: resource.ComposeTestCheckFunc(
//	                acctest.CheckAPICallNotMade(rec, &step2Mark, "Pinpoint", "GetApplicationSettings"),
//	                acctest.CheckAPICallNotMade(rec, &step2Mark, "Pinpoint", "UpdateApplicationSettings"),
//	            ),
//	        },
//	    },
//	})
//
// To compose the recorder with other ConfigureWrappers, use
// [ProtoV5ProviderFactoriesWithWrappers] and [APICallRecorderWrapper]
// directly.
func ProtoV5ProviderFactoriesWithCallRecorder(ctx context.Context, t *testing.T) (
	map[string]func() (tfprotov5.ProviderServer, error),
	*apicall.Recorder,
) {
	t.Helper()

	rec := apicall.NewRecorder()
	return ProtoV5ProviderFactoriesWithWrappers(ctx, t, APICallRecorderWrapper(rec)), rec
}

// CheckAPICallMade fails if service.operation was not recorded since the
// cursor pointed to by since, or since the start of recording when since is
// nil. The pointer indirection lets PreConfig populate the cursor after the
// Check slice is built.
//
// Service is the Smithy ServiceID (e.g. "Pinpoint", exposed as
// <package>.ServiceID).
func CheckAPICallMade(rec *apicall.Recorder, since *apicall.Cursor, service, operation string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if rec == nil {
			return fmt.Errorf("CheckAPICallMade: recorder is nil")
		}
		var cursor apicall.Cursor
		if since != nil {
			cursor = *since
		}
		if rec.ContainsSince(cursor, service, operation) {
			return nil
		}
		return fmt.Errorf("expected AWS API call %s.%s, not made; calls since cursor: %s",
			service, operation, formatCalls(rec.CallsSince(cursor)))
	}
}

// CheckAPICallNotMade fails if service.operation was recorded since the
// cursor pointed to by since, or since the start of recording when since is
// nil.
func CheckAPICallNotMade(rec *apicall.Recorder, since *apicall.Cursor, service, operation string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if rec == nil {
			return fmt.Errorf("CheckAPICallNotMade: recorder is nil")
		}
		var cursor apicall.Cursor
		if since != nil {
			cursor = *since
		}
		if !rec.ContainsSince(cursor, service, operation) {
			return nil
		}
		return fmt.Errorf("unexpected AWS API call %s.%s; calls since cursor: %s",
			service, operation, formatCalls(rec.CallsSince(cursor)))
	}
}

// formatCalls renders calls compactly for failure messages.
func formatCalls(calls []apicall.Call) string {
	if len(calls) == 0 {
		return "(none)"
	}
	parts := make([]string, len(calls))
	for i, c := range calls {
		if c.Err != nil {
			parts[i] = fmt.Sprintf("%s.%s(err=%v)", c.Service, c.Operation, c.Err)
		} else {
			parts[i] = fmt.Sprintf("%s.%s", c.Service, c.Operation)
		}
	}
	return strings.Join(parts, ", ")
}
