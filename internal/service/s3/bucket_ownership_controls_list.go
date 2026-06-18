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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_s3_bucket_ownership_controls")
func newBucketOwnershipControlsResourceAsListResource() inttypes.ListResourceForSDK {
	return newListResourceBaseBucketPropertySDK(
		resourceBucketOwnershipControls(),
		newBucketOwnershipControlsListHandler,
	)
}

var _ bucketPropertyListHandlerSDK = bucketOwnershipControlsListHandler{}

func newBucketOwnershipControlsListHandler(lister listResourceSDK) bucketPropertyListHandlerSDK {
	return bucketOwnershipControlsListHandler{
		baseBucketPropertyListHandlerSDK: newBaseBucketPropertyListHandlerSDK(lister),
	}
}

type bucketOwnershipControlsListHandler struct {
	baseBucketPropertyListHandlerSDK
}

func (l bucketOwnershipControlsListHandler) parseQuery(ctx context.Context, config tfsdk.Config) (diags diag.Diagnostics) {
	return parseQuery[listBucketOwnershipControlsModel](ctx, config)
}

func (l bucketOwnershipControlsListHandler) list(ctx context.Context, request list.ListRequest, conn *s3.Client, buckets iter.Seq2[awstypes.Bucket, error]) iter.Seq[list.ListResult] {
	return func(yield func(list.ListResult) bool) {
		for bucket, err := range buckets {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing S3 Ownership Controls resources: %w", err))
				yield(result)
				return
			}

			bucketName := aws.ToString(bucket.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrBucket), bucketName)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(bucketName)
			rd.Set(names.AttrBucket, bucketName)

			// There is always a Bucket Ownership Control associated with a Bucket (1:1)
			// So only read it if resource data is requested.
			if request.IncludeResource {
				tflog.Info(ctx, "Reading S3 Ownership Controls")
				policy, err := findBucketOwnershipControls(ctx, conn, bucketName)
				if retry.NotFound(err) {
					tflog.Warn(ctx, "Resource disappeared during listing, skipping")
					continue
				}
				if err != nil {
					tflog.Error(ctx, "Reading S3 Ownership Controls", map[string]any{
						"error": err.Error(),
					})
					continue
				}

				if err := resourceBucketOwnershipControlsFlatten(ctx, policy, rd); err != nil {
					tflog.Error(ctx, "Reading S3 Ownership Controls", map[string]any{
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

type listBucketOwnershipControlsModel struct {
	framework.WithRegionModel
}
