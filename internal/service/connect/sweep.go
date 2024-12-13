// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_connect_instance", &resource.Sweeper{
		Name: "aws_connect_instance",
		F:    sweepInstances,
	})
}

func sweepInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	const maxResults = 10
	input := &connect.ListInstancesInput{
		MaxResults: aws.Int32(maxResults),
	}
	conn := client.ConnectClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := connect.NewListInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Connect Instance sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Connect Instances (%s): %w", region, err)
		}

		for _, v := range page.InstanceSummaryList {
			r := resourceInstance()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Connect Instances (%s): %w", region, err)
	}

	return nil
}
