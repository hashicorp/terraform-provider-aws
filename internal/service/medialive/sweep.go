// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package medialive

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	resource.AddTestSweepers("aws_medialive_channel", &resource.Sweeper{
		Name: "aws_medialive_channel",
		F:    sweepChannels,
	})

	resource.AddTestSweepers("aws_medialive_input", &resource.Sweeper{
		Name: "aws_medialive_input",
		F:    sweepInputs,
	})

	resource.AddTestSweepers("aws_medialive_input_security_group", &resource.Sweeper{
		Name: "aws_medialive_input_security_group",
		F:    sweepInputSecurityGroups,
		Dependencies: []string{
			"aws_medialive_input",
		},
	})

	resource.AddTestSweepers("aws_medialive_multiplex", &resource.Sweeper{
		Name: "aws_medialive_multiplex",
		F:    sweepMultiplexes,
	})
}

func sweepChannels(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.MediaLiveClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &medialive.ListChannelsInput{}

	pages := medialive.NewListChannelsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Println("[WARN] Skipping MediaLive Channels sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving MediaLive Channels: %w", err)
		}

		for _, channel := range page.Channels {
			id := aws.ToString(channel.Id)
			log.Printf("[INFO] Deleting MediaLive Channels: %s", id)

			r := ResourceChannel()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping MediaLive Channels for %s: %w", region, err)
	}

	return nil
}

func sweepInputs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.MediaLiveClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &medialive.ListInputsInput{}

	pages := medialive.NewListInputsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Println("[WARN] Skipping MediaLive Inputs sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving MediaLive Inputs: %w", err)
		}

		for _, input := range page.Inputs {
			id := aws.ToString(input.Id)
			log.Printf("[INFO] Deleting MediaLive Input: %s", id)

			r := ResourceInput()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping MediaLive Inputs for %s: %w", region, err)
	}

	return nil
}

func sweepInputSecurityGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.MediaLiveClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &medialive.ListInputSecurityGroupsInput{}

	pages := medialive.NewListInputSecurityGroupsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Println("[WARN] Skipping MediaLive Input Security Groups sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving MediaLive Input Security Groups: %w", err)
		}

		for _, group := range page.InputSecurityGroups {
			id := aws.ToString(group.Id)
			log.Printf("[INFO] Deleting MediaLive Input Security Group: %s", id)

			r := ResourceInputSecurityGroup()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping MediaLive Input Security Groups for %s: %w", region, err)
	}

	return nil
}

func sweepMultiplexes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.MediaLiveClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &medialive.ListMultiplexesInput{}

	pages := medialive.NewListMultiplexesPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Println("[WARN] Skipping MediaLive Multiplexes sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving MediaLive Multiplexes: %w", err)
		}

		for _, multiplex := range page.Multiplexes {
			id := aws.ToString(multiplex.Id)
			log.Printf("[INFO] Deleting MediaLive Multiplex: %s", id)

			r := ResourceMultiplex()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping MediaLive Multiplexes for %s: %w", region, err)
	}

	return nil
}
