// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func RegisterSweepers() {
	awsv2.Register("aws_organizations_account", sweepAccounts)

	awsv2.Register("aws_organizations_organizational_unit", sweepOrganizationalUnits,
		"aws_organizations_account")
}

func sweepAccounts(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.OrganizationsClient(ctx)

	var sweepResources []sweep.Sweepable

	orgInput := organizations.DescribeOrganizationInput{}
	orgOutput, err := conn.DescribeOrganization(ctx, &orgInput)
	if errs.IsA[*awstypes.AWSOrganizationsNotInUseException](err) {
		tflog.Info(ctx, "Skipping sweeper", map[string]any{
			"skip_reason": "Not part of an AWS Organization",
		})
		return nil, nil
	}

	r := resourceAccount()

	input := organizations.ListAccountsInput{}
	pages := organizations.NewListAccountsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, account := range page.Accounts {
			if aws.ToString(account.Id) == aws.ToString(orgOutput.Organization.MasterAccountId) {
				tflog.Info(ctx, "Skipping resource", map[string]any{
					"attr.id":     account.Id,
					"skip_reason": "Organization management account",
				})
				continue
			}

			if account.Status != awstypes.AccountStatusActive {
				tflog.Info(ctx, "Skipping resource", map[string]any{
					"attr.id":     account.Id,
					"skip_reason": fmt.Sprintf("Account status is %q", account.Status),
				})
				continue
			}

			// TODO: Figure out how to not do the retry

			d := r.Data(nil)
			d.SetId(aws.ToString(account.Id))
			d.Set("close_on_deletion", true)

			sweepResources = append(sweepResources, newAccountSweeper(r, d, client))
		}
	}

	return sweepResources, nil
}

type accountSweeper struct {
	d         *schema.ResourceData
	sweepable sweep.Sweepable
}

func newAccountSweeper(resource *schema.Resource, d *schema.ResourceData, client *conns.AWSClient) *accountSweeper {
	return &accountSweeper{
		d:         d,
		sweepable: sdk.NewSweepResource(resource, d, client),
	}
}

func (as accountSweeper) Delete(ctx context.Context, optFns ...tfresource.OptionsFunc) error {
	if err := as.sweepable.Delete(ctx, optFns...); err != nil {
		if strings.Contains(err.Error(), "exceeded close account quota") {
			tflog.Info(ctx, "Ignoring error", map[string]any{
				"error": err.Error(),
			})
			return nil
		}
		return err
	}
	return nil
}

func sweepOrganizationalUnits(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.OrganizationsClient(ctx)

	orgInput := organizations.DescribeOrganizationInput{}
	orgOutput, err := conn.DescribeOrganization(ctx, &orgInput)
	if errs.IsA[*awstypes.AWSOrganizationsNotInUseException](err) {
		tflog.Info(ctx, "Skipping sweeper", map[string]any{
			"skip_reason": "Not part of an AWS Organization",
		})
		return nil, nil
	}
	if aws.ToString(orgOutput.Organization.MasterAccountId) != client.AccountID(ctx) {
		tflog.Info(ctx, "Skipping sweeper", map[string]any{
			"skip_reason": "Not the management account of an AWS Organization",
		})
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	r := resourceOrganizationalUnit()
	var sweepResources []sweep.Sweepable

	rootsInput := organizations.ListRootsInput{}
	pages := organizations.NewListRootsPaginator(conn, &rootsInput)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, root := range page.Roots {
			childOUs, err := sweepListOrganizationalUnits(ctx, client, r, aws.ToString(root.Id))
			if err != nil {
				return nil, err
			}
			sweepResources = append(sweepResources, childOUs...)
		}
	}

	return sweepResources, nil
}

func sweepListOrganizationalUnits(ctx context.Context, client *conns.AWSClient, r *schema.Resource, parentID string) ([]sweep.Sweepable, error) {
	conn := client.OrganizationsClient(ctx)

	var sweepResources []sweep.Sweepable

	input := organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(parentID),
	}
	pages := organizations.NewListOrganizationalUnitsForParentPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, ou := range page.OrganizationalUnits {
			childOUs, err := sweepListOrganizationalUnits(ctx, client, r, aws.ToString(ou.Id))
			if err != nil {
				return nil, err
			}
			sweepResources = append(sweepResources, childOUs...)

			d := r.Data(nil)
			d.SetId(aws.ToString(ou.Id))

			sweepResources = append(sweepResources, newOrganizationalUnitSweeper(r, d, client))
		}
	}

	return sweepResources, nil
}

type organizationalUnitSweeper struct {
	d         *schema.ResourceData
	sweepable sweep.Sweepable
}

func newOrganizationalUnitSweeper(resource *schema.Resource, d *schema.ResourceData, client *conns.AWSClient) *organizationalUnitSweeper {
	return &organizationalUnitSweeper{
		d:         d,
		sweepable: sdk.NewSweepResource(resource, d, client),
	}
}

func (ous organizationalUnitSweeper) Delete(ctx context.Context, optFns ...tfresource.OptionsFunc) error {
	if err := ous.sweepable.Delete(ctx, optFns...); err != nil {
		if strings.Contains(err.Error(), "OrganizationalUnitNotEmptyException:") {
			tflog.Info(ctx, "Ignoring error", map[string]any{
				"error": err.Error(),
			})
			return nil
		}
		return err
	}
	return nil
}
