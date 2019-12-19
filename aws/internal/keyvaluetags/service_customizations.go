// This file contains customizations for each AWS Go SDK service.

package keyvaluetags

import (
	"github.com/aws/aws-sdk-go/aws"
)

// KinesisTagInput transforms a standard KeyValueTags to the form
// required for input to the Kinesis AddTagsToStreamInput() method.
func KinesisTagInput(tags KeyValueTags) map[string]*string {
	return aws.StringMap(tags.Map())
}
