//go:build !generate
// +build !generate

package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	ErrCodeNoSuchTagSet      = "NoSuchTagSet"
	ErrCodeNoSuchTagSetError = "NoSuchTagSetError"
)

// Custom S3 tag service update functions using the same format as generated code.

// BucketListTags lists S3 bucket tags.
// The identifier is the bucket name.
func BucketListTags(ctx context.Context, conn *s3.S3, identifier string) (tftags.KeyValueTags, error) {
	input := &s3.GetBucketTaggingInput{
		Bucket: aws.String(identifier),
	}

	output, err := conn.GetBucketTaggingWithContext(ctx, input)

	// S3 API Reference (https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketTagging.html)
	// lists the special error as NoSuchTagSetError, however the existing logic used NoSuchTagSet
	// and the AWS Go SDK has neither as a constant.
	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchTagSet, ErrCodeNoSuchTagSetError) {
		return tftags.New(nil), nil
	}

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTags(output.TagSet), nil
}

// BucketUpdateTags updates S3 bucket tags.
// The identifier is the bucket name.
func BucketUpdateTags(ctx context.Context, conn *s3.S3, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	// We need to also consider any existing ignored tags.
	allTags, err := BucketListTags(ctx, conn, identifier)

	if err != nil {
		return fmt.Errorf("error listing resource tags (%s): %w", identifier, err)
	}

	ignoredTags := allTags.Ignore(oldTags).Ignore(newTags)

	if len(newTags)+len(ignoredTags) > 0 {
		input := &s3.PutBucketTaggingInput{
			Bucket: aws.String(identifier),
			Tagging: &s3.Tagging{
				TagSet: Tags(newTags.Merge(ignoredTags)),
			},
		}

		_, err := conn.PutBucketTaggingWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error setting resource tags (%s): %w", identifier, err)
		}
	} else if len(oldTags) > 0 && len(ignoredTags) == 0 {
		input := &s3.DeleteBucketTaggingInput{
			Bucket: aws.String(identifier),
		}

		_, err := conn.DeleteBucketTaggingWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error deleting resource tags (%s): %w", identifier, err)
		}
	}

	return nil
}

// ObjectListTags lists S3 object tags.
func ObjectListTags(ctx context.Context, conn *s3.S3, bucket, key string) (tftags.KeyValueTags, error) {
	input := &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	var output *s3.GetObjectTaggingOutput

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		output, err = conn.GetObjectTaggingWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchKey) {
			return resource.RetryableError(fmt.Errorf("getting object tagging %s, retrying: %w", bucket, err))
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		output, err = conn.GetObjectTaggingWithContext(ctx, input)
	}

	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchTagSet, ErrCodeNoSuchTagSetError) {
		return tftags.New(nil), nil
	}

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTags(output.TagSet), nil
}

// ObjectUpdateTags updates S3 object tags.
func ObjectUpdateTags(ctx context.Context, conn *s3.S3, bucket, key string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	// We need to also consider any existing ignored tags.
	allTags, err := ObjectListTags(ctx, conn, bucket, key)

	if err != nil {
		return fmt.Errorf("error listing resource tags (%s/%s): %w", bucket, key, err)
	}

	ignoredTags := allTags.Ignore(oldTags).Ignore(newTags)

	if len(newTags)+len(ignoredTags) > 0 {
		input := &s3.PutObjectTaggingInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Tagging: &s3.Tagging{
				TagSet: Tags(newTags.Merge(ignoredTags).IgnoreAWS()),
			},
		}

		_, err := conn.PutObjectTaggingWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error setting resource tags (%s/%s): %w", bucket, key, err)
		}
	} else if len(oldTags) > 0 && len(ignoredTags) == 0 {
		input := &s3.DeleteObjectTaggingInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		}

		_, err := conn.DeleteObjectTaggingWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error deleting resource tags (%s/%s): %w", bucket, key, err)
		}
	}

	return nil
}
