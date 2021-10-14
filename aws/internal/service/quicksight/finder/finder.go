package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func FindGroupMembership(conn *quicksight.QuickSight, listInput *quicksight.ListGroupMembershipsInput, userName string) (bool, error) {

	found := false

	for {
		resp, err := conn.ListGroupMemberships(listInput)
		if err != nil {
			return false, err
		}

		for _, member := range resp.GroupMemberList {
			if aws.StringValue(member.MemberName) == userName {
				found = true
				break
			}
		}

		if found || resp.NextToken == nil {
			break
		}

		listInput.NextToken = resp.NextToken
	}

	return found, nil
}
