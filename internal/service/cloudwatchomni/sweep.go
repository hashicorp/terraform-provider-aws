// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_cloudwatchomni_access_grant", sweepAccessGrants)
	awsv2.Register("aws_cloudwatchomni_access_profile", sweepAccessProfiles)
	awsv2.Register("aws_cloudwatchomni_domain", sweepDomains, "aws_cloudwatchomni_space")
	awsv2.Register("aws_cloudwatchomni_space", sweepSpaces, "aws_cloudwatchomni_access_grant", "aws_cloudwatchomni_access_profile")
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

	pages := cloudwatchomni.NewListSpacesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			if aws.ToString(v.Region) != conn.Options().Region {
				continue
			}

			spaceID := aws.ToString(v.SpaceId)
			input := cloudwatchomni.ListAccessProfilesInput{
				SpaceId: aws.String(spaceID),
			}
			pages := cloudwatchomni.NewListAccessProfilesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, v := range page.Items {
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
		}
	}

	return sweepResources, nil
}

func sweepDomains(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudwatchomni.ListDomainsInput
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := cloudwatchomni.NewListDomainsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			sweepResources = append(sweepResources, framework.NewSweepResource(newDomainResource, client,
				framework.NewAttribute("domain_id", aws.ToString(v.DomainId))),
			)
		}
	}

	return sweepResources, nil
}

func sweepSpaces(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudwatchomni.ListSpacesInput
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := cloudwatchomni.NewListSpacesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			if aws.ToString(v.Region) != conn.Options().Region {
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newSpaceResource, client,
				framework.NewAttribute("space_id", aws.ToString(v.SpaceId))),
			)
		}
	}

	return sweepResources, nil
}

func sweepViews(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := cloudwatchomni.ListViewsInput{
		Type: awstypes.ViewTypeUser,
	}
	conn := client.CloudWatchOmniClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := cloudwatchomni.NewListViewsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			sweepResources = append(sweepResources, framework.NewSweepResource(newViewResource, client,
				framework.NewAttribute(names.AttrName, aws.ToString(v.Name))),
			)
		}
	}

	return sweepResources, nil
}
