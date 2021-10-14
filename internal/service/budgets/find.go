package budgets

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func FindActionByAccountIDActionIDAndBudgetName(conn *budgets.Budgets, accountID, actionID, budgetName string) (*budgets.Action, error) {
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

func FindBudgetByAccountIDAndBudgetName(conn *budgets.Budgets, accountID, budgetName string) (*budgets.Budget, error) {
	input := &budgets.DescribeBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	}

	output, err := conn.DescribeBudget(input)

	if tfawserr.ErrCodeEquals(err, budgets.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Budget == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Budget, nil
}

func FindNotificationsByAccountIDAndBudgetName(conn *budgets.Budgets, accountID, budgetName string) ([]*budgets.Notification, error) {
	input := &budgets.DescribeNotificationsForBudgetInput{
		AccountId:  aws.String(accountID),
		BudgetName: aws.String(budgetName),
	}
	var output []*budgets.Notification

	err := conn.DescribeNotificationsForBudgetPages(input, func(page *budgets.DescribeNotificationsForBudgetOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, notification := range page.Notifications {
			if notification == nil {
				continue
			}

			output = append(output, notification)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, budgets.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSubscribersByAccountIDBudgetNameAndNotification(conn *budgets.Budgets, accountID, budgetName string, notification *budgets.Notification) ([]*budgets.Subscriber, error) {
	input := &budgets.DescribeSubscribersForNotificationInput{
		AccountId:    aws.String(accountID),
		BudgetName:   aws.String(budgetName),
		Notification: notification,
	}
	var output []*budgets.Subscriber

	err := conn.DescribeSubscribersForNotificationPages(input, func(page *budgets.DescribeSubscribersForNotificationOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, subscriber := range page.Subscribers {
			if subscriber == nil {
				continue
			}

			output = append(output, subscriber)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, budgets.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}
