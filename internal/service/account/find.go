package account

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/account"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAlternateContactByAccountIDAndContactType(ctx context.Context, conn *account.Account, accountID, contactType string) (*account.AlternateContact, error) { // nosemgrep:account-in-func-name
	input := &account.GetAlternateContactInput{
		AlternateContactType: aws.String(contactType),
	}

	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	output, err := conn.GetAlternateContactWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, account.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AlternateContact == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AlternateContact, nil
}
