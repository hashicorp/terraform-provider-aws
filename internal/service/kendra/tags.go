// Custom Kendra tag service update functions using the same format as generated code.
// Modified to support AWS Go SDK v2.

package kendra

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// ListTags lists kendra service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(ctx context.Context, conn *kendra.Client, identifier string) (tftags.KeyValueTags, error) {
	input := &kendra.ListTagsForResourceInput{
		ResourceARN: aws.String(identifier),
	}

	output, err := conn.ListTagsForResource(ctx, input)

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTags(output.Tags), nil
}

// []*SERVICE.Tag handling

// Tags returns kendra service tags.
func Tags(tags tftags.KeyValueTags) []types.Tag {
	result := make([]types.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from kendra service tags.
func KeyValueTags(tags []types.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.ToString(tag.Key)] = tag.Value
	}

	return tftags.New(m)
}

// UpdateTags updates kendra service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func UpdateTags(ctx context.Context, conn *kendra.Client, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &kendra.UntagResourceInput{
			ResourceARN: aws.String(identifier),
			TagKeys:     removedTags.IgnoreAWS().Keys(),
		}

		_, err := conn.UntagResource(ctx, input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &kendra.TagResourceInput{
			ResourceARN: aws.String(identifier),
			Tags:        Tags(updatedTags.IgnoreAWS()),
		}

		_, err := conn.TagResource(ctx, input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
