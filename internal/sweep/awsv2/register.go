// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsv2

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/internal/log"
)

func Register(name string, f sweep.SweeperFn, dependencies ...string) {
	resource.AddTestSweepers(name, &resource.Sweeper{
		Name: name,
		F: func(region string) error {
			ctx := sweep.Context(region)
			ctx = log.WithResourceType(ctx, name)

			client, err := sweep.SharedRegionalSweepClient(ctx, region)
			if err != nil {
				return fmt.Errorf("getting client: %w", err)
			}
			tflog.Info(ctx, "listing resources")
			sweepResources, err := f(ctx, client)

			if SkipSweepError(err) {
				tflog.Warn(ctx, "Skipping sweeper", map[string]any{
					"error": err.Error(),
				})
				return nil
			}
			if err != nil {
				return fmt.Errorf("listing %q (%s): %w", name, region, err)
			}

			err = sweep.SweepOrchestrator(ctx, sweepResources)
			if err != nil {
				return fmt.Errorf("sweeping %q (%s): %w", name, region, err)
			}

			return nil
		},
		Dependencies: dependencies,
	})
}
