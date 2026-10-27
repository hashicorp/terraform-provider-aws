// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_cloudwatchomni_access_grant", sweepAccessGrants)
	awsv2.Register("aws_cloudwatchomni_access_profile", sweepAccessProfiles)
	awsv2.Register("aws_cloudwatchomni_dashboard", sweepDashboards)
	awsv2.Register("aws_cloudwatchomni_domain", sweepDomains, "aws_cloudwatchomni_space")
	awsv2.Register("aws_cloudwatchomni_organization_access_grant", sweepOrganizationAccessGrants)
	awsv2.Register("aws_cloudwatchomni_organization_domain", sweepOrganizationDomains, "aws_cloudwatchomni_space", "aws_cloudwatchomni_organization_access_grant")
	awsv2.Register("aws_cloudwatchomni_space", sweepSpaces,
		"aws_cloudwatchomni_access_grant",
		"aws_cloudwatchomni_access_profile",
		"aws_cloudwatchomni_dashboard",
	)
	awsv2.Register("aws_cloudwatchomni_view", sweepViews)
}

func sweepAccessGrants(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listRegionalAccessGrants(ctx, conn) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		if v.GrantType == awstypes.AccessGrantTypeServiceManaged {
			continue
		}

		sweepResources = append(sweepResources, framework.NewSweepResource(newAccessGrantResource, client,
			framework.NewAttribute("grant_id", aws.ToString(v.GrantId))),
		)
	}

	return sweepResources, nil
}

func sweepAccessProfiles(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudwatchomni.ListSpacesInput
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listSpaces(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		if aws.ToString(v.Region) != conn.Options().Region {
			continue
		}

		spaceID := aws.ToString(v.SpaceId)
		input := cloudwatchomni.ListAccessProfilesInput{
			SpaceId: aws.String(spaceID),
		}
		for v, err := range listAccessProfiles(ctx, conn, &input) {
			if err != nil {
				return nil, smarterr.NewError(err)
			}

			// Service-managed profiles cannot be deleted.
			if v.ProfileType == awstypes.AccessProfileTypeServiceManaged {
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newAccessProfileResource, client,
				framework.NewAttribute("profile_id", aws.ToString(v.ProfileId)),
				framework.NewAttribute("space_id", spaceID)),
			)
		}
	}

	return sweepResources, nil
}

func sweepDashboards(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudwatchomni.ListSpacesInput
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listSpaces(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		if aws.ToString(v.Region) != conn.Options().Region {
			continue
		}

		spaceID := aws.ToString(v.SpaceId)
		input := cloudwatchomni.ListOmniDashboardsInput{
			SpaceId: aws.String(spaceID),
		}
		for v, err := range listDashboards(ctx, conn, &input) {
			if err != nil {
				return nil, smarterr.NewError(err)
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newDashboardResource, client,
				framework.NewAttribute("dashboard_id", aws.ToString(v.DashboardId)),
				framework.NewAttribute("space_id", spaceID)),
			)
		}
	}

	return sweepResources, nil
}

func sweepDomains(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudwatchomni.ListDomainsInput
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listDomains(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		// Organization domains are swept by sweepOrganizationDomains.
		if isOrganizationDomainARN(aws.ToString(v.DomainArn)) {
			continue
		}

		sweepResources = append(sweepResources, framework.NewSweepResource(newDomainResource, client,
			framework.NewAttribute("domain_id", aws.ToString(v.DomainId))),
		)
	}

	return sweepResources, nil
}

func sweepOrganizationDomains(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudwatchomni.ListDomainsInput
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listDomains(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		domainARN := aws.ToString(v.DomainArn)
		if !isOrganizationDomainARN(domainARN) {
			continue
		}

		// Only the organization's management account can delete an organization domain, and only in its home region.
		parsedARN, err := arn.Parse(domainARN)
		if err != nil {
			return nil, smarterr.NewError(err)
		}
		if parsedARN.AccountID != client.AccountID(ctx) || parsedARN.Region != conn.Options().Region {
			continue
		}

		sweepResources = append(sweepResources, framework.NewSweepResource(newOrganizationDomainResource, client,
			framework.NewAttribute("domain_id", aws.ToString(v.DomainId))),
		)
	}

	return sweepResources, nil
}

func sweepOrganizationAccessGrants(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudwatchomni.ListDomainAccessGrantsForOrganizationInput
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	// Organization access grants are organization-wide and can only be deleted in their domain's home
	// Region, so sweep only grants whose domain is homed in this account and Region.
	localDomainIDs, err := findLocalOrganizationDomainIDs(ctx, conn, client.AccountID(ctx))
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	for v, err := range tfiter.ConcatValuesWithError(listOrganizationAccessGrantPages(ctx, conn, &input)) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		if v.GrantType == awstypes.AccessGrantTypeServiceManaged {
			continue
		}

		if _, ok := localDomainIDs[aws.ToString(v.DomainId)]; !ok {
			continue
		}

		sweepResources = append(sweepResources, framework.NewSweepResource(newOrganizationAccessGrantResource, client,
			framework.NewAttribute("grant_id", aws.ToString(v.GrantId))),
		)
	}

	return sweepResources, nil
}

func sweepSpaces(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudwatchomni.ListSpacesInput
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listSpaces(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		if aws.ToString(v.Region) != conn.Options().Region {
			continue
		}

		sweepResources = append(sweepResources, framework.NewSweepResource(newSpaceResource, client,
			framework.NewAttribute("space_id", aws.ToString(v.SpaceId))),
		)
	}

	return sweepResources, nil
}

func sweepViews(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := cloudwatchomni.ListViewsInput{
		Type: awstypes.ViewTypeUser,
	}
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listViews(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		sweepResources = append(sweepResources, framework.NewSweepResource(newViewResource, client,
			framework.NewAttribute(names.AttrName, aws.ToString(v.Name))),
		)
	}

	return sweepResources, nil
}
