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

// @SDKListResource("aws_s3_bucket_server_side_encryption_configuration")
func newBucketServerSideEncryptionConfigurationResourceAsListResource() inttypes.ListResourceForSDK {
	return newListResourceBaseBucketProperty(
		resourceBucketServerSideEncryptionConfiguration(),
		newBucketServerSideEncryptionConfigurationListHandler,
	)
}

type listBucketServerSideEncryptionConfigurationModel struct {
	framework.WithRegionModel
}

var _ bucketPropertyListHandlerSDK = bucketServerSideEncryptionConfigurationListHandler{}

func newBucketServerSideEncryptionConfigurationListHandler(lister listResourceSDK) bucketPropertyListHandlerSDK {
	return bucketServerSideEncryptionConfigurationListHandler{
		baseBucketPropertyListHandlerSDK: newBaseBucketPropertyListHandlerSDK(lister),
	}
}

type bucketServerSideEncryptionConfigurationListHandler struct {
	baseBucketPropertyListHandlerSDK
}

func (l bucketServerSideEncryptionConfigurationListHandler) parseQuery(ctx context.Context, config tfsdk.Config) (diags diag.Diagnostics) {
	return parseQuery[listBucketServerSideEncryptionConfigurationModel](ctx, config)
}

func (l bucketServerSideEncryptionConfigurationListHandler) list(ctx context.Context, request list.ListRequest, conn *s3.Client, buckets iter.Seq2[awstypes.Bucket, error]) iter.Seq[list.ListResult] {
	return func(yield func(list.ListResult) bool) {
		for bucket, err := range buckets {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing S3 Bucket Server Side Encryption Configuration resources: %w", err))
				yield(result)
				return
			}

			bucketName := aws.ToString(bucket.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrBucket), bucketName)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(bucketName)
			rd.Set(names.AttrBucket, bucketName)

			tflog.Info(ctx, "Reading S3 Bucket Server Side Encryption Configuration")
			var expectedOwner string
			if isGeneralPurposeBucket(bucketName) {
				expectedOwner = l.Meta().AccountID(ctx)
			}
			sse, err := findServerSideEncryptionConfiguration(ctx, conn, bucketName, expectedOwner)
			if retry.NotFound(err) {
				tflog.Debug(ctx, "Bucket has no Server Side Encryption Configuration, skipping")
				continue
			}
			if err != nil {
				tflog.Error(ctx, "Reading S3 Bucket Server Side Encryption Configuration", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			if err := resourceBucketServerSideEncryptionConfigurationFlatten(ctx, sse, rd); err != nil {
				tflog.Error(ctx, "Reading S3 Bucket Server Side Encryption Configuration", map[string]any{
					"error": err.Error(),
				})
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
