//go:build !generate
// +build !generate

package s3control

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// Custom S3control tagging functions using similar formatting as other service generated code.

// bucketListTags lists S3control bucket tags.
// The identifier is the bucket ARN.
func bucketListTags(conn *s3control.S3Control, identifier string) (tftags.KeyValueTags, error) {
	parsedArn, err := arn.Parse(identifier)

	if err != nil {
		return tftags.New(nil), fmt.Errorf("error parsing S3 Control Bucket ARN (%s): %w", identifier, err)
	}

	input := &s3control.GetBucketTaggingInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(identifier),
	}

	output, err := conn.GetBucketTagging(input)

	if tfawserr.ErrCodeEquals(err, "NoSuchTagSet") {
		return tftags.New(nil), nil
	}

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTags(output.TagSet), nil
}

// bucketUpdateTags updates S3control bucket tags.
// The identifier is the bucket ARN.
func bucketUpdateTags(conn *s3control.S3Control, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	parsedArn, err := arn.Parse(identifier)

	if err != nil {
		return fmt.Errorf("error parsing S3 Control Bucket ARN (%s): %w", identifier, err)
	}

	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	// We need to also consider any existing ignored tags.
	allTags, err := bucketListTags(conn, identifier)

	if err != nil {
		return fmt.Errorf("error listing resource tags (%s): %w", identifier, err)
	}

	ignoredTags := allTags.Ignore(oldTags).Ignore(newTags)

	if len(newTags)+len(ignoredTags) > 0 {
		input := &s3control.PutBucketTaggingInput{
			AccountId: aws.String(parsedArn.AccountID),
			Bucket:    aws.String(identifier),
			Tagging: &s3control.Tagging{
				TagSet: Tags(newTags.Merge(ignoredTags)),
			},
		}

		_, err := conn.PutBucketTagging(input)

		if err != nil {
			return fmt.Errorf("error setting resource tags (%s): %w", identifier, err)
		}
	} else if len(oldTags) > 0 && len(ignoredTags) == 0 {
		input := &s3control.DeleteBucketTaggingInput{
			AccountId: aws.String(parsedArn.AccountID),
			Bucket:    aws.String(identifier),
		}

		_, err := conn.DeleteBucketTagging(input)

		if err != nil {
			return fmt.Errorf("error deleting resource tags (%s): %w", identifier, err)
		}
	}

	return nil
}

func StorageLensTags(tags tftags.KeyValueTags) []*s3control.StorageLensTag {
	result := make([]*s3control.StorageLensTag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &s3control.StorageLensTag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

func KeyValueTagsFromStorageLensTags(tags []*s3control.StorageLensTag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(m)
}

func storageLensConfigurationListTags(ctx context.Context, conn *s3control.S3Control, accountID, configID string) (tftags.KeyValueTags, error) {
	input := &s3control.GetStorageLensConfigurationTaggingInput{
		AccountId: aws.String(accountID),
		ConfigId:  aws.String(configID),
	}

	output, err := conn.GetStorageLensConfigurationTaggingWithContext(ctx, input)

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTagsFromStorageLensTags(output.Tags), nil
}

func storageLensConfigurationUpdateTags(ctx context.Context, conn *s3control.S3Control, accountID, configID string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	// We need to also consider any existing ignored tags.
	allTags, err := storageLensConfigurationListTags(ctx, conn, accountID, configID)

	if err != nil {
		return fmt.Errorf("listing tags: %s", err)
	}

	ignoredTags := allTags.Ignore(oldTags).Ignore(newTags)

	if len(newTags)+len(ignoredTags) > 0 {
		input := &s3control.PutStorageLensConfigurationTaggingInput{
			AccountId: aws.String(accountID),
			ConfigId:  aws.String(configID),
			Tags:      StorageLensTags(newTags.Merge(ignoredTags)),
		}

		_, err := conn.PutStorageLensConfigurationTaggingWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("setting tags: %s", err)
		}
	} else if len(oldTags) > 0 && len(ignoredTags) == 0 {
		input := &s3control.DeleteStorageLensConfigurationTaggingInput{
			AccountId: aws.String(accountID),
			ConfigId:  aws.String(configID),
		}

		_, err := conn.DeleteStorageLensConfigurationTaggingWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("deleting tags: %s", err)
		}
	}

	return nil
}
