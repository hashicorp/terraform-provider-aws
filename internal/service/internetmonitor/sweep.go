// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package internetmonitor

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/internetmonitor"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	resource.AddTestSweepers("aws_internetmonitor_monitor", &resource.Sweeper{
		Name: "aws_internetmonitor_monitor",
		F:    sweepMonitors,
	})
}

func sweepMonitors(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &internetmonitor.ListMonitorsInput{}
	conn := client.InternetMonitorClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := internetmonitor.NewListMonitorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Internet Monitor Monitor sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Internet Monitor Monitors (%s): %w", region, err)
		}

		for _, v := range page.Monitors {
			r := resourceMonitor()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.MonitorName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Internet Monitor Monitors (%s): %w", region, err)
	}

	return nil
}
