// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
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
	l := listResourceBucketPolicy{}
	l.SetResourceSchema(resourceBucketPolicy())
	return &l
}

var _ list.ListResource = &listResourceBucketPolicy{}

type listResourceBucketPolicy struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *listResourceBucketPolicy) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query listBucketPolicyModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing S3 Bucket Policy")
	stream.Results = func(yield func(list.ListResult) bool) {
		tflog.Info(ctx, "Listing General Purpose Buckets")
		gpConn := l.Meta().S3Client(ctx)
		gpInput := s3.ListBucketsInput{
			BucketRegion: aws.String(l.Meta().Region(ctx)),
			MaxBuckets:   aws.Int32(int32(request.Limit)),
		}
		var count int64
		for result := range l.list(ctx, request, gpConn, listBuckets(ctx, gpConn, &gpInput)) {
			count++
			if !yield(result) {
				return
			}
		}

		limit := request.Limit - count
		if limit <= 0 {
			tflog.Info(ctx, "Limit reached, skipping Directory Buckets")
		}

		tflog.Info(ctx, "Listing Directory Buckets")
		dirConn := l.Meta().S3ExpressClient(ctx)
		dirInput := s3.ListDirectoryBucketsInput{
			MaxDirectoryBuckets: aws.Int32(int32(limit)),
		}
		for result := range l.list(ctx, request, dirConn, listDirectoryBuckets(ctx, dirConn, &dirInput)) {
			if !yield(result) {
				return
			}
		}
	}
}

func (l *listResourceBucketPolicy) list(ctx context.Context, request list.ListRequest, conn *s3.Client, buckets iter.Seq2[types.Bucket, error]) iter.Seq[list.ListResult] {
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

			// A Bucket Policy is optionally associated with a Bucket (1-0..1)
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
