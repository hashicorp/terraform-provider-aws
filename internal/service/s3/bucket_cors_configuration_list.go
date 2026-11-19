// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_s3_bucket_cors_configuration")
func newBucketCORSConfigurationResourceAsListResource() inttypes.ListResourceForSDK {
	l := bucketCORSConfigurationListResource{}
	l.SetResourceSchema(resourceBucketCorsConfiguration())
	return &l
}

var _ list.ListResource = &bucketCORSConfigurationListResource{}

type bucketCORSConfigurationListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *bucketCORSConfigurationListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().S3Client(ctx)
	expectedBucketOwner := l.Meta().AccountID(ctx)

	tflog.Info(ctx, "Listing S3 Bucket CORS Configurations")

	stream.Results = func(yield func(list.ListResult) bool) {
		input := s3.ListBucketsInput{
			BucketRegion: aws.String(l.Meta().Region(ctx)),
		}
		for bucket, err := range listBuckets(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing S3 Bucket CORS Configuration resources: %w", err)))
				return
			}

			bucketName := aws.ToString(bucket.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrBucket), bucketName)

			corsRules, err := findCORSRules(ctx, conn, bucketName, expectedBucketOwner)
			if retry.NotFound(err) {
				tflog.Debug(ctx, "Bucket has no CORS configuration, skipping")
				continue
			}
			if err != nil {
				tflog.Error(ctx, "Reading S3 Bucket CORS Configuration", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(bucketName)
			rd.Set(names.AttrBucket, bucketName)

			if request.IncludeResource {
				if err := resourceBucketCorsConfigurationFlatten(bucketName, corsRules, rd); err != nil {
					tflog.Error(ctx, "Flattening S3 Bucket CORS Configuration", map[string]any{
						"error": err.Error(),
					})
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
