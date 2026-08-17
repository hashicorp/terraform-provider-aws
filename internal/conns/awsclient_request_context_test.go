// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"bytes"
	"context"
	"strings"
	"testing"

	baselogging "github.com/hashicorp/aws-sdk-go-base/v2/logging"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
)

const (
	requestBodyPayload  = "request-body-payload"
	responseBodyPayload = "response-body-payload"
)

// requestContextTestSetup builds a context with a tflog root logger backed
// by buf and a non-null baselogging logger registered, so that
// MaskSensitiveValuesByKey does not short-circuit on a NullLogger and
// tflog records flow into buf where the test can inspect them. It also
// returns an AWSClient configured with that same logger so that
// c.RegisterLogger is a real operation.
func requestContextTestSetup(t *testing.T) (context.Context, *AWSClient, *bytes.Buffer) {
	t.Helper()

	var buf bytes.Buffer
	ctx := tflogtest.RootLogger(context.Background(), &buf)

	// baselogging.NewTfLogger reads the tflog logger from ctx and wraps it
	// into a non-null aws-sdk-go-base logger; without this,
	// MaskSensitiveValuesByKey treats the logger as a no-op and returns ctx
	// unchanged, which would hide a regression in its own behavior.
	ctx, logger := baselogging.NewTfLogger(ctx)

	c := &AWSClient{logger: logger}
	return ctx, c, &buf
}

// emitHTTPBodyRecord writes a tflog record carrying both HTTP body fields,
// so tests can assert how the active context masks (or doesn't mask) them.
func emitHTTPBodyRecord(ctx context.Context) {
	tflog.Info(ctx, "AWS request", map[string]any{
		logging.HTTPKeyRequestBody:  requestBodyPayload,
		logging.HTTPKeyResponseBody: responseBodyPayload,
	})
}

// TestAWSClientRequestContext_DoesNotMaskHTTPBodies guards #48461: the
// per-request context wired into ordinary resources, data sources, and
// list resources must not redact http.request.body / http.response.body
// in tflog output, since doing so removes the AWS SDK debug trace
// developers rely on.
func TestAWSClientRequestContext_DoesNotMaskHTTPBodies(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	ctx, c, buf := requestContextTestSetup(t)
	ctx = c.RequestContext(ctx)

	emitHTTPBodyRecord(ctx)

	out := buf.String()
	if !strings.Contains(out, requestBodyPayload) {
		t.Errorf("RequestContext masked http.request.body unexpectedly; output: %s", out)
	}
	if !strings.Contains(out, responseBodyPayload) {
		t.Errorf("RequestContext masked http.response.body unexpectedly; output: %s", out)
	}
}

// TestAWSClientEphemeralRequestContext_MasksHTTPBodies covers the
// invariant that ephemeral resources and actions, whose API responses
// commonly carry secrets, get http.request.body / http.response.body
// redacted in tflog output.
func TestAWSClientEphemeralRequestContext_MasksHTTPBodies(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	ctx, c, buf := requestContextTestSetup(t)
	ctx = c.EphemeralRequestContext(ctx)

	emitHTTPBodyRecord(ctx)

	out := buf.String()
	if strings.Contains(out, requestBodyPayload) {
		t.Errorf("EphemeralRequestContext failed to mask http.request.body; output: %s", out)
	}
	if strings.Contains(out, responseBodyPayload) {
		t.Errorf("EphemeralRequestContext failed to mask http.response.body; output: %s", out)
	}
	if !strings.Contains(out, "***") {
		t.Errorf("EphemeralRequestContext output is missing the redaction marker; output: %s", out)
	}
}

// TestAWSClientRequestContext_NilReceiverIsSafe documents the contract
// that the RequestContext family is callable on a nil *AWSClient (e.g.
// during early test setup), returning ctx unchanged.
func TestAWSClientRequestContext_NilReceiverIsSafe(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	ctx := context.Background()
	var c *AWSClient

	if got := c.RequestContext(ctx); got != ctx {
		t.Errorf("RequestContext on nil receiver mutated ctx")
	}
	if got := c.EphemeralRequestContext(ctx); got != ctx {
		t.Errorf("EphemeralRequestContext on nil receiver mutated ctx")
	}
}
