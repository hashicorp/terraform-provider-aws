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

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_s3_bucket_policy")
func newBucketPolicyResourceAsListResource() inttypes.ListResourceForSDK {
	return newListResourceBaseBucketProperty(
		resourceBucketPolicy(),
		newBucketPolicyListHandler,
	)
}

var _ bucketPropertyListHandlerSDK = bucketPolicyListHandler{}

func newBucketPolicyListHandler(lister listResourceSDK) bucketPropertyListHandlerSDK {
	return bucketPolicyListHandler{
		baseBucketPropertyListHandlerSDK: newBaseBucketPropertyListHandlerSDK(lister),
	}
}

type bucketPolicyListHandler struct {
	baseBucketPropertyListHandlerSDK
}

func (l bucketPolicyListHandler) parseQuery(ctx context.Context, config tfsdk.Config) (diags diag.Diagnostics) {
	return parseQuery[listBucketPolicyModel](ctx, config)
}

func (l bucketPolicyListHandler) list(ctx context.Context, request list.ListRequest, conn *s3.Client, buckets iter.Seq2[awstypes.Bucket, error]) iter.Seq[list.ListResult] {
	return func(yield func(list.ListResult) bool) {
		for bucket, err := range buckets {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(fmt.Errorf("listing S3 Bucket Policy resources: %w", err))
				yield(result)
				return
			}

			bucketName := aws.ToString(bucket.Name)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrBucket), bucketName)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(bucketName)
			rd.Set(names.AttrBucket, bucketName)

			// A Bucket Policy is optionally associated with a Bucket (1:0..1)
			// So always try to read it to see if it is present.
			tflog.Info(ctx, "Reading S3 Bucket Policy")
			policy, err := findBucketPolicy(ctx, conn, bucketName)
			if retry.NotFound(err) {
				tflog.Debug(ctx, "Bucket has no policy, skipping")
				continue
			}
			if err != nil {
				tflog.Error(ctx, "Reading S3 Bucket Policy", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			if err := resourceBucketPolicyFlatten(ctx, policy, rd); err != nil {
				tflog.Error(ctx, "Reading S3 Bucket Policy", map[string]any{
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

type listBucketPolicyModel struct {
	framework.WithRegionModel
}
