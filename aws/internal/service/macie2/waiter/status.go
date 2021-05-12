package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// MemberRelationshipStatus fetches the Member and its relationship status
func MemberRelationshipStatus(conn *macie2.Macie2, adminAccountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		adminAccount, err := getMemberNotAssociated(conn, adminAccountID)

		if err != nil {
			return nil, "Unknown", err
		}

		if adminAccount == nil {
			return adminAccount, "NotFound", nil
		}

		return adminAccount, aws.StringValue(adminAccount.RelationshipStatus), nil
	}
}

// TODO: Migrate to shared internal package for aws package and this package
func getMemberNotAssociated(conn *macie2.Macie2, adminAccountID string) (*macie2.Member, error) {
	input := &macie2.ListMembersInput{
		OnlyAssociated: aws.String("false"),
	}
	var result *macie2.Member

	err := conn.ListMembersPages(input, func(page *macie2.ListMembersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, member := range page.Members {
			if member == nil {
				continue
			}

			if aws.StringValue(member.AdministratorAccountId) == adminAccountID {
				result = member
				return false
			}
		}

		return !lastPage
	})

	return result, err
}
