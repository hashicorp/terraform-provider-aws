// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package resourcegroups

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_resourcegroups_group", &resource.Sweeper{
		Name: "aws_resourcegroups_group",
		F:    sweepGroups,
	})
}

func sweepGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ResourceGroupsConn(ctx)
	input := &resourcegroups.ListGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListGroupsPagesWithContext(ctx, input, func(page *resourcegroups.ListGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.GroupIdentifiers {
			r := ResourceGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.GroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Resource Groups Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Resource Groups Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Resource Groups Groups (%s): %w", region, err)
	}

	return nil
}
