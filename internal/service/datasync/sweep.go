// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_datasync_agent", &resource.Sweeper{
		Name: "aws_datasync_agent",
		F:    sweepAgents,
		Dependencies: []string{
			"aws_datasync_location",
		},
	})

	// Pseudo-resource for any DataSync location resource type.
	resource.AddTestSweepers("aws_datasync_location", &resource.Sweeper{
		Name: "aws_datasync_location",
		F:    sweepLocations,
		Dependencies: []string{
			"aws_datasync_task",
		},
	})

	resource.AddTestSweepers("aws_datasync_task", &resource.Sweeper{
		Name: "aws_datasync_task",
		F:    sweepTasks,
	})
}

func sweepAgents(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.DataSyncConn(ctx)
	input := &datasync.ListAgentsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListAgentsPagesWithContext(ctx, input, func(page *datasync.ListAgentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Agents {
			r := ResourceAgent()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.AgentArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DataSync Agent sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing DataSync Agents (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DataSync Agents (%s): %w", region, err)
	}

	return nil
}

func sweepLocations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.DataSyncConn(ctx)
	input := &datasync.ListLocationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListLocationsPagesWithContext(ctx, input, func(page *datasync.ListLocationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Locations {
			sweepable := &sweepableLocation{
				arn:  aws.StringValue(v.LocationArn),
				conn: conn,
			}

			sweepResources = append(sweepResources, sweepable)
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DataSync Location sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing DataSync Locations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DataSync Locations (%s): %w", region, err)
	}

	return nil
}

type sweepableLocation struct {
	arn  string
	conn *datasync.DataSync
}

func (sweepable *sweepableLocation) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	log.Printf("[DEBUG] Deleting DataSync Location: %s", sweepable.arn)
	_, err := sweepable.conn.DeleteLocationWithContext(ctx, &datasync.DeleteLocationInput{
		LocationArn: aws.String(sweepable.arn),
	})

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting DataSync Location (%s): %w", sweepable.arn, err)
	}

	return nil
}

func sweepTasks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.DataSyncConn(ctx)
	input := &datasync.ListTasksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListTasksPagesWithContext(ctx, input, func(page *datasync.ListTasksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Tasks {
			r := resourceTask()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.TaskArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DataSync Task sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing DataSync Tasks (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DataSync Tasks (%s): %w", region, err)
	}

	return nil
}
