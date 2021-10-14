package finder

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func AdminAccount(conn *securityhub.SecurityHub, adminAccountID string) (*securityhub.AdminAccount, error) {
	input := &securityhub.ListOrganizationAdminAccountsInput{}
	var result *securityhub.AdminAccount

	err := conn.ListOrganizationAdminAccountsPages(input, func(page *securityhub.ListOrganizationAdminAccountsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, adminAccount := range page.AdminAccounts {
			if adminAccount == nil {
				continue
			}

			if aws.StringValue(adminAccount.AccountId) == adminAccountID {
				result = adminAccount
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func Insight(ctx context.Context, conn *securityhub.SecurityHub, arn string) (*securityhub.Insight, error) {
	input := &securityhub.GetInsightsInput{
		InsightArns: aws.StringSlice([]string{arn}),
		MaxResults:  aws.Int64(1),
	}

	output, err := conn.GetInsightsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Insights) == 0 {
		return nil, nil
	}

	return output.Insights[0], nil
}

func StandardsControlByStandardsSubscriptionARNAndStandardsControlARN(ctx context.Context, conn *securityhub.SecurityHub, standardsSubscriptionARN, standardsControlARN string) (*securityhub.StandardsControl, error) {
	input := &securityhub.DescribeStandardsControlsInput{
		StandardsSubscriptionArn: aws.String(standardsSubscriptionARN),
	}
	var output *securityhub.StandardsControl

	err := conn.DescribeStandardsControlsPagesWithContext(ctx, input, func(page *securityhub.DescribeStandardsControlsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, control := range page.Controls {
			if aws.StringValue(control.StandardsControlArn) == standardsControlARN {
				output = control

				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}

func StandardsSubscriptionByARN(conn *securityhub.SecurityHub, arn string) (*securityhub.StandardsSubscription, error) {
	input := &securityhub.GetEnabledStandardsInput{
		StandardsSubscriptionArns: aws.StringSlice([]string{arn}),
	}

	output, err := conn.GetEnabledStandards(input)

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if output == nil || len(output.StandardsSubscriptions) == 0 || output.StandardsSubscriptions[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	// TODO Check for multiple results.
	// TODO https://github.com/hashicorp/terraform-provider-aws/pull/17613.

	subscription := output.StandardsSubscriptions[0]

	if status := aws.StringValue(subscription.StandardsStatus); status == securityhub.StandardsStatusFailed {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return subscription, nil
}
