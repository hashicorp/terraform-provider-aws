// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sweep

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

func Context(region string) context.Context {
	ctx := context.Background()

	ctx = tfsdklog.RegisterStdlogSink(ctx)

	ctx = logger(ctx, "sweeper", region)

	return ctx
}

func logger(ctx context.Context, loggerName, region string) context.Context {
	ctx = tfsdklog.NewRootProviderLogger(ctx,
		tfsdklog.WithLevel(hclog.Debug),
		tfsdklog.WithLogName(loggerName),
		tfsdklog.WithoutLocation(),
	)
	ctx = tflog.SetField(ctx, "sweeper_region", region)

	return ctx
}
