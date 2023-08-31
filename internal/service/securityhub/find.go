// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAdminAccount(ctx context.Context, conn *securityhub.SecurityHub, adminAccountID string) (*securityhub.AdminAccount, error) {
	input := &securityhub.ListOrganizationAdminAccountsInput{}
	var result *securityhub.AdminAccount

	err := conn.ListOrganizationAdminAccountsPagesWithContext(ctx, input, func(page *securityhub.ListOrganizationAdminAccountsOutput, lastPage bool) bool {
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

func FindInsight(ctx context.Context, conn *securityhub.SecurityHub, arn string) (*securityhub.Insight, error) {
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

func FindStandardsControlByStandardsSubscriptionARNAndStandardsControlARN(ctx context.Context, conn *securityhub.SecurityHub, standardsSubscriptionARN, standardsControlARN string) (*securityhub.StandardsControl, error) {
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
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
