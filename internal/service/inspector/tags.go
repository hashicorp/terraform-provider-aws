//go:build !generate
// +build !generate

package inspector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// Custom Inspector tag service update functions using the same format as generated code.

// updateTags updates WorkSpaces resource tags.
// The identifier is the resource ARN.
func updateTags(conn *inspector.Inspector, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	if len(newTags) > 0 {
		input := &inspector.SetTagsForResourceInput{
			ResourceArn: aws.String(identifier),
			Tags:        newTags.IgnoreAws().InspectorTags(),
		}

		_, err := conn.SetTagsForResource(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	} else if len(oldTags) > 0 {
		input := &inspector.SetTagsForResourceInput{
			ResourceArn: aws.String(identifier),
		}

		_, err := conn.SetTagsForResource(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
