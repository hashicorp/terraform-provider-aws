// Code generated by internal/generate/tags/main.go; DO NOT EDIT.
package storagegateway

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/aws/aws-sdk-go/service/storagegateway/storagegatewayiface"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
)

// ListTags lists storagegateway service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(ctx context.Context, conn storagegatewayiface.StorageGatewayAPI, identifier string) (tftags.KeyValueTags, error) {
	input := &storagegateway.ListTagsForResourceInput{
		ResourceARN: aws.String(identifier),
	}

	output, err := conn.ListTagsForResourceWithContext(ctx, input)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, output.Tags), nil
}

func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier string) error {
	tags, err := ListTags(ctx, meta.(*conns.AWSClient).StorageGatewayConn(), identifier)

	if err != nil {
		return err
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = types.Some(tags)
	}

	return nil
}

// []*SERVICE.Tag handling

// Tags returns storagegateway service tags.
func Tags(tags tftags.KeyValueTags) []*storagegateway.Tag {
	result := make([]*storagegateway.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &storagegateway.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from storagegateway service tags.
func KeyValueTags(ctx context.Context, tags []*storagegateway.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(ctx, m)
}

// GetTagsIn returns storagegateway service tags from Context.
// nil is returned if there are no input tags.
func GetTagsIn(ctx context.Context) []*storagegateway.Tag {
	if inContext, ok := tftags.FromContext(ctx); ok {
		if tags := Tags(inContext.TagsIn); len(tags) > 0 {
			return tags
		}
	}

	return nil
}

// SetTagsOut sets storagegateway service tags in Context.
func SetTagsOut(ctx context.Context, tags []*storagegateway.Tag) {
	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = types.Some(KeyValueTags(ctx, tags))
	}
}

// UpdateTags updates storagegateway service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.

func UpdateTags(ctx context.Context, conn storagegatewayiface.StorageGatewayAPI, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &storagegateway.RemoveTagsFromResourceInput{
			ResourceARN: aws.String(identifier),
			TagKeys:     aws.StringSlice(removedTags.IgnoreAWS().Keys()),
		}

		_, err := conn.RemoveTagsFromResourceWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &storagegateway.AddTagsToResourceInput{
			ResourceARN: aws.String(identifier),
			Tags:        Tags(updatedTags.IgnoreAWS()),
		}

		_, err := conn.AddTagsToResourceWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier string, oldTags, newTags any) error {
	return UpdateTags(ctx, meta.(*conns.AWSClient).StorageGatewayConn(), identifier, oldTags, newTags)
}
