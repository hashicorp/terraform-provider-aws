// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
)

// listTags lists ssm service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func listTags(ctx context.Context, conn *ssm.Client, identifier, resourceType string, optFns ...func(*ssm.Options)) (tftags.KeyValueTags, error) {
	switch resourceType {
	case "Activation":
		return activationTags(ctx, conn, identifier)

	default:
		input := ssm.ListTagsForResourceInput{
			ResourceId:   aws.String(identifier),
			ResourceType: awstypes.ResourceTypeForTagging(resourceType),
		}

		output, err := conn.ListTagsForResource(ctx, &input, optFns...)
		if err != nil {
			return tftags.New(ctx, nil), err
		}

		return keyValueTags(ctx, output.TagList), nil
	}
}

// ListTags lists ssm service tags and set them in Context.
// It is called from outside this package.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier, resourceType string) error {
	tags, err := listTags(ctx, meta.(*conns.AWSClient).SSMClient(ctx), identifier, resourceType)
	if err != nil {
		return err
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}

func activationTags(ctx context.Context, client *ssm.Client, id string) (tftags.KeyValueTags, error) {
	out, err := findActivationByID(ctx, client, id)
	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return keyValueTags(ctx, out.Tags), nil
}
