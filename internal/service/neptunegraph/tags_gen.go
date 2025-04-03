// Code generated by internal/generate/tags/main.go; DO NOT EDIT.
package neptunegraph

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// listTags lists neptunegraph service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func listTags(ctx context.Context, conn *neptunegraph.Client, identifier string, optFns ...func(*neptunegraph.Options)) (tftags.KeyValueTags, error) {
	input := neptunegraph.ListTagsForResourceInput{
		ResourceArn: aws.String(identifier),
	}

	output, err := conn.ListTagsForResource(ctx, &input, optFns...)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTags(ctx, output.Tags), nil
}

// ListTags lists neptunegraph service tags and set them in Context.
// It is called from outside this package.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier string) error {
	tags, err := listTags(ctx, meta.(*conns.AWSClient).NeptuneGraphClient(ctx), identifier)

	if err != nil {
		return err
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}

// map[string]string handling

// svcTags returns neptunegraph service tags.
func svcTags(tags tftags.KeyValueTags) map[string]string {
	return tags.Map()
}

// keyValueTags creates tftags.KeyValueTags from neptunegraph service tags.
func keyValueTags(ctx context.Context, tags map[string]string) tftags.KeyValueTags {
	return tftags.New(ctx, tags)
}

// getTagsIn returns neptunegraph service tags from Context.
// nil is returned if there are no input tags.
func getTagsIn(ctx context.Context) map[string]string {
	if inContext, ok := tftags.FromContext(ctx); ok {
		if tags := svcTags(inContext.TagsIn.UnwrapOrDefault()); len(tags) > 0 {
			return tags
		}
	}

	return nil
}

// setTagsOut sets neptunegraph service tags in Context.
func setTagsOut(ctx context.Context, tags map[string]string) {
	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(keyValueTags(ctx, tags))
	}
}

// updateTags updates neptunegraph service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func updateTags(ctx context.Context, conn *neptunegraph.Client, identifier string, oldTagsMap, newTagsMap any, optFns ...func(*neptunegraph.Options)) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, identifier)

	removedTags := oldTags.Removed(newTags)
	removedTags = removedTags.IgnoreSystem(names.NeptuneGraph)
	if len(removedTags) > 0 {
		input := neptunegraph.UntagResourceInput{
			ResourceArn: aws.String(identifier),
			TagKeys:     removedTags.Keys(),
		}

		_, err := conn.UntagResource(ctx, &input, optFns...)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	updatedTags := oldTags.Updated(newTags)
	updatedTags = updatedTags.IgnoreSystem(names.NeptuneGraph)
	if len(updatedTags) > 0 {
		input := neptunegraph.TagResourceInput{
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

// UpdateTags updates neptunegraph service tags.
// It is called from outside this package.
func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier string, oldTags, newTags any) error {
	return updateTags(ctx, meta.(*conns.AWSClient).NeptuneGraphClient(ctx), identifier, oldTags, newTags)
}
