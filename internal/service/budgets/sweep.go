// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package budgets

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/budgets"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_budgets_budget", sweepBudgets, "aws_budgets_budget_action")
	awsv2.Register("aws_budgets_budget_action", sweepBudgetActions)
}

func sweepBudgets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) { // nosemgrep:ci.budgets-in-func-name
	conn := client.BudgetsClient(ctx)
	accountID := client.AccountID(ctx)
	input := budgets.DescribeBudgetsInput{
		AccountId: aws.String(accountID),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := budgets.NewDescribeBudgetsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Budgets {
			// skip budgets we have configured to track our spend
			budgetName := aws.ToString(v.BudgetName)
			if !strings.HasPrefix(budgetName, "tf-acc") {
				continue
			}

			r := resourceBudget()
			d := r.Data(nil)
			d.SetId(budgetCreateResourceID(accountID, budgetName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepBudgetActions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.BudgetsClient(ctx)
	accountID := client.AccountID(ctx)
	input := budgets.DescribeBudgetActionsForAccountInput{
		AccountId: aws.String(accountID),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := budgets.NewDescribeBudgetActionsForAccountPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Actions {
			r := resourceBudgetAction()
			d := r.Data(nil)
			d.SetId(budgetActionCreateResourceID(accountID, aws.ToString(v.ActionId), aws.ToString(v.BudgetName)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
