package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Flattens security group identifiers into a []string, where the elements returned are the GroupIDs
func FlattenGroupIdentifiers(dtos []*ec2.GroupIdentifier) []string {
	ids := make([]string, 0, len(dtos))
	for _, v := range dtos {
		group_id := aws.StringValue(v.GroupId)
		ids = append(ids, group_id)
	}
	return ids
}
