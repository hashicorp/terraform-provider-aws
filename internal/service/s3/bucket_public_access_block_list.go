// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_s3_bucket_public_access_block")
func newBucketPublicAccessBlockResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceBucketPublicAccessBlock{}
	l.SetResourceSchema(resourceBucketPublicAccessBlock())
	return &l
}

var _ list.ListResource = &listResourceBucketPublicAccessBlock{}

type listResourceBucketPublicAccessBlock struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceBucketPublicAccessBlock) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().S3Client(ctx)

	var query listBucketPublicAccessBlockModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing S3 Bucket Public Access Block")
	stream.Results = func(yield func(list.ListResult) bool) {
		input := s3.ListBucketsInput{
			BucketRegion: aws.String(l.Meta().Region(ctx)),
			MaxBuckets:   aws.Int32(int32(request.Limit)),
		}
		for bucket, err := range listBuckets(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			bucketName := aws.ToString(bucket.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrBucket), bucketName)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(bucketName)
			rd.Set(names.AttrBucket, bucketName)

			// A Bucket Policy is optionally associated with a Bucket (1-0..1)
			// So always try to read it to see if it is present.
			tflog.Info(ctx, "Reading S3 Bucket Public Access Block")
			diags := resourceBucketPublicAccessBlockRead(ctx, rd, l.Meta())
			if diags.HasError() {
				tflog.Error(ctx, "Reading S3 Bucket Public Access Block", map[string]any{
					"diags": sdkdiag.DiagnosticsString(diags),
				})
				continue
			}
			if rd.Id() == "" {
				tflog.Warn(ctx, "Resource disappeared during listing, skipping")
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

type listBucketPublicAccessBlockModel struct {
	framework.WithRegionModel
}
