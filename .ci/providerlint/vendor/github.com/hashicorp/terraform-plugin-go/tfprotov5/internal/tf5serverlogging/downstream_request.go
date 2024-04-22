// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tf5serverlogging

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-go/internal/logging"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/diag"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/internal/funcerr"
)

// DownstreamRequest sets a request duration start time context key and
// generates a TRACE "Sending request downstream" log.
func DownstreamRequest(ctx context.Context) context.Context {
	requestStart := time.Now()
	ctx = context.WithValue(ctx, ContextKeyDownstreamRequestStartTime{}, requestStart)

	logging.ProtocolTrace(ctx, "Sending request downstream")

	return ctx
}

// DownstreamResponse generates the following logging:
//
//   - TRACE "Received downstream response" log with request duration and
//     diagnostic severity counts
//   - Per-diagnostic logs
func DownstreamResponse(ctx context.Context, diagnostics diag.Diagnostics) {
	responseFields := map[string]interface{}{
		logging.KeyDiagnosticErrorCount:   diagnostics.ErrorCount(),
		logging.KeyDiagnosticWarningCount: diagnostics.WarningCount(),
	}

	if requestStart, ok := ctx.Value(ContextKeyDownstreamRequestStartTime{}).(time.Time); ok {
		responseFields[logging.KeyRequestDurationMs] = time.Since(requestStart).Milliseconds()
	}

	logging.ProtocolTrace(ctx, "Received downstream response", responseFields)
	diagnostics.Log(ctx)
}

// DownstreamResponseWithError generates the following logging:
//
//   - TRACE "Received downstream response" log with request duration and
//     whether a function error is present
//   - Log with function error details
func DownstreamResponseWithError(ctx context.Context, funcErr *tfprotov5.FunctionError) {
	fe := (*funcerr.FunctionError)(funcErr)

	responseFields := map[string]interface{}{
		logging.KeyFunctionErrorExists: fe.HasError(),
	}

	if requestStart, ok := ctx.Value(ContextKeyDownstreamRequestStartTime{}).(time.Time); ok {
		responseFields[logging.KeyRequestDurationMs] = time.Since(requestStart).Milliseconds()
	}

	logging.ProtocolTrace(ctx, "Received downstream response", responseFields)
	fe.Log(ctx)
}
