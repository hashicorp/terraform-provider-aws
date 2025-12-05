// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
)

// ListTags lists redshift service tags and sets them in Context.
// It is called from outside this package.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier string) error {
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	input := redshift.DescribeTagsInput{
		ResourceName: aws.String(identifier),
	}
	output, err := conn.DescribeTags(ctx, &input)
	if err != nil {
		return smarterr.NewError(err)
	}

	awsTags := make([]awstypes.Tag, 0, len(output.TaggedResources))
	for _, tag := range output.TaggedResources {
		if tag.Tag != nil {
			awsTags = append(awsTags, *tag.Tag)
		}
	}

	tags := keyValueTags(ctx, awsTags)

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}
