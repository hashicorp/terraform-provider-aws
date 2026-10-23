// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package exampletags

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/exampletags"
	awstypes "github.com/aws/aws-sdk-go-v2/service/exampletags/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfexampletagsgen "github.com/hashicorp/terraform-provider-aws/internal/service/exampletagsgen"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
)

// ListTags lists the tags of a foo, bar, or baz, and sets them in Context.
// It is called by the transparent tagging interceptor.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier, resourceType string) error {
	var (
		tags tftags.KeyValueTags
		err  error
	)
	c := meta.(*conns.AWSClient)
	conn := c.ExampleTagsClient(ctx)

	switch resourceType {
	case "Foo":
		// Attempt the cross-service tagging API first, fall back to ListFooTags.
		tags, err = tfexampletagsgen.ListTags(ctx, c.ExampleTagsGenClient(ctx), identifier)
		if errs.Contains(err, "is not authorized to perform: exampletagsgen:ListTagsForResource") {
			tags, err = fooListTags(ctx, conn, identifier)
		}

	case "Bar", "Baz":
		tags, err = barListTags(ctx, conn, identifier)

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

// UpdateTags updates the tags of a foo, bar, or baz.
// It is called by the transparent tagging interceptor.
func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier, resourceType string, oldTags, newTags any) error {
	return updateTags(ctx, meta.(*conns.AWSClient), identifier, resourceType, oldTags, newTags)
}

func updateTags(ctx context.Context, c *conns.AWSClient, identifier, resourceType string, oldTags, newTags any) error {
	conn := c.ExampleTagsClient(ctx)

	switch resourceType {
	case "Foo":
		// Attempt the cross-service tagging API first, fall back to Tag/UntagFoo.
		err := tfexampletagsgen.UpdateTags(ctx, c.ExampleTagsGenClient(ctx), identifier, oldTags, newTags)
		if errs.Contains(err, "is not authorized to perform: exampletagsgen:TagResource") {
			return fooUpdateTags(ctx, conn, identifier, oldTags, newTags)
		}
		return err

	case "Bar", "Baz":
		return barUpdateTags(ctx, conn, identifier, oldTags, newTags)
	}

	return fmt.Errorf("unsupported resource type: %s", resourceType)
}

func fooListTags(ctx context.Context, conn *exampletags.Client, identifier string) (tftags.KeyValueTags, error) {
	input := exampletags.ListFooTagsInput{
		FooArn: aws.String(identifier),
	}

	output, err := conn.ListFooTags(ctx, &input)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTags(ctx, output.Tags), nil
}

func fooUpdateTags(ctx context.Context, conn *exampletags.Client, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags, newTags := tftags.New(ctx, oldTagsMap), tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := exampletags.UntagFooInput{
			TagKeys: removedTags.Keys(),
			FooArn:  aws.String(identifier),
		}

		if _, err := conn.UntagFoo(ctx, &input); err != nil {
			return err
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := exampletags.TagFooInput{
			Tags:   svcTags(updatedTags),
			FooArn: aws.String(identifier),
		}

		if _, err := conn.TagFoo(ctx, &input); err != nil {
			return err
		}
	}

	return nil
}

func barListTags(ctx context.Context, conn *exampletags.Client, identifier string) (tftags.KeyValueTags, error) {
	input := exampletags.ListBarTagsInput{
		BarArn: aws.String(identifier),
	}

	output, err := conn.ListBarTags(ctx, &input)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTags(ctx, output.Tags), nil
}

// listTags lists tags through the service's generic tagging API, for resource types without their
// own.
func listTags(ctx context.Context, conn *exampletags.Client, identifier string) (tftags.KeyValueTags, error) {
	input := exampletags.ListTagsForResourceInput{
		ResourceArn: aws.String(identifier),
	}

	output, err := conn.ListTagsForResource(ctx, &input)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTags(ctx, output.Tags), nil
}

func barUpdateTags(ctx context.Context, conn *exampletags.Client, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags, newTags := tftags.New(ctx, oldTagsMap), tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := exampletags.UntagBarInput{
			BarArn:  aws.String(identifier),
			TagKeys: removedTags.Keys(),
		}

		if _, err := conn.UntagBar(ctx, &input); err != nil {
			return err
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := exampletags.TagBarInput{
			BarArn: aws.String(identifier),
			Tags:   svcTags(updatedTags),
		}

		if _, err := conn.TagBar(ctx, &input); err != nil {
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
