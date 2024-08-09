// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeconnections

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codeconnections"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_codeconnections_connection", &resource.Sweeper{
		Name: "aws_codeconnections_connection",
		F:    sweepConnections,
	})

	resource.AddTestSweepers("aws_codeconnections_host", &resource.Sweeper{
		Name: "aws_codeconnections_host",
		F:    sweepHosts,
		Dependencies: []string{
			"aws_codeconnections_connection",
		},
	})
}

func sweepConnections(region string) error {
	ctx := sweep.Context(region)
	if region == names.USGovEast1RegionID || region == names.USGovWest1RegionID {
		log.Printf("[WARN] Skipping CodeConnections Connection sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CodeConnectionsClient(ctx)
	input := &codeconnections.ListConnectionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := codeconnections.NewListConnectionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CodeConnections Connection sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CodeConnections Connections (%s): %w", region, err)
		}

		for _, v := range page.Connections {
			r := resourceConnection()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ConnectionArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeConnections Connections (%s): %w", region, err)
	}

	return nil
}

func sweepHosts(region string) error {
	ctx := sweep.Context(region)
	if region == names.USGovEast1RegionID || region == names.USGovWest1RegionID {
		log.Printf("[WARN] Skipping CodeConnections Host sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CodeConnectionsClient(ctx)
	input := &codeconnections.ListHostsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := codeconnections.NewListHostsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CodeConnections Host sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CodeConnections Hosts (%s): %w", region, err)
		}

		for _, v := range page.Hosts {
			r := resourceHost()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.HostArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeConnections Hosts (%s): %w", region, err)
	}

	return nil
}
