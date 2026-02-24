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
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_s3_directory_bucket")
func newDirectoryBucketResourceAsListResource() list.ListResourceWithConfigure {
	return &directoryBucketListResource{}
}

var _ list.ListResource = &directoryBucketListResource{}

type directoryBucketListResource struct {
	directoryBucketResource
	framework.WithList
}

func (r *directoryBucketListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := r.Meta().S3ExpressClient(ctx)

	var query listDirectoryBucketModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)
		var input s3.ListDirectoryBucketsInput
		for item, err := range listDirectoryBuckets(ctx, conn, &input) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			bucketName := aws.ToString(item.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrBucket), bucketName)

			bucket, err := findDirectoryBucket(ctx, conn, bucketName)
			if err != nil {
				tflog.Error(ctx, "Reading S3 Directory Bucket", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			var data directoryBucketResourceModel
			r.SetResult(ctx, r.Meta(), request.IncludeResource, &data, &result, func() {
				flattenDirectoryBucketResource(ctx, bucket, &data, &result.Diagnostics)
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

type listDirectoryBucketModel struct {
	framework.WithRegionModel
}

func listDirectoryBuckets(ctx context.Context, conn *s3.Client, input *s3.ListDirectoryBucketsInput) iter.Seq2[awstypes.Bucket, error] {
	return func(yield func(awstypes.Bucket, error) bool) {
		pages := s3.NewListDirectoryBucketsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Bucket{}, fmt.Errorf("listing S3 Directory Bucket resources: %w", err))
				return
			}

			for _, item := range page.Buckets {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
