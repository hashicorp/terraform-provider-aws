// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Custom RAM tag functions using the same format as generated code.

// resourceUpdateTags updates RAM resource tags.
func resourceUpdateTags(ctx context.Context, conn *ram.Client, identifier string, oldTagsMap, newTagsMap any, optFns ...func(*ram.Options)) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, identifier)

	removedTags := oldTags.Removed(newTags)
	removedTags = removedTags.IgnoreSystem(names.RAM)
	if len(removedTags) > 0 {
		input := ram.UntagResourceInput{
			ResourceArn: aws.String(identifier),
			TagKeys:     removedTags.Keys(),
		}

		_, err := conn.UntagResource(ctx, &input, optFns...)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	updatedTags := oldTags.Updated(newTags)
	updatedTags = updatedTags.IgnoreSystem(names.RAM)
	if len(updatedTags) > 0 {
		input := ram.TagResourceInput{
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

// resourceShareUpdateTags updates RAM resource share tags.
func resourceShareUpdateTags(ctx context.Context, conn *ram.Client, identifier string, oldTagsMap, newTagsMap any, optFns ...func(*ram.Options)) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	ctx = tflog.SetField(ctx, logging.KeyResourceId, identifier)

	removedTags := oldTags.Removed(newTags)
	removedTags = removedTags.IgnoreSystem(names.RAM)
	if len(removedTags) > 0 {
		input := ram.UntagResourceInput{
			ResourceShareArn: aws.String(identifier),
			TagKeys:          removedTags.Keys(),
		}

		_, err := conn.UntagResource(ctx, &input, optFns...)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	updatedTags := oldTags.Updated(newTags)
	updatedTags = updatedTags.IgnoreSystem(names.RAM)
	if len(updatedTags) > 0 {
		input := ram.TagResourceInput{
			ResourceShareArn: aws.String(identifier),
			Tags:             svcTags(updatedTags),
		}

		_, err := conn.TagResource(ctx, &input, optFns...)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// UpdateTags updates RAM service tags.
// It is called from outside this package.
func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier, resourceType string, oldTags, newTags any) error {
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	switch resourceType {
	case "ResourceShare":
		return resourceShareUpdateTags(ctx, conn, identifier, oldTags, newTags) // nosemgrep:ci.semgrep.pluginsdk.append-Update-to-diags

	default:
		return resourceUpdateTags(ctx, conn, identifier, oldTags, newTags)
	}
}
