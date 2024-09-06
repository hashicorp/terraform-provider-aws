// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const eventualConsistencyTimeout = 5 * time.Minute

// createTags creates ec2 service tags for new resources.
func createTags(ctx context.Context, conn *ec2.Client, identifier string, tags []awstypes.Tag, optFns ...func(*ec2.Options)) error {
	if len(tags) == 0 {
		return nil
	}

	newTagsMap := keyValueTags(ctx, tags)

	_, err := tfresource.RetryWhenAWSErrCodeContains(ctx, eventualConsistencyTimeout, func() (interface{}, error) {
		return nil, updateTags(ctx, conn, identifier, nil, newTagsMap, optFns...)
	}, ".NotFound")

	if err != nil {
		return fmt.Errorf("tagging resource (%s): %w", identifier, err)
	}

	return nil
}

// getTagSpecificationsIn returns AWS SDK for Go v2 EC2 service tags from Context.
// nil is returned if there are no input tags.
func getTagSpecificationsIn(ctx context.Context, resourceType awstypes.ResourceType) []awstypes.TagSpecification {
	tags := getTagsIn(ctx)

	if len(tags) == 0 {
		return nil
	}

	return []awstypes.TagSpecification{
		{
			ResourceType: resourceType,
			Tags:         tags,
		},
	}
}

// tagSpecificationsFromMap returns the tag specifications for the given tag key/value map and resource type.
func tagSpecificationsFromMap(ctx context.Context, m map[string]interface{}, t awstypes.ResourceType) []awstypes.TagSpecification {
	if len(m) == 0 {
		return nil
	}

	return []awstypes.TagSpecification{
		{
			ResourceType: t,
			Tags:         Tags(tftags.New(ctx, m).IgnoreAWS()),
		},
	}
}

// tagSpecificationsFromKeyValue returns the tag specifications for the given tag key/value tags and resource type.
func tagSpecificationsFromKeyValue(tags tftags.KeyValueTags, resourceType string) []awstypes.TagSpecification {
	if len(tags) == 0 {
		return nil
	}

	return []awstypes.TagSpecification{
		{
			ResourceType: awstypes.ResourceType(resourceType),
			Tags:         Tags(tags.IgnoreAWS()),
		},
	}
}

// tagsFromTagDescriptions returns the tags from the given tag descriptions.
// No attempt is made to remove duplicates.
func tagsFromTagDescriptions(tds []awstypes.TagDescription) []awstypes.Tag {
	if len(tds) == 0 {
		return nil
	}

	tags := []awstypes.Tag{}
	for _, td := range tds {
		tags = append(tags, awstypes.Tag{
			Key:   td.Key,
			Value: td.Value,
		})
	}

	return tags
}

func tagsSchemaConflictsWith(conflictsWith []string) *schema.Schema {
	v := tftags.TagsSchema()
	v.ConflictsWith = conflictsWith

	return v
}
