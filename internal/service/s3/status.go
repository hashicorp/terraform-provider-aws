// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func lifecycleConfigurationRulesStatus(ctx context.Context, conn *s3.S3, bucket, expectedBucketOwner string, rules []*s3.LifecycleRule) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &s3.GetBucketLifecycleConfigurationInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketLifecycleConfigurationWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchLifecycleConfiguration, s3.ErrCodeNoSuchBucket) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", &retry.NotFoundError{
				Message:     "Empty result",
				LastRequest: input,
			}
		}

		for _, expectedRule := range rules {
			found := false

			for _, actualRule := range output.Rules {
				if aws.StringValue(actualRule.ID) != aws.StringValue(expectedRule.ID) {
					continue
				}
				found = true
				if aws.StringValue(actualRule.Status) != aws.StringValue(expectedRule.Status) {
					return output, LifecycleConfigurationRulesStatusNotReady, nil
				}
			}

			if !found {
				return output, LifecycleConfigurationRulesStatusNotReady, nil
			}
		}

		return output, LifecycleConfigurationRulesStatusReady, nil
	}
}

func bucketVersioningStatus(ctx context.Context, conn *s3.S3, bucket, expectedBucketOwner string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &s3.GetBucketVersioningInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketVersioningWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", &retry.NotFoundError{
				Message:     "Empty result",
				LastRequest: input,
			}
		}

		if output.Status == nil {
			return output, BucketVersioningStatusDisabled, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
