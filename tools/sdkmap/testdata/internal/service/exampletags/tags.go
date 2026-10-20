// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package exampletags

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/exampletags"
	awstypes "github.com/aws/aws-sdk-go-v2/service/exampletags/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfexampletagsgen "github.com/hashicorp/terraform-provider-aws/internal/service/exampletagsgen"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
)

// ListTags lists the tags of a widget or gadget, and sets them in Context.
// It is called by the transparent tagging interceptor.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier, resourceType string) error {
	var (
		tags tftags.KeyValueTags
		err  error
	)
	c := meta.(*conns.AWSClient)
	conn := c.ExampleTagsClient(ctx)

	switch resourceType {
	case "Widget":
		// Attempt the cross-service tagging API first, fall back to ListWidgetTags.
		tags, err = tfexampletagsgen.ListTags(ctx, c.ExampleTagsGenClient(ctx), identifier)
		if errs.Contains(err, "is not authorized to perform: exampletagsgen:ListTagsForResource") {
			tags, err = widgetListTags(ctx, conn, identifier)
		}

	case "Gadget":
		tags, err = gadgetListTags(ctx, conn, identifier)

	default:
		return nil
	}

	if err != nil {
		return err
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}

// UpdateTags updates the tags of a widget or gadget.
// It is called by the transparent tagging interceptor.
func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier, resourceType string, oldTags, newTags any) error {
	c := meta.(*conns.AWSClient)
	conn := c.ExampleTagsClient(ctx)

	switch resourceType {
	case "Widget":
		// Attempt the cross-service tagging API first, fall back to Tag/UntagWidget.
		err := tfexampletagsgen.UpdateTags(ctx, c.ExampleTagsGenClient(ctx), identifier, oldTags, newTags)
		if errs.Contains(err, "is not authorized to perform: exampletagsgen:TagResource") {
			return widgetUpdateTags(ctx, conn, identifier, oldTags, newTags)
		}
		return err

	case "Gadget":
		return gadgetUpdateTags(ctx, conn, identifier, oldTags, newTags)
	}

	return nil
}

func widgetListTags(ctx context.Context, conn *exampletags.Client, identifier string) (tftags.KeyValueTags, error) {
	input := exampletags.ListWidgetTagsInput{
		WidgetArn: aws.String(identifier),
	}

	output, err := conn.ListWidgetTags(ctx, &input)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTags(ctx, output.Tags), nil
}

func widgetUpdateTags(ctx context.Context, conn *exampletags.Client, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags, newTags := tftags.New(ctx, oldTagsMap), tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := exampletags.UntagWidgetInput{
			TagKeys:   removedTags.Keys(),
			WidgetArn: aws.String(identifier),
		}

		if _, err := conn.UntagWidget(ctx, &input); err != nil {
			return err
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := exampletags.TagWidgetInput{
			Tags:      svcTags(updatedTags),
			WidgetArn: aws.String(identifier),
		}

		if _, err := conn.TagWidget(ctx, &input); err != nil {
			return err
		}
	}

	return nil
}

func gadgetListTags(ctx context.Context, conn *exampletags.Client, identifier string) (tftags.KeyValueTags, error) {
	input := exampletags.ListGadgetTagsInput{
		GadgetArn: aws.String(identifier),
	}

	output, err := conn.ListGadgetTags(ctx, &input)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTags(ctx, output.Tags), nil
}

func gadgetUpdateTags(ctx context.Context, conn *exampletags.Client, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags, newTags := tftags.New(ctx, oldTagsMap), tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := exampletags.UntagGadgetInput{
			GadgetArn: aws.String(identifier),
			TagKeys:   removedTags.Keys(),
		}

		if _, err := conn.UntagGadget(ctx, &input); err != nil {
			return err
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := exampletags.TagGadgetInput{
			GadgetArn: aws.String(identifier),
			Tags:      svcTags(updatedTags),
		}

		if _, err := conn.TagGadget(ctx, &input); err != nil {
			return err
		}
	}

	return nil
}

// getTagsIn returns tags from Context, for use in a create request.
func getTagsIn(ctx context.Context) []awstypes.Tag {
	if inContext, ok := tftags.FromContext(ctx); ok {
		return svcTags(inContext.TagsIn.UnwrapOrDefault())
	}

	return nil
}

// setTagsOut sets tags in Context, from a read response.
func setTagsOut(ctx context.Context, tags []awstypes.Tag) {
	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(keyValueTags(ctx, tags))
	}
}

func keyValueTags(ctx context.Context, tags []awstypes.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.ToString(tag.Key)] = tag.Value
	}

	return tftags.New(ctx, m)
}

func svcTags(tags tftags.KeyValueTags) []awstypes.Tag {
	result := make([]awstypes.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		result = append(result, awstypes.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	return result
}
