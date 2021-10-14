package budgets

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func statusAction(conn *budgets.Budgets, accountID, actionID, budgetName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindActionByAccountIDActionIDAndBudgetName(conn, accountID, actionID, budgetName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
