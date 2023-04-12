package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const eventualConsistencyTimeout = 5 * time.Minute

// CreateTags creates ec2 service tags for new resources.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func CreateTags(ctx context.Context, conn ec2iface.EC2API, identifier string, tagsMap interface{}) error {
	tags := tftags.New(ctx, tagsMap)
	input := &ec2.CreateTagsInput{
		Resources: aws.StringSlice([]string{identifier}),
		Tags:      Tags(tags.IgnoreAWS()),
	}

	_, err := tfresource.RetryWhenNotFound(ctx, eventualConsistencyTimeout, func() (interface{}, error) {
		output, err := conn.CreateTagsWithContext(ctx, input)

		if tfawserr.ErrCodeContains(err, ".NotFound") {
			err = &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		return output, err
	})

	if err != nil {
		return fmt.Errorf("tagging resource (%s): %w", identifier, err)
	}

	return nil
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

// getTagSpecificationsIn returns EC2 service tags from Context.
// nil is returned if there are no input tags.
func getTagSpecificationsIn(ctx context.Context, resourceType string) []*ec2.TagSpecification {
	tags := GetTagsIn(ctx)

	if len(tags) == 0 {
		return nil
	}

	return []*ec2.TagSpecification{
		{
			ResourceType: aws.String(resourceType),
			Tags:         tags,
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
