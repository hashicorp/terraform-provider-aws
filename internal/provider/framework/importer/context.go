// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"context"
)

type contextKey int

var awsClientKey contextKey = 1 // nosemgrep:ci.aws-in-var-name

type AWSClient interface {
	AccountID(context.Context) string
	Region(ctx context.Context) string
}

func Context(ctx context.Context, client AWSClient) context.Context {
	return context.WithValue(ctx, awsClientKey, client)
}

func Client(ctx context.Context) AWSClient {
	val := ctx.Value(awsClientKey)
	if client, ok := val.(AWSClient); ok {
		return client
	}
	return nil
}
