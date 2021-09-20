package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/securityhub/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// AdminStatus NotFound
	AdminStatusNotFound = "NotFound"

	// AdminStatus Unknown
	AdminStatusUnknown = "Unknown"

	StandardsStatusNotFound = "NotFound"
)

// AdminAccountAdminStatus fetches the AdminAccount and its AdminStatus
func AdminAccountAdminStatus(conn *securityhub.SecurityHub, adminAccountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		adminAccount, err := finder.AdminAccount(conn, adminAccountID)

		if err != nil {
			return nil, AdminStatusUnknown, err
		}

		if adminAccount == nil {
			return adminAccount, AdminStatusNotFound, nil
		}

		return adminAccount, aws.StringValue(adminAccount.Status), nil
	}
}

func StandardsSubscriptionStatus(conn *securityhub.SecurityHub, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.StandardsSubscriptionByARN(conn, arn)

		if tfresource.NotFound(err) {
			// Return a fake result and status to deal with the INCOMPLETE subscription status
			// being a target for both Create and Delete.
			return "", StandardsStatusNotFound, nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.StandardsStatus), nil
	}
}
