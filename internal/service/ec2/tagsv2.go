// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// createTags creates ec2 service tags for new resources.
func createTagsV2(ctx context.Context, conn *ec2.Client, identifier string, tags []types.Tag, optFns ...func(*ec2.Options)) error {
	if len(tags) == 0 {
		return nil
	}

	newTagsMap := keyValueTagsV2(ctx, tags)

	_, err := tfresource.RetryWhenAWSErrCodeContains(ctx, eventualConsistencyTimeout, func() (interface{}, error) {
		return nil, updateTagsV2(ctx, conn, identifier, nil, newTagsMap, optFns...)
	}, ".NotFound")

	if err != nil {
		return fmt.Errorf("tagging resource (%s): %w", identifier, err)
	}

	return nil
}

// getTagSpecificationsInV2 returns AWS SDK for Go v2 EC2 service tags from Context.
// nil is returned if there are no input tags.
func getTagSpecificationsInV2(ctx context.Context, resourceType types.ResourceType) []types.TagSpecification {
	tags := getTagsInV2(ctx)

	if len(tags) == 0 {
		return nil
	}

	return []types.TagSpecification{
		{
			ResourceType: resourceType,
			Tags:         tags,
		},
	}
}
