// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_sfn_activity", &resource.Sweeper{
		Name: "aws_sfn_activity",
		F:    sweepActivities,
	})

	resource.AddTestSweepers("aws_sfn_state_machine", &resource.Sweeper{
		Name: "aws_sfn_state_machine",
		F:    sweepStateMachines,
	})
}

func sweepActivities(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.SFNClient(ctx)
	input := &sfn.ListActivitiesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sfn.NewListActivitiesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Step Functions Activity sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Step Functions Activities (%s): %w", region, err)
		}

		for _, v := range page.Activities {
			r := resourceActivity()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ActivityArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Step Functions Activities (%s): %w", region, err)
	}

	return nil
}

func sweepStateMachines(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.SFNClient(ctx)
	input := &sfn.ListStateMachinesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sfn.NewListStateMachinesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Step Functions State Machine sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Step Functions State Machines (%s): %w", region, err)
		}

		for _, v := range page.StateMachines {
			r := resourceStateMachine()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.StateMachineArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Step Functions State Machines (%s): %w", region, err)
	}

	return nil
}
