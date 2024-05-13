// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datasync/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
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
	conn := client.DataSyncClient(ctx)
	input := &datasync.ListAgentsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := datasync.NewListAgentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing DataSync Agents (%s): %w", region, err)
		}

		for _, v := range page.Agents {
			r := ResourceAgent()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AgentArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
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
	conn := client.DataSyncClient(ctx)
	input := &datasync.ListLocationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := datasync.NewListLocationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing DataSync Locations (%s): %w", region, err)
		}

		for _, v := range page.Locations {
			sweepable := &sweepableLocation{
				arn:  aws.ToString(v.LocationArn),
				conn: conn,
			}

			sweepResources = append(sweepResources, sweepable)
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DataSync Locations (%s): %w", region, err)
	}

	return nil
}

type sweepableLocation struct {
	arn  string
	conn *datasync.Client
}

func (sweepable *sweepableLocation) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	log.Printf("[DEBUG] Deleting DataSync Location: %s", sweepable.arn)
	_, err := sweepable.conn.DeleteLocation(ctx, &datasync.DeleteLocationInput{
		LocationArn: aws.String(sweepable.arn),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
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
	conn := client.DataSyncClient(ctx)
	input := &datasync.ListTasksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := datasync.NewListTasksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing DataSync Locations (%s): %w", region, err)
		}

		for _, v := range page.Tasks {
			r := resourceTask()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TaskArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DataSync Tasks (%s): %w", region, err)
	}

	return nil
}
