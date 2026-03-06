// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_s3_bucket_versioning")
func newBucketVersioningResourceAsListResource() inttypes.ListResourceForSDK {
	l := bucketVersioningListResource{}
	l.SetResourceSchema(resourceBucketVersioning())
	l.handler = bucketVersioningListHandler{
		lister: &l,
	}
	return &l
}

var _ list.ListResource = &bucketVersioningListResource{}

type bucketVersioningListResource struct {
	framework.ListResourceWithSDKv2Resource
	handler bucketPropertyListHandlerSDK
}

func (l *bucketVersioningListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
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

type listBucketVersioningModel struct {
	framework.WithRegionModel
}

var _ bucketPropertyListHandlerSDK = bucketVersioningListHandler{}

type bucketVersioningListHandler struct {
	lister listResourceSDK
}

func (l bucketVersioningListHandler) Meta() *conns.AWSClient {
	return l.lister.Meta()
}

func (l bucketVersioningListHandler) ResourceData() *schema.ResourceData {
	return l.lister.ResourceData()
}

func (l bucketVersioningListHandler) SetResult(ctx context.Context, awsClient *conns.AWSClient, includeResource bool, result *list.ListResult, rd *schema.ResourceData) {
	l.lister.SetResult(ctx, awsClient, includeResource, result, rd)
}

func (l bucketVersioningListHandler) parseQuery(ctx context.Context, config tfsdk.Config) (diags diag.Diagnostics) {
	return parseQuery[listBucketVersioningModel](ctx, config)
}

func (l bucketVersioningListHandler) list(ctx context.Context, request list.ListRequest, conn *s3.Client, buckets iter.Seq2[awstypes.Bucket, error]) iter.Seq[list.ListResult] {
	return func(yield func(list.ListResult) bool) {
		for bucket, err := range buckets {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing S3 Bucket Versioning resources: %w", err))
				yield(result)
				return
			}

			bucketName := aws.ToString(bucket.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrBucket), bucketName)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(bucketName)
			rd.Set(names.AttrBucket, bucketName)

			// There is always a Bucket Versioning Configuration associated with a Bucket (1:1)
			// So only read it if resource data is requested.
			if request.IncludeResource {
				tflog.Info(ctx, "Reading S3 Bucket Versioning")
				versioning, err := findBucketVersioning(ctx, conn, bucketName, l.Meta().AccountID(ctx))
				if retry.NotFound(err) {
					tflog.Warn(ctx, "Resource disappeared during listing, skipping")
					continue
				}
				if err != nil {
					tflog.Error(ctx, "Reading S3 Bucket Versioning", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				if err := resourceBucketVersioningFlatten(ctx, versioning, rd); err != nil {
					tflog.Error(ctx, "Reading S3 Bucket Versioning", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = bucketName

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &result, rd)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}
