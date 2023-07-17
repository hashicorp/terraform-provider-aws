// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package sfn

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.SFNConn(ctx)
	input := &sfn.ListActivitiesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListActivitiesPagesWithContext(ctx, input, func(page *sfn.ListActivitiesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Activities {
			r := ResourceActivity()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ActivityArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Step Functions Activity sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Step Functions Activities (%s): %w", region, err)
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
	conn := client.SFNConn(ctx)
	input := &sfn.ListStateMachinesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListStateMachinesPagesWithContext(ctx, input, func(page *sfn.ListStateMachinesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.StateMachines {
			r := ResourceStateMachine()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.StateMachineArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Step Functions State Machine sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Step Functions State Machines (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Step Functions State Machines (%s): %w", region, err)
	}

	return nil
}
