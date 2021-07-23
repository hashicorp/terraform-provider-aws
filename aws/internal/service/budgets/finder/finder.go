package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func ActionByAccountIDActionIDAndBudgetName(conn *budgets.Budgets, accountID, actionID, budgetName string) (*budgets.Action, error) {
	input := &budgets.DescribeBudgetActionInput{
		AccountId:  aws.String(accountID),
		ActionId:   aws.String(actionID),
		BudgetName: aws.String(budgetName),
	}

	output, err := conn.DescribeBudgetAction(input)

	if tfawserr.ErrCodeEquals(err, budgets.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Action == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Action, nil
}
