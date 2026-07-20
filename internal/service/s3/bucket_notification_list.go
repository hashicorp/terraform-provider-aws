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
// @SDKListResource("aws_s3_bucket_notification")
func newBucketNotificationResourceAsListResource() inttypes.ListResourceForSDK {
	l := bucketNotificationListResource{}
	l.SetResourceSchema(resourceBucketNotification())
	return &l
}

var _ list.ListResource = &bucketNotificationListResource{}

type bucketNotificationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *bucketNotificationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().S3Client(ctx)

	var query listBucketNotificationModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing S3 Bucket Notifications")

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

			// Directory buckets do not support standard notification configuration.
			if isDirectoryBucket(bucketName) {
				continue
			}

			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrBucket), bucketName)

			output, err := findBucketNotificationConfiguration(ctx, conn, bucketName, "")
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				tflog.Error(ctx, "Reading S3 Bucket Notification", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(bucketName)
			rd.Set(names.AttrBucket, bucketName)

			if request.IncludeResource {
				if diags := resourceBucketNotificationFlatten(output, rd); diags.HasError() {
					tflog.Error(ctx, "Flattening S3 Bucket Notification", map[string]any{
						"diags": sdkdiag.DiagnosticsString(diags),
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

type listBucketNotificationModel struct {
	framework.WithRegionModel
}
