// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cleanrooms

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cleanrooms"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_cleanrooms_collaboration", &resource.Sweeper{
		Name: "aws_cleanrooms_collaboration",
		F:    sweepCollaborations,
	})
	resource.AddTestSweepers("aaws_cleanrooms_configured_table", &resource.Sweeper{
		Name: "aws_cleanrooms_configured_table",
		F:    sweepConfiguredTables,
	})
	resource.AddTestSweepers("aaws_cleanrooms_membership", &resource.Sweeper{
		Name: "aws_cleanrooms_membership",
		F:    sweepMemberships,
	})
}

func sweepCollaborations(region string) error {
	ctx := sweep.Context(region)

	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CleanRoomsClient(ctx)
	input := &cleanrooms.ListCollaborationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cleanrooms.NewListCollaborationsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Cleanrooms Collaborations sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Cleanrooms Collaborations: %w", err)
		}

		for _, c := range page.CollaborationList {
			id := aws.ToString(c.Id)
			r := ResourceCollaboration()
			d := r.Data(nil)
			d.SetId(id)

			log.Printf("[INFO] Deleting Cleanrooms Collaboration: %s", id)
			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Cleanrooms Collaborations for %s: %w", region, err)
	}

	return nil
}

func sweepConfiguredTables(region string) error {
	ctx := sweep.Context(region)

	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CleanRoomsClient(ctx)
	input := &cleanrooms.ListConfiguredTablesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cleanrooms.NewListConfiguredTablesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Cleanrooms Configured Tables sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Cleanrooms Configured Tables: %w", err)
		}

		for _, c := range page.ConfiguredTableSummaries {
			id := aws.ToString(c.Id)
			r := ResourceConfiguredTable()
			d := r.Data(nil)
			d.SetId(id)

			log.Printf("[INFO] Deleting Cleanrooms Configured Table: %s", id)
			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Cleanrooms Configured Tables for %s: %w", region, err)
	}

	return nil
}

func sweepMemberships(region string) error {
	ctx := sweep.Context(region)

	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CleanRoomsClient(ctx)
	input := &cleanrooms.ListMembershipsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cleanrooms.NewListMembershipsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Cleanrooms Memberships sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Cleanrooms Memberships: %w", err)
		}

		for _, c := range page.MembershipSummaries {
			id := aws.ToString(c.Id)

			log.Printf("[INFO] Deleting Cleanrooms Membership: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceMembership, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Cleanrooms Memberships for %s: %w", region, err)
	}

	return nil
}
