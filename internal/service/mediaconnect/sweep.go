// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediaconnect

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_mediaconnect_flow", &resource.Sweeper{
		Name: "aws_mediaconnect_flow",
		F:    sweepFlows,
	})
}

func sweepFlows(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.MediaConnectClient(ctx)
	in := &mediaconnect.ListFlowsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := mediaconnect.NewListFlowsPaginator(conn, in)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MediaConnect Flows sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving MediaConnect Flows (%s): %w", region, err)
		}

		for _, flow := range page.Flows {
			id := aws.ToString(flow.FlowArn)
			log.Printf("[INFO] Deleting MediaConnect Flows: %s", id)

			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceFlow, client,
				framework.NewAttribute("id", aws.ToString(flow.FlowArn)),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping MediaConnect Flows for %s: %w", region, err)
	}

	return nil
}
