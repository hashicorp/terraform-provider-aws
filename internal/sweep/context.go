// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sweep

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

func Context(region string) context.Context {
	ctx := context.Background()

	ctx = tfsdklog.RegisterStdlogSink(ctx)

	ctx = logger(ctx, "sweeper", region)

	return ctx
}
