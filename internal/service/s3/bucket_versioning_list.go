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

// @SDKListResource("aws_s3_bucket_versioning")
func newBucketVersioningResourceAsListResource() inttypes.ListResourceForSDK {
	return newListResourceBaseBucketProperty(
		resourceBucketVersioning(),
		newBucketVersioningListHandler,
	)
}

type listBucketVersioningModel struct {
	framework.WithRegionModel
}

var _ bucketPropertyListHandlerSDK = bucketVersioningListHandler{}

func newBucketVersioningListHandler(lister listResourceSDK) bucketPropertyListHandlerSDK {
	return bucketVersioningListHandler{
		baseBucketPropertyListHandlerSDK: newBaseBucketPropertyListHandlerSDK(lister),
	}
}

type bucketVersioningListHandler struct {
	baseBucketPropertyListHandlerSDK
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
