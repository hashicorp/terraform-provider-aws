// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !generate

package kinesis

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Custom Kinesis tag functions using the same format as generated code.

// listTags lists kinesis service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func listTags(ctx context.Context, conn *kinesis.Client, identifier string, optFns ...func(*kinesis.Options)) (tftags.KeyValueTags, error) {
	input := kinesis.ListTagsForResourceInput{
		ResourceARN: aws.String(identifier),
	}

	output, err := conn.ListTagsForResource(ctx, &input, optFns...)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTags(ctx, output.Tags), nil
}

// updateTags updates kinesis service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func updateTags(ctx context.Context, conn *kinesis.Client, identifier string, oldTagsMap, newTagsMap any, optFns ...func(*kinesis.Options)) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, identifier)

	removedTags := oldTags.Removed(newTags)
	removedTags = removedTags.IgnoreSystem(names.Kinesis)
	if len(removedTags) > 0 {
		input := kinesis.UntagResourceInput{
			ResourceARN: aws.String(identifier),
			TagKeys:     removedTags.Keys(),
		}

		_, err := conn.UntagResource(ctx, &input, optFns...)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	updatedTags := oldTags.Updated(newTags)
	updatedTags = updatedTags.IgnoreSystem(names.Kinesis)
	if len(updatedTags) > 0 {
		input := kinesis.TagResourceInput{
			ResourceARN: aws.String(identifier),
			Tags:        updatedTags.Map(),
		}

		_, err := conn.TagResource(ctx, &input, optFns...)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// listTagsForStream lists kinesis stream tags.
// The identifier is the stream name.
func listTagsForStream(ctx context.Context, conn *kinesis.Client, identifier string, optFns ...func(*kinesis.Options)) (tftags.KeyValueTags, error) {
	input := kinesis.ListTagsForStreamInput{
		StreamName: aws.String(identifier),
	}

	var output []awstypes.Tag

	err := listTagsForStreamPages(ctx, conn, &input, func(page *kinesis.ListTagsForStreamOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Tags...)

		return !lastPage
	}, optFns...)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTags(ctx, output), nil
}

// updateTagsForStream updates kinesis stream tags.
// The identifier is the stream name.
func updateTagsForStream(ctx context.Context, conn *kinesis.Client, identifier string, oldTagsMap, newTagsMap any, optFns ...func(*kinesis.Options)) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, identifier)

	const (
		chunkSize = 10
	)
	removedTags := oldTags.Removed(newTags)
	removedTags = removedTags.IgnoreSystem(names.Kinesis)
	if len(removedTags) > 0 {
		for _, removedTags := range removedTags.Chunks(chunkSize) {
			input := kinesis.RemoveTagsFromStreamInput{
				StreamName: aws.String(identifier),
				TagKeys:    removedTags.Keys(),
			}

			_, err := conn.RemoveTagsFromStream(ctx, &input, optFns...)

			if err != nil {
				return fmt.Errorf("untagging resource (%s): %w", identifier, err)
			}
		}
	}

	updatedTags := oldTags.Updated(newTags)
	updatedTags = updatedTags.IgnoreSystem(names.Kinesis)
	if len(updatedTags) > 0 {
		for _, updatedTags := range updatedTags.Chunks(chunkSize) {
			input := kinesis.AddTagsToStreamInput{
				StreamName: aws.String(identifier),
				Tags:       updatedTags.IgnoreAWS().Map(),
			}

			_, err := conn.AddTagsToStream(ctx, &input, optFns...)

			if err != nil {
				return fmt.Errorf("tagging resource (%s): %w", identifier, err)
			}
		}
	}

	return nil
}

// ListTags lists kinesis service tags and set them in Context.
// It is called from outside this package.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier, resourceType string) error {
	var (
		tags tftags.KeyValueTags
		err  error
	)
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)

	switch resourceType {
	case "Stream":
		tags, err = listTagsForStream(ctx, conn, identifier)

	default:
		tags, err = listTags(ctx, conn, identifier)
	}

	if err != nil {
		return err
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}

// UpdateTags updates kinesis service tags.
// It is called from outside this package.
func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier, resourceType string, oldTags, newTags any) error {
	conn := meta.(*conns.AWSClient).KinesisClient(ctx)

	switch resourceType {
	case "Stream":
		return updateTagsForStream(ctx, conn, identifier, oldTags, newTags)

	default:
		return updateTags(ctx, conn, identifier, oldTags, newTags)
	}
}
