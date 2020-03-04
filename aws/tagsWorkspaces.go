package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

// WorkspacesUpdateTags custom function, which updates workspaces service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func WorkspacesUpdateTags(conn *workspaces.WorkSpaces, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := keyvaluetags.New(oldTagsMap)
	newTags := keyvaluetags.New(newTagsMap)

	if len(newTags) == 0 {
		input := &workspaces.DeleteTagsInput{
			ResourceId: aws.String(identifier),
			TagKeys:    aws.StringSlice(oldTags.Keys()),
		}

		if _, err := conn.DeleteTags(input); err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	} else {
		input := &workspaces.CreateTagsInput{
			ResourceId: aws.String(identifier),
			Tags:       newTags.IgnoreAws().WorkspacesTags(),
		}

		if _, err := conn.CreateTags(input); err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
