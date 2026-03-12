// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

type bucketPropertyListHandlerSDK interface {
	parseQuery(ctx context.Context, config tfsdk.Config) diag.Diagnostics
	list(ctx context.Context, request list.ListRequest, conn *s3.Client, buckets iter.Seq2[awstypes.Bucket, error]) iter.Seq[list.ListResult]
}

var _ list.ListResource = &listResourceBaseBucketPropertySDK{}

func newListResourceBaseBucketPropertySDK(resource *schema.Resource, f func(listResourceSDK) bucketPropertyListHandlerSDK) inttypes.ListResourceForSDK {
	l := listResourceBaseBucketPropertySDK{}
	l.SetResourceSchema(resource)
	l.handler = f(&l)
	return &l
}

type listResourceBaseBucketPropertySDK struct {
	framework.ListResourceWithSDKv2Resource
	handler bucketPropertyListHandlerSDK
}

func (l *listResourceBaseBucketPropertySDK) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	if diags := l.handler.parseQuery(ctx, request.Config); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	tflog.Info(ctx, "Listing Resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		tflog.Info(ctx, "Listing General Purpose Buckets")
		gpConn := l.Meta().S3Client(ctx)
		gpInput := s3.ListBucketsInput{
			BucketRegion: aws.String(l.Meta().Region(ctx)),
			MaxBuckets:   aws.Int32(int32(request.Limit)),
		}
		var count int64
		for result := range l.handler.list(ctx, request, gpConn, listBuckets(ctx, gpConn, &gpInput)) {
			count++
			if !yield(result) {
				return
			}
		}

		limit := request.Limit - count
		if limit <= 0 {
			tflog.Info(ctx, "Limit reached, skipping Directory Buckets")
		}

		tflog.Info(ctx, "Listing Directory Buckets")
		dirConn := l.Meta().S3ExpressClient(ctx)
		dirInput := s3.ListDirectoryBucketsInput{
			MaxDirectoryBuckets: aws.Int32(int32(limit)),
		}
		for result := range l.handler.list(ctx, request, dirConn, listDirectoryBuckets(ctx, dirConn, &dirInput)) {
			if !yield(result) {
				return
			}
		}
	}
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

type baseBucketPropertyListHandlerSDK struct {
	lister listResourceSDK
}

func newBaseBucketPropertyListHandlerSDK(lister listResourceSDK) baseBucketPropertyListHandlerSDK {
	return baseBucketPropertyListHandlerSDK{
		lister: lister,
	}
}

func (l baseBucketPropertyListHandlerSDK) Meta() *conns.AWSClient {
	return l.lister.Meta()
}

func (l baseBucketPropertyListHandlerSDK) ResourceData() *schema.ResourceData {
	return l.lister.ResourceData()
}

func (l baseBucketPropertyListHandlerSDK) SetResult(ctx context.Context, awsClient *conns.AWSClient, includeResource bool, result *list.ListResult, rd *schema.ResourceData) {
	l.lister.SetResult(ctx, awsClient, includeResource, result, rd)
}
