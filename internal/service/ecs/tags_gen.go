// Code generated by internal/generate/tags/main.go; DO NOT EDIT.
package ecs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// findTag fetches an individual ecs service tag for a resource.
// Returns whether the key value and any errors. A NotFoundError is used to signal that no value was found.
// This function will optimise the handling over listTags, if possible.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func findTag(ctx context.Context, conn *ecs.Client, identifier, key string, optFns ...func(*ecs.Options)) (*string, error) {
	listTags, err := listTags(ctx, conn, identifier, optFns...)

	if err != nil {
		return nil, err
	}

	if !listTags.KeyExists(key) {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return listTags.KeyValue(key), nil
}

// listTags lists ecs service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func listTags(ctx context.Context, conn *ecs.Client, identifier string, optFns ...func(*ecs.Options)) (tftags.KeyValueTags, error) {
	input := ecs.ListTagsForResourceInput{
		ResourceArn: aws.String(identifier),
	}

	output, err := conn.ListTagsForResource(ctx, &input, optFns...)

	if tfawserr.ErrMessageContains(err, "InvalidParameterException", "The specified cluster is inactive. Specify an active cluster and try again.") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: &input,
		}
	}

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTags(ctx, output.Tags), nil
}

// ListTags lists ecs service tags and set them in Context.
// It is called from outside this package.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier string) error {
	tags, err := listTags(ctx, meta.(*conns.AWSClient).ECSClient(ctx), identifier)

	if err != nil {
		return err
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}

// []*SERVICE.Tag handling

// svcTags returns ecs service tags.
func svcTags(tags tftags.KeyValueTags) []awstypes.Tag {
	result := make([]awstypes.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := awstypes.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// keyValueTags creates tftags.KeyValueTags from ecs service tags.
func keyValueTags(ctx context.Context, tags []awstypes.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.ToString(tag.Key)] = tag.Value
	}

	return tftags.New(ctx, m)
}

// getTagsIn returns ecs service tags from Context.
// nil is returned if there are no input tags.
func getTagsIn(ctx context.Context) []awstypes.Tag {
	if inContext, ok := tftags.FromContext(ctx); ok {
		if tags := svcTags(inContext.TagsIn.UnwrapOrDefault()); len(tags) > 0 {
			return tags
		}
	}

	return nil
}

// setTagsOut sets ecs service tags in Context.
func setTagsOut(ctx context.Context, tags []awstypes.Tag) {
	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(keyValueTags(ctx, tags))
	}
}

// createTags creates ecs service tags for new resources.
func createTags(ctx context.Context, conn *ecs.Client, identifier string, tags []awstypes.Tag, optFns ...func(*ecs.Options)) error {
	if len(tags) == 0 {
		return nil
	}

	return updateTags(ctx, conn, identifier, nil, keyValueTags(ctx, tags), optFns...)
}

// updateTags updates ecs service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func updateTags(ctx context.Context, conn *ecs.Client, identifier string, oldTagsMap, newTagsMap any, optFns ...func(*ecs.Options)) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, identifier)

	removedTags := oldTags.Removed(newTags)
	removedTags = removedTags.IgnoreSystem(names.ECS)
	if len(removedTags) > 0 {
		input := ecs.UntagResourceInput{
			ResourceArn: aws.String(identifier),
			TagKeys:     removedTags.Keys(),
		}

		_, err := conn.UntagResource(ctx, &input, optFns...)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	updatedTags := oldTags.Updated(newTags)
	updatedTags = updatedTags.IgnoreSystem(names.ECS)
	if len(updatedTags) > 0 {
		input := ecs.TagResourceInput{
			ResourceArn: aws.String(identifier),
			Tags:        svcTags(updatedTags),
		}

		_, err := conn.TagResource(ctx, &input, optFns...)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// UpdateTags updates ecs service tags.
// It is called from outside this package.
func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier string, oldTags, newTags any) error {
	return updateTags(ctx, meta.(*conns.AWSClient).ECSClient(ctx), identifier, oldTags, newTags)
}
