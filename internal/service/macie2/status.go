package macie2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// statusMemberRelationship fetches the Member and its relationship status
func statusMemberRelationship(conn *macie2.Macie2, adminAccountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		adminAccount, err := findMemberNotAssociated(conn, adminAccountID)

		if err != nil {
			return nil, "Unknown", err
		}

		if adminAccount == nil {
			return adminAccount, "NotFound", nil
		}

		return adminAccount, aws.StringValue(adminAccount.RelationshipStatus), nil
	}
}
