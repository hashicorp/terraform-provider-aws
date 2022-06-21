package account

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/account"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	statusFound      = "FOUND"
	statusUpdated    = "UPDATED"
	statusNotUpdated = "NOT_UPDATED"
)

func statusAlternateContact(ctx context.Context, conn *account.Account, accountID, contactType string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAlternateContactByAccountIDAndContactType(ctx, conn, accountID, contactType)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, statusFound, nil
	}
}

func statusAlternateContactUpdate(ctx context.Context, conn *account.Account, accountID, contactType, email, name, phone, title string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAlternateContactByAccountIDAndContactType(ctx, conn, accountID, contactType)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if email == aws.StringValue(output.EmailAddress) &&
			name == aws.StringValue(output.Name) &&
			phone == aws.StringValue(output.PhoneNumber) &&
			title == aws.StringValue(output.Title) {
			return output, statusUpdated, nil
		}

		return output, statusNotUpdated, nil
	}
}
