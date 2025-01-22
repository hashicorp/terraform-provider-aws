// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sweep

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/internal/log"
)

func Context(region string) context.Context {
	ctx := context.Background()

	ctx = tfsdklog.RegisterStdlogSink(ctx)

	ctx = log.Logger(ctx, "sweeper", region)

	return ctx
}
