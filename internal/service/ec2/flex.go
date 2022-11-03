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

// Like ec2.GroupIdentifier but with additional rule description.
type GroupIdentifier struct {
	// The ID of the security group.
	GroupId *string

	// The name of the security group.
	GroupName *string

	Description *string
}

// Flattens an array of UserSecurityGroups into a []*GroupIdentifier
func FlattenSecurityGroups(list []*ec2.UserIdGroupPair, ownerId *string) []*GroupIdentifier {
	result := make([]*GroupIdentifier, 0, len(list))
	for _, g := range list {
		var userId *string
		if aws.StringValue(g.UserId) != "" && (ownerId == nil || aws.StringValue(ownerId) != aws.StringValue(g.UserId)) {
			userId = g.UserId
		}
		// userid nil here for same vpc groups

		vpc := aws.StringValue(g.GroupName) == ""
		var id *string
		if vpc {
			id = g.GroupId
		} else {
			id = g.GroupName
		}

		// id is groupid for vpcs
		// id is groupname for non vpc (classic)

		if userId != nil {
			id = aws.String(*userId + "/" + *id)
		}

		if vpc {
			result = append(result, &GroupIdentifier{
				GroupId:     id,
				Description: g.Description,
			})
		} else {
			result = append(result, &GroupIdentifier{
				GroupId:     g.GroupId,
				GroupName:   id,
				Description: g.Description,
			})
		}
	}
	return result
}
