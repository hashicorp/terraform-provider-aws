// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package rum

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchrum"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_rum_app_monitor", &resource.Sweeper{
		Name: "aws_rum_app_monitor",
		F:    sweepAppMonitors,
	})
}

func sweepAppMonitors(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.RUMConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.ListAppMonitorsPagesWithContext(ctx, &cloudwatchrum.ListAppMonitorsInput{}, func(resp *cloudwatchrum.ListAppMonitorsOutput, lastPage bool) bool {
		if len(resp.AppMonitorSummaries) == 0 {
			log.Print("[DEBUG] No RUM App Monitors to sweep")
			return !lastPage
		}

		for _, c := range resp.AppMonitorSummaries {
			r := ResourceAppMonitor()
			d := r.Data(nil)
			d.SetId(aws.StringValue(c.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing RUM App Monitors: %w", err))
		// in case work can be done, don't jump out yet
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping RUM App Monitors for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping RUM App Monitor sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}
