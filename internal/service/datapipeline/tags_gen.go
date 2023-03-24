// Code generated by internal/generate/tags/main.go; DO NOT EDIT.
package datapipeline

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/aws/aws-sdk-go/service/datapipeline/datapipelineiface"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
)

// []*SERVICE.Tag handling

// Tags returns datapipeline service tags.
func Tags(tags tftags.KeyValueTags) []*datapipeline.Tag {
	result := make([]*datapipeline.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &datapipeline.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from datapipeline service tags.
func KeyValueTags(ctx context.Context, tags []*datapipeline.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(ctx, m)
}

// GetTagsIn returns datapipeline service tags from Context.
// nil is returned if there are no input tags.
func GetTagsIn(ctx context.Context) []*datapipeline.Tag {
	if inContext, ok := tftags.FromContext(ctx); ok {
		if tags := Tags(inContext.TagsIn); len(tags) > 0 {
			return tags
		}
	}

	return nil
}

// SetTagsOut sets datapipeline service tags in Context.
func SetTagsOut(ctx context.Context, tags []*datapipeline.Tag) {
	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = types.Some(KeyValueTags(ctx, tags))
	}
}

// UpdateTags updates datapipeline service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.

func UpdateTags(ctx context.Context, conn datapipelineiface.DataPipelineAPI, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &datapipeline.RemoveTagsInput{
			PipelineId: aws.String(identifier),
			TagKeys:    aws.StringSlice(removedTags.IgnoreAWS().Keys()),
		}

		_, err := conn.RemoveTagsWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &datapipeline.AddTagsInput{
			PipelineId: aws.String(identifier),
			Tags:       Tags(updatedTags.IgnoreAWS()),
		}

		_, err := conn.AddTagsWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier string, oldTags, newTags any) error {
	return UpdateTags(ctx, meta.(*conns.AWSClient).DataPipelineConn(), identifier, oldTags, newTags)
}
