// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package workmail

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workmail"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_workmail_organization", sweepOrganizations)
	awsv2.Register("aws_workmail_user", sweepUsers)
}

func sweepOrganizations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := workmail.ListOrganizationsInput{}
	conn := client.WorkMailClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := workmail.NewListOrganizationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.OrganizationSummaries {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newOrganizationResource, client,
				sweepfw.NewAttribute("organization_id", aws.ToString(v.OrganizationId))),
			)
		}
	}

	return sweepResources, nil
}

func sweepUsers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.WorkMailClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := workmail.NewListOrganizationsPaginator(conn, &workmail.ListOrganizationsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, organization := range page.OrganizationSummaries {
			organizationID := aws.ToString(organization.OrganizationId)

			input := workmail.ListUsersInput{
				OrganizationId: aws.String(organizationID),
			}

			userPages := workmail.NewListUsersPaginator(conn, &input)
			for userPages.HasMorePages() {
				page, err := userPages.NextPage(ctx)
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, user := range page.Users {
					sweepResources = append(sweepResources, sweepfw.NewSweepResource(newUserResource, client,
						sweepfw.NewAttribute("organization_id", organizationID),
						sweepfw.NewAttribute("user_id", aws.ToString(user.Id))),
					)
				}
			}
		}
	}

	return sweepResources, nil
}
