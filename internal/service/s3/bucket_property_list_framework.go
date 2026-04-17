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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

type bucketPropertyListHandlerFramework interface {
	parseQuery(ctx context.Context, config tfsdk.Config) diag.Diagnostics
	list(ctx context.Context, request list.ListRequest, conn *s3.Client, buckets iter.Seq2[awstypes.Bucket, error]) iter.Seq[list.ListResult]
}

type listResourceFramework interface {
	framework.WithMeta
	SetResult(ctx context.Context, awsClient *conns.AWSClient, includeResource bool, data any, result *list.ListResult, f framework.FlattenFunc)
}

type baseBucketPropertyListHandlerFramework struct {
	lister listResourceFramework
}

func newBaseBucketPropertyListHandlerFramework(lister listResourceFramework) baseBucketPropertyListHandlerFramework {
	return baseBucketPropertyListHandlerFramework{
		lister: lister,
	}
}

func (l baseBucketPropertyListHandlerFramework) Meta() *conns.AWSClient {
	return l.lister.Meta()
}

func (l baseBucketPropertyListHandlerFramework) SetResult(ctx context.Context, awsClient *conns.AWSClient, includeResource bool, data any, result *list.ListResult, f func()) {
	l.lister.SetResult(ctx, awsClient, includeResource, data, result, f)
}
