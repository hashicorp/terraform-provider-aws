// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package ram

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_ram_resource_share", &resource.Sweeper{
		Name: "aws_ram_resource_share",
		F:    sweepResourceShares,
	})
}

func sweepResourceShares(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.RAMConn(ctx)
	input := &ram.GetResourceSharesInput{
		ResourceOwner: aws.String(ram.ResourceOwnerSelf),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.GetResourceSharesPagesWithContext(ctx, input, func(page *ram.GetResourceSharesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResourceShares {
			if aws.StringValue(v.Status) == ram.ResourceShareStatusDeleted {
				continue
			}

			r := ResourceResourceShare()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ResourceShareArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RAM Resource Share sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RAM Resource Shares (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RAM Resource Shares (%s): %w", region, err)
	}

	return nil
}
