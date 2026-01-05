// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// Custom S3 tag functions using the same format as generated code.

func bucketCreateTags(ctx context.Context, conn *s3control.Client, identifier string, tags []awstypes.S3Tag) error {
	if len(tags) == 0 {
		return nil
	}

	return bucketUpdateTags(ctx, conn, identifier, nil, keyValueTagsFromS3Tags(ctx, tags))
}

// bucketListTags lists S3control bucket tags.
// The identifier is the bucket ARN.
func bucketListTags(ctx context.Context, conn *s3control.Client, identifier string, optFns ...func(*s3control.Options)) (tftags.KeyValueTags, error) {
	parsedArn, err := arn.Parse(identifier)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	input := s3control.GetBucketTaggingInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(identifier),
	}

	output, err := conn.GetBucketTagging(ctx, &input, optFns...)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchTagSet) {
		return tftags.New(ctx, nil), nil
	}

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTagsFromS3Tags(ctx, output.TagSet), nil
}

// bucketUpdateTags updates S3control bucket tags.
// The identifier is the bucket ARN.
func bucketUpdateTags(ctx context.Context, conn *s3control.Client, identifier string, oldTagsMap, newTagsMap any, optFns ...func(*s3control.Options)) error {
	parsedArn, err := arn.Parse(identifier)

	if err != nil {
		return err
	}

	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	// We need to also consider any existing ignored tags.
	allTags, err := bucketListTags(ctx, conn, identifier)

	if err != nil {
		return fmt.Errorf("listing resource tags (%s): %w", identifier, err)
	}

	ignoredTags := allTags.Ignore(oldTags).Ignore(newTags)

	if len(newTags)+len(ignoredTags) > 0 {
		input := s3control.PutBucketTaggingInput{
			AccountId: aws.String(parsedArn.AccountID),
			Bucket:    aws.String(identifier),
			Tagging: &awstypes.Tagging{
				TagSet: svcS3Tags(newTags.Merge(ignoredTags)),
			},
		}

		_, err := conn.PutBucketTagging(ctx, &input, optFns...)

		if err != nil {
			return fmt.Errorf("setting resource tags (%s): %w", identifier, err)
		}
	} else if len(oldTags) > 0 && len(ignoredTags) == 0 {
		input := s3control.DeleteBucketTaggingInput{
			AccountId: aws.String(parsedArn.AccountID),
			Bucket:    aws.String(identifier),
		}

		_, err := conn.DeleteBucketTagging(ctx, &input, optFns...)

		if err != nil {
			return fmt.Errorf("deleting resource tags (%s): %w", identifier, err)
		}
	}

	return nil
}
