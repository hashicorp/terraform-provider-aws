// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package smerr

import (
	"context"

	"github.com/YakDriver/smarterr"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ID = smarterr.ID
)

// This is smarterr wrapping to inject private context into keyvals for the SDK and Framework diagnostics.

// Append enriches smarterr.Append with resource and service context if available.
func Append(ctx context.Context, diags sdkdiag.Diagnostics, err error, keyvals ...any) sdkdiag.Diagnostics {
	return smarterr.Append(ctx, diags, err, injectContext(ctx, keyvals...)...)
}

// AppendOne enriches smarterr.AppendOne with resource and service context if available.
func AppendOne(ctx context.Context, existing sdkdiag.Diagnostics, incoming sdkdiag.Diagnostic, keyvals ...any) sdkdiag.Diagnostics {
	return smarterr.AppendOne(ctx, existing, incoming, injectContext(ctx, keyvals...)...)
}

// AppendEnrich enriches smarterr.AppendEnrich with resource and service context if available.
func AppendEnrich(ctx context.Context, existing sdkdiag.Diagnostics, incoming sdkdiag.Diagnostics, keyvals ...any) sdkdiag.Diagnostics {
	return smarterr.AppendEnrich(ctx, existing, incoming, injectContext(ctx, keyvals...)...)
}

// AddError enriches smarterr.AddError with resource and service context if available.
func AddError(ctx context.Context, diags *fwdiag.Diagnostics, err error, keyvals ...any) {
	smarterr.AddError(ctx, diags, err, injectContext(ctx, keyvals...)...)
}

// AddOne enriches smarterr.AddOne with resource and service context if available.
func AddOne(ctx context.Context, existing *fwdiag.Diagnostics, incoming fwdiag.Diagnostic, keyvals ...any) {
	smarterr.AddOne(ctx, existing, incoming, injectContext(ctx, keyvals...)...)
}

// AddEnrich enriches smarterr.AddEnrich with resource and service context if available.
func AddEnrich(ctx context.Context, existing *fwdiag.Diagnostics, incoming fwdiag.Diagnostics, keyvals ...any) {
	smarterr.AddEnrich(ctx, existing, incoming, injectContext(ctx, keyvals...)...)
}

// EnrichAppend enriches smarterr.EnrichAppend with resource and service context if available.
//
// Deprecated: Use AddEnrich instead.
func EnrichAppend(ctx context.Context, existing *fwdiag.Diagnostics, incoming fwdiag.Diagnostics, keyvals ...any) {
	smarterr.AddEnrich(ctx, existing, incoming, injectContext(ctx, keyvals...)...)
}

func injectContext(ctx context.Context, keyvals ...any) []any {
	if inctx, ok := conns.FromContext(ctx); ok {
		srv := inctx.ServicePackageName()
		if v, err := names.HumanFriendly(srv); err == nil {
			srv = v
		}
		keyvals = append(keyvals, smarterr.ResourceName, inctx.ResourceName(), smarterr.ServiceName, srv)
	}
	return keyvals
}
