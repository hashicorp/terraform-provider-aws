// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

// Tagging for Lambda Web functions is not generated: the Lambda Web API has no
// TagResource, UntagResource or ListTagsForResource operation, and
// GetWebFunction does not return tags. Tags set through CreateWebFunction are
// readable and writable through the classic Lambda tagging operations against
// the web function ARN (arn:aws:lambda:<region>:<account>:web-function/<name>),
// which share the "lambda" ARN namespace, so this file implements the tagging
// interfaces with the Lambda client instead of the Lambda Web one.

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// listTags lists Lambda Web function tags through the Lambda tagging API.
func listTags(ctx context.Context, conn *lambda.Client, identifier string, optFns ...func(*lambda.Options)) (tftags.KeyValueTags, error) {
	input := lambda.ListTagsInput{
		Resource: aws.String(identifier),
	}

	output, err := conn.ListTags(ctx, &input, optFns...)

	if err != nil {
		return tftags.New(ctx, nil), smarterr.NewError(err)
	}

	return keyValueTags(ctx, output.Tags), nil
}

// ListTags lists Lambda Web function tags and sets them in Context.
// It is called from outside this package.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier string) error {
	tags, err := listTags(ctx, meta.(*conns.AWSClient).LambdaClient(ctx), identifier)

	if err != nil {
		return smarterr.NewError(err)
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}

// svcTags returns Lambda Web service tags.
func svcTags(tags tftags.KeyValueTags) map[string]string {
	return tags.Map()
}

// keyValueTags creates tftags.KeyValueTags from Lambda Web service tags.
func keyValueTags(ctx context.Context, tags map[string]string) tftags.KeyValueTags {
	return tftags.New(ctx, tags)
}

// getTagsIn returns Lambda Web service tags from Context.
// nil is returned if there are no input tags.
func getTagsIn(ctx context.Context) map[string]string {
	if inContext, ok := tftags.FromContext(ctx); ok {
		if tags := svcTags(inContext.TagsIn.UnwrapOrDefault()); len(tags) > 0 {
			return tags
		}
	}

	return nil
}

// updateTags updates Lambda Web function tags through the Lambda tagging API.
func updateTags(ctx context.Context, conn *lambda.Client, identifier string, oldTagsMap, newTagsMap any, optFns ...func(*lambda.Options)) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, identifier)

	removedTags := oldTags.Removed(newTags)
	removedTags = removedTags.IgnoreSystem(names.Lambda)
	if len(removedTags) > 0 {
		input := lambda.UntagResourceInput{
			Resource: aws.String(identifier),
			TagKeys:  removedTags.Keys(),
		}

		_, err := conn.UntagResource(ctx, &input, optFns...)

		if err != nil {
			return smarterr.NewError(err)
		}
	}

	updatedTags := oldTags.Updated(newTags)
	updatedTags = updatedTags.IgnoreSystem(names.Lambda)
	if len(updatedTags) > 0 {
		input := lambda.TagResourceInput{
			Resource: aws.String(identifier),
			Tags:     svcTags(updatedTags),
		}

		_, err := conn.TagResource(ctx, &input, optFns...)

		if err != nil {
			return smarterr.NewError(err)
		}
	}

	return nil
}

// UpdateTags updates Lambda Web function tags.
// It is called from outside this package.
func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier string, oldTags, newTags any) error {
	return updateTags(ctx, meta.(*conns.AWSClient).LambdaClient(ctx), identifier, oldTags, newTags)
}
