// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

type bucketPropertyListHandlerSDK interface {
	parseQuery(ctx context.Context, config tfsdk.Config) diag.Diagnostics
	list(ctx context.Context, request list.ListRequest, conn *s3.Client, buckets iter.Seq2[awstypes.Bucket, error]) iter.Seq[list.ListResult]
}

type listResourceSDK interface {
	framework.WithMeta
	ResourceData() *schema.ResourceData
	SetResult(ctx context.Context, awsClient *conns.AWSClient, includeResource bool, result *list.ListResult, rd *schema.ResourceData)
}

func parseQuery[T any](ctx context.Context, config tfsdk.Config) (diags diag.Diagnostics) {
	var query T
	if config.Raw.IsKnown() && !config.Raw.IsNull() {
		diags = config.Get(ctx, &query)
	}
	return diags
}
