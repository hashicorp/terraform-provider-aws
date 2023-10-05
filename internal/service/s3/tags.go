// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !generate
// +build !generate

package s3

import (
	"context"
	"fmt"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	s3_sdkv2 "github.com/aws/aws-sdk-go-v2/service/s3"
	s3types_sdkv2 "github.com/aws/aws-sdk-go-v2/service/s3/types"
	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	s3_sdkv1 "github.com/aws/aws-sdk-go/service/s3"
	s3iface_sdkv1 "github.com/aws/aws-sdk-go/service/s3/s3iface"
	tfawserr_sdkv1 "github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	tfawserr_sdkv2 "github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// Custom S3 tag service update functions using the same format as generated code.

// BucketListTags lists S3 bucket tags.
// The identifier is the bucket name.
func BucketListTags(ctx context.Context, conn s3iface_sdkv1.S3API, identifier string) (tftags.KeyValueTags, error) {
	input := &s3_sdkv1.GetBucketTaggingInput{
		Bucket: aws_sdkv1.String(identifier),
	}

	output, err := conn.GetBucketTaggingWithContext(ctx, input)

	// S3 API Reference (https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketTagging.html)
	// lists the special error as NoSuchTagSetError, however the existing logic used NoSuchTagSet
	// and the AWS Go SDK has neither as a constant.
	if tfawserr_sdkv1.ErrCodeEquals(err, errCodeNoSuchTagSet, errCodeNoSuchTagSetError) {
		return tftags.New(ctx, nil), nil
	}

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, output.TagSet), nil
}

// BucketUpdateTags updates S3 bucket tags.
// The identifier is the bucket name.
func BucketUpdateTags(ctx context.Context, conn s3iface_sdkv1.S3API, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	// We need to also consider any existing ignored tags.
	allTags, err := BucketListTags(ctx, conn, identifier)

	if err != nil {
		return fmt.Errorf("listing resource tags (%s): %w", identifier, err)
	}

	ignoredTags := allTags.Ignore(oldTags).Ignore(newTags)

	if len(newTags)+len(ignoredTags) > 0 {
		input := &s3_sdkv1.PutBucketTaggingInput{
			Bucket: aws_sdkv1.String(identifier),
			Tagging: &s3_sdkv1.Tagging{
				TagSet: Tags(newTags.Merge(ignoredTags)),
			},
		}

		_, err := conn.PutBucketTaggingWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("setting resource tags (%s): %w", identifier, err)
		}
	} else if len(oldTags) > 0 && len(ignoredTags) == 0 {
		input := &s3_sdkv1.DeleteBucketTaggingInput{
			Bucket: aws_sdkv1.String(identifier),
		}

		_, err := conn.DeleteBucketTaggingWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("deleting resource tags (%s): %w", identifier, err)
		}
	}

	return nil
}

// ObjectListTags lists S3 object tags.
func ObjectListTags(ctx context.Context, conn *s3_sdkv2.Client, bucket, key string) (tftags.KeyValueTags, error) {
	input := &s3_sdkv2.GetObjectTaggingInput{
		Bucket: aws_sdkv2.String(bucket),
		Key:    aws_sdkv2.String(key),
	}

	output, err := conn.GetObjectTagging(ctx, input)

	if tfawserr_sdkv2.ErrCodeEquals(err, errCodeNoSuchTagSet, errCodeNoSuchTagSetError) {
		return tftags.New(ctx, nil), nil
	}

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTagsV2(ctx, output.TagSet), nil
}

// ObjectUpdateTags updates S3 object tags.
func ObjectUpdateTags(ctx context.Context, conn *s3_sdkv2.Client, bucket, key string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	// We need to also consider any existing ignored tags.
	allTags, err := ObjectListTags(ctx, conn, bucket, key)

	if err != nil {
		return fmt.Errorf("listing resource tags (%s/%s): %w", bucket, key, err)
	}

	ignoredTags := allTags.Ignore(oldTags).Ignore(newTags)

	if len(newTags)+len(ignoredTags) > 0 {
		input := &s3_sdkv2.PutObjectTaggingInput{
			Bucket: aws_sdkv2.String(bucket),
			Key:    aws_sdkv2.String(key),
			Tagging: &s3types_sdkv2.Tagging{
				TagSet: tagsV2(newTags.Merge(ignoredTags)),
			},
		}

		_, err := conn.PutObjectTagging(ctx, input)

		if err != nil {
			return fmt.Errorf("setting resource tags (%s/%s): %w", bucket, key, err)
		}
	} else if len(oldTags) > 0 && len(ignoredTags) == 0 {
		input := &s3_sdkv2.DeleteObjectTaggingInput{
			Bucket: aws_sdkv2.String(bucket),
			Key:    aws_sdkv2.String(key),
		}

		_, err := conn.DeleteObjectTagging(ctx, input)

		if err != nil {
			return fmt.Errorf("deleting resource tags (%s/%s): %w", bucket, key, err)
		}
	}

	return nil
}
