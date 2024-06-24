// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_resourceexplorer2_index", &resource.Sweeper{
		Name: "aws_resourceexplorer2_index",
		F:    sweepIndexes,
	})
}

func sweepIndexes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ResourceExplorer2Client(ctx)
	input := &resourceexplorer2.ListIndexesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := resourceexplorer2.NewListIndexesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Resource Explorer Index sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Resource Explorer Indexes (%s): %w", region, err)
		}

		for _, v := range page.Indexes {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceIndex, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.Arn)),
			))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Resource Explorer Indexes (%s): %w", region, err)
	}

	return nil
}
