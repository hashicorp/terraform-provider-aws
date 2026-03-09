// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_s3_bucket_lifecycle_configuration")
func newBucketLifecycleConfigurationResourceAsListResource() list.ListResourceWithConfigure {
	return &bucketLifecycleConfigurationListResource{}
}

var _ list.ListResource = &bucketLifecycleConfigurationListResource{}

type bucketLifecycleConfigurationListResource struct {
	bucketLifecycleConfigurationResource
	framework.WithList
}

func (l *bucketLifecycleConfigurationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query listBucketLifecycleConfigurationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		tflog.Info(ctx, "Listing General Purpose Buckets")
		gpConn := l.Meta().S3Client(ctx)
		gpInput := s3.ListBucketsInput{
			BucketRegion: aws.String(l.Meta().Region(ctx)),
			MaxBuckets:   aws.Int32(int32(request.Limit)),
		}
		var count int64
		for result := range l.list(ctx, request, gpConn, listBuckets(ctx, gpConn, &gpInput)) {
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
		for result := range l.list(ctx, request, gpConn, listDirectoryBuckets(ctx, dirConn, &dirInput)) {
			if !yield(result) {
				return
			}
		}
	}
}

func (l *bucketLifecycleConfigurationListResource) list(ctx context.Context, request list.ListRequest, conn *s3.Client, buckets iter.Seq2[types.Bucket, error]) iter.Seq[list.ListResult] {
	return func(yield func(list.ListResult) bool) {
		for bucket, err := range buckets {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing S3 Bucket Lifecycle Configuration resources: %w", err))
				yield(result)
				return
			}

			bucketName := aws.ToString(bucket.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrBucket), bucketName)

			result := request.NewListResult(ctx)

			item, err := findBucketLifecycleConfiguration(ctx, conn, bucketName, l.Meta().AccountID(ctx))
			if retry.NotFound(err) {
				tflog.Debug(ctx, "Bucket has no Lifecycle Configuration, skipping")
				continue
			}
			if err != nil {
				tflog.Error(ctx, "Reading S3 Bucket Lifecycle Configuration", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			var data bucketLifecycleConfigurationResourceModel

			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				flattenBucketLifecycleConfigurationResource(ctx, item, &data, &result.Diagnostics)
				data.Bucket = fwflex.StringValueToFramework(ctx, bucketName)
				data.ID = data.Bucket

				result.DisplayName = bucketName
			})

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

type listBucketLifecycleConfigurationModel struct {
	framework.WithRegionModel
}
