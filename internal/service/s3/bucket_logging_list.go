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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_s3_bucket_logging")
func newBucketLoggingResourceAsListResource() inttypes.ListResourceForSDK {
	l := bucketLoggingListResource{}
	l.SetResourceSchema(resourceBucketLogging())
	return &l
}

var _ list.ListResource = &bucketLoggingListResource{}

type bucketLoggingListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *bucketLoggingListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().S3Client(ctx)

	var query listBucketLoggingModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing S3 Bucket Logging")

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

			loggingEnabled, err := findLoggingEnabled(ctx, conn, bucketName, "")
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				tflog.Error(ctx, "Reading S3 Bucket Logging", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(bucketName)
			rd.Set(names.AttrBucket, bucketName)

			if request.IncludeResource {
				if err := resourceBucketLoggingFlatten(loggingEnabled, rd); err != nil {
					tflog.Error(ctx, "Flattening S3 Bucket Logging", map[string]any{
						"diags": sdkdiag.DiagnosticsString(sdkdiag.AppendFromErr(nil, err)),
					})
					continue
				}
				if rd.Id() == "" {
					tflog.Warn(ctx, "Resource disappeared during listing, skipping")
					continue
				}
			}

			result.DisplayName = bucketName

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
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

type listBucketLoggingModel struct {
	framework.WithRegionModel
}
