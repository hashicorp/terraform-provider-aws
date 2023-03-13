package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// tagSpecificationsFromKeyValueTags returns the tag specifications for the given KeyValueTags object and resource type.
func tagSpecificationsFromKeyValueTags(tags tftags.KeyValueTags, t string) []*ec2.TagSpecification {
	if len(tags) == 0 {
		return nil
	}

	return []*ec2.TagSpecification{
		{
			ResourceType: aws.String(t),
			Tags:         Tags(tags.IgnoreAWS()),
		},
	}
}

// tagSpecificationsFromMap returns the tag specifications for the given tag key/value map and resource type.
func tagSpecificationsFromMap(ctx context.Context, m map[string]interface{}, t string) []*ec2.TagSpecification {
	if len(m) == 0 {
		return nil
	}

	return []*ec2.TagSpecification{
		{
			ResourceType: aws.String(t),
			Tags:         Tags(tftags.New(ctx, m).IgnoreAWS()),
		},
	}
}

// tagsFromTagDescriptions returns the tags from the given tag descriptions.
// No attempt is made to remove duplicates.
func tagsFromTagDescriptions(tds []*ec2.TagDescription) []*ec2.Tag {
	if len(tds) == 0 {
		return nil
	}

	tags := []*ec2.Tag{}
	for _, td := range tds {
		tags = append(tags, &ec2.Tag{
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
