// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53profiles

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53profiles"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_route53profiles_profile", &resource.Sweeper{
		Name: "aws_route53profiles_profile",
		F:    sweepProfiles,
		Dependencies: []string{
			"aws_route53profiles_association",
		},
	})

	resource.AddTestSweepers("aws_route53profiles_association", &resource.Sweeper{
		Name: "aws_route53profiles_association",
		F:    sweepProfileAssociations,
	})
}

func sweepProfiles(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.Route53ProfilesClient(ctx)
	input := &route53profiles.ListProfilesInput{}
	var sweepResources []sweep.Sweepable

	pages := route53profiles.NewListProfilesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			tflog.Warn(ctx, "Skipping sweeper", map[string]any{
				"error": err.Error(),
			})
			return nil
		}
		if err != nil {
			return fmt.Errorf("listing Route53Profiles Profiles (%s): %w", region, err)
		}

		for _, profile := range page.ProfileSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceProfile, client,
				framework.NewAttribute(names.AttrID, aws.ToString(profile.Id))))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Route53Profiles Profiles for %s: %w", region, err)
	}

	return nil
}

func sweepProfileAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.Route53ProfilesClient(ctx)
	input := &route53profiles.ListProfileAssociationsInput{}
	var sweepResources []sweep.Sweepable

	pages := route53profiles.NewListProfileAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			tflog.Warn(ctx, "Skipping sweeper", map[string]any{
				"error": err.Error(),
			})
			return nil
		}
		if err != nil {
			return fmt.Errorf("listing Route53Profiles Profiles (%s): %w", region, err)
		}

		for _, associations := range page.ProfileAssociations {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceProfile, client,
				framework.NewAttribute(names.AttrID, aws.ToString(associations.Id))))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Route53Profiles Profile Associations for %s: %w", region, err)
	}

	return nil
}
