// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// nameTag returns the value of the object's Name tag, if it has one.
func nameTag(tags []awstypes.Tag) (string, bool) {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == "Name" {
			return aws.ToString(tag.Value), true
		}
	}

	return "", false
}

// listDisplayName labels a list result "<Name tag> (<id>)" when the object
// carries a Name tag, and the id alone otherwise. EC2 objects are identified by
// an opaque id, so the tag is what a reader recognizes.
func listDisplayName(tags []awstypes.Tag, id string) string {
	if name, ok := nameTag(tags); ok {
		return fmt.Sprintf("%s (%s)", name, id)
	}

	return id
}
