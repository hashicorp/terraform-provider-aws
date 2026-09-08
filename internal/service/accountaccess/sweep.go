// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	// Applications must be swept after Entitlements: DeleteApplication fails
	// while child Entitlements remain.
	awsv2.Register("aws_accountaccess_application", sweepApplications, "aws_accountaccess_entitlement")
	awsv2.Register("aws_accountaccess_entitlement", sweepEntitlements)
}

func sweepApplications(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AccountAccessClient(ctx)
	var input accountaccess.ListApplicationsInput
	var sweepResources []sweep.Sweepable

	pages := accountaccess.NewListApplicationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Applications {
			sweepResources = append(sweepResources, framework.NewSweepResource(newApplicationResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.ApplicationArn)),
			))
		}
	}

	return sweepResources, nil
}

func sweepEntitlements(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AccountAccessClient(ctx)
	accountID := client.AccountID(ctx)
	var sweepResources []sweep.Sweepable
	var applicationInput accountaccess.ListApplicationsInput

	for application, err := range listApplications(ctx, conn, &applicationInput) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		applicationARN := aws.ToString(application.ApplicationArn)
		if applicationARN == "" {
			continue
		}

		entitlementInput := accountaccess.ListEntitlementsInput{
			ApplicationArn: aws.String(applicationARN),
			Filter: &awstypes.EntitlementFilter{
				PrincipalRole: &awstypes.PrincipalRoleEntitlementFilter{
					Account: aws.String(accountID),
				},
			},
		}
		for entitlement, err := range listEntitlements(ctx, conn, &entitlementInput) {
			if err != nil {
				return nil, smarterr.NewError(err)
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newEntitlementResource, client,
				framework.NewAttribute("application_arn", applicationARN),
				framework.NewAttribute("entitlement_id", aws.ToString(entitlement.EntitlementId)),
			))
		}
	}

	return sweepResources, nil
}
