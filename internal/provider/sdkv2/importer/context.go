// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"

	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type contextKey int

const (
	identitySpecKey contextKey = 1
	importSpecKey   contextKey = 2
)

func Context(ctx context.Context, identity *inttypes.Identity, importSpec *inttypes.SDKv2Import) context.Context {
	ctx = context.WithValue(ctx, identitySpecKey, identity)
	ctx = context.WithValue(ctx, importSpecKey, importSpec)
	return ctx
}

func identitySpec(ctx context.Context) inttypes.Identity {
	val := ctx.Value(identitySpecKey)
	if identity, ok := val.(*inttypes.Identity); ok {
		return *identity
	}
	return inttypes.Identity{}
}

func importSpec(ctx context.Context) inttypes.SDKv2Import {
	val := ctx.Value(importSpecKey)
	if importSpec, ok := val.(*inttypes.SDKv2Import); ok {
		return *importSpec
	}
	return inttypes.SDKv2Import{}
}
