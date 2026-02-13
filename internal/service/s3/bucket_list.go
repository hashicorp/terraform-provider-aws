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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_s3_bucket")
func newBucketResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceBucket{}
	l.SetResourceSchema(resourceBucket())
	return &l
}

var _ list.ListResource = &listResourceBucket{}

type listResourceBucket struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceBucket) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().S3Client(ctx)

	var query listBucketModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing S3 Bucket")
	stream.Results = func(yield func(list.ListResult) bool) {
		input := s3.ListBucketsInput{
			BucketRegion: aws.String(l.Meta().Region(ctx)),
			MaxBuckets:   aws.Int32(int32(request.Limit)),
		}
		for item, err := range listBuckets(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			bucketName := aws.ToString(item.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrBucket), bucketName)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(bucketName)
			rd.Set(names.AttrBucket, bucketName)

			tflog.Info(ctx, "Reading S3 Bucket")
			diags := resourceBucketRead(ctx, rd, l.Meta())
			if diags.HasError() {
				tflog.Error(ctx, "Reading S3 Bucket", map[string]any{
					names.AttrBucket: bucketName,
					"diags":          sdkdiag.DiagnosticsString(diags),
				})
				continue
			}
			if rd.Id() == "" {
				// Resource is logically deleted
				continue
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

type listBucketModel struct {
	framework.WithRegionModel
}

func listBuckets(ctx context.Context, conn *s3.Client, input *s3.ListBucketsInput) iter.Seq2[awstypes.Bucket, error] {
	return func(yield func(awstypes.Bucket, error) bool) {
		output, err := conn.ListBuckets(ctx, input)
		if err != nil {
			yield(awstypes.Bucket{}, fmt.Errorf("listing S3 Bucket resources: %w", err))
			return
		}

		for _, item := range output.Buckets {
			if !yield(item, nil) {
				return
			}
		}
	}
}
