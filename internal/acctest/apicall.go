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
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

// ProtoV5ProviderFactoriesWithCallRecorder returns ProtoV5 provider
// factories that, on each Configure, attach a fresh API call recorder to the
// resulting *conns.AWSClient. Tests can then use the returned recorder with
// the CheckAPICall* TestCheckFuncs (or directly via the apicall package) to
// assert which AWS SDK for Go v2 operations a resource did or did not
// invoke.
//
// The recorder is shared across all factory invocations within the test, so
// retries, refreshes, and apply-after-plan all record into the same log. Use
// recorder.Mark / CallsSince / ContainsSince to scope assertions to a window
// (typically a single resource.TestStep).
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
func ProtoV5ProviderFactoriesWithCallRecorder(ctx context.Context, t *testing.T) (
	map[string]func() (tfprotov5.ProviderServer, error),
	*apicall.Recorder,
) {
	t.Helper()

	rec := apicall.NewRecorder()

	factories := map[string]func() (tfprotov5.ProviderServer, error){
		ProviderName: func() (tfprotov5.ProviderServer, error) {
			providerServerFactory, primary, err := provider.ProtoV5ProviderServerFactory(ctx)
			if err != nil {
				return nil, err
			}
			primary.ConfigureContextFunc = wrapConfigureWithCallRecorder(primary.ConfigureContextFunc, rec)
			return providerServerFactory(), nil
		},
	}

	return factories, rec
}

// wrapConfigureWithCallRecorder returns a ConfigureContextFunc that runs the
// original configure and, if it produced an *conns.AWSClient, attaches the
// recorder to it. Errors and warnings from the original configure are
// returned unchanged.
func wrapConfigureWithCallRecorder(original schema.ConfigureContextFunc, rec *apicall.Recorder) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		v, ds := original(ctx, d)
		if c, ok := v.(*conns.AWSClient); ok && c != nil {
			c.SetCallRecorder(rec)
		}
		return v, ds
	}
}

// CheckAPICallMade returns a TestCheckFunc that fails if the named AWS API
// operation was not recorded since the cursor pointed to by since (or, if
// since is nil, since the start of recording).
//
// Service is the Smithy ServiceID (e.g. "Pinpoint", "S3"). Operation is the
// operation name (e.g. "GetApplicationSettings"). The Smithy ServiceID for a
// given AWS SDK for Go v2 client is exposed as <package>.ServiceID at the
// package root.
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
		return fmt.Errorf("expected AWS API call %s.%s to have been made, but it was not; recorded since cursor: %s",
			service, operation, formatCalls(rec.CallsSince(cursor)))
	}
}

// CheckAPICallNotMade returns a TestCheckFunc that fails if the named AWS
// API operation was recorded since the cursor pointed to by since (or, if
// since is nil, since the start of recording).
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
		return fmt.Errorf("expected AWS API call %s.%s to NOT have been made, but it was; recorded since cursor: %s",
			service, operation, formatCalls(rec.CallsSince(cursor)))
	}
}

// formatCalls returns a compact, human-readable rendering of the given calls
// suitable for inclusion in test failure messages.
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
