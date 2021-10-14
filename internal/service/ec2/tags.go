package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// getInstanceTagValue returns instance tag value by name
func getInstanceTagValue(conn *ec2.EC2, instanceId string, tagKey string) (*string, error) {
	tagsResp, err := conn.DescribeTags(&ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("resource-id"),
				Values: []*string{aws.String(instanceId)},
			},
			{
				Name:   aws.String("key"),
				Values: []*string{aws.String(tagKey)},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(tagsResp.Tags) != 1 {
		return nil, nil
	}

	return tagsResp.Tags[0].Value, nil
}

// ec2TagSpecificationsFromKeyValueTags returns the tag specifications for the given KeyValueTags object and resource type.
func ec2TagSpecificationsFromKeyValueTags(tags tftags.KeyValueTags, t string) []*ec2.TagSpecification {
	if len(tags) == 0 {
		return nil
	}

	return []*ec2.TagSpecification{
		{
			ResourceType: aws.String(t),
			Tags:         tags.IgnoreAws().Ec2Tags(),
		},
	}
}

// ec2TagSpecificationsFromMap returns the tag specifications for the given tag key/value map and resource type.
func ec2TagSpecificationsFromMap(m map[string]interface{}, t string) []*ec2.TagSpecification {
	if len(m) == 0 {
		return nil
	}

	return []*ec2.TagSpecification{
		{
			ResourceType: aws.String(t),
			Tags:         tftags.New(m).IgnoreAws().Ec2Tags(),
		},
	}
}

// ec2TagsFromTagDescriptions returns the tags from the given tag descriptions.
// No attempt is made to remove duplicates.
func ec2TagsFromTagDescriptions(tds []*ec2.TagDescription) []*ec2.Tag {
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
	return &schema.Schema{
		ConflictsWith: conflictsWith,
		Type:          schema.TypeMap,
		Optional:      true,
		Elem:          &schema.Schema{Type: schema.TypeString},
	}
}
