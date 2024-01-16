// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sweep

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

const (
	loggingKeySweeperRegion = "sweeper_region"

	// Copied from:
	// * https://github.com/hashicorp/terraform-plugin-sdk/blob/ffbf0104398c0aa91aa3a82aff4b67e260677454/internal/logging/keys.go#L29
	// * https://github.com/hashicorp/terraform-plugin-framework/blob/743126edc3b04e735c05f2ddfa42e990b7231600/internal/logging/keys.go#L29
	loggingKeyResourceType = "tf_resource_type"
)

func logger(ctx context.Context, loggerName, region string) context.Context {
	ctx = tfsdklog.NewRootProviderLogger(ctx,
		tfsdklog.WithLevel(hclog.Debug),
		tfsdklog.WithLogName(loggerName),
		tfsdklog.WithoutLocation(),
	)
	ctx = tflog.SetField(ctx, loggingKeySweeperRegion, region)

	return ctx
}

func logWithResourceType(ctx context.Context, resourceType string) context.Context {
	return tflog.SetField(ctx, loggingKeyResourceType, resourceType)
}
