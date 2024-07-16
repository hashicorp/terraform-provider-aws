// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_transfer_server", &resource.Sweeper{
		Name: "aws_transfer_server",
		F:    sweepServers,
	})

	resource.AddTestSweepers("aws_transfer_workflow", &resource.Sweeper{
		Name: "aws_transfer_workflow",
		F:    sweepWorkflows,
		Dependencies: []string{
			"aws_transfer_server",
		},
	})
}

func sweepServers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.TransferClient(ctx)
	input := &transfer.ListServersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := transfer.NewListServersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Transfer Server sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Transfer Servers (%s): %w", region, err)
		}

		for _, server := range page.Servers {
			r := resourceServer()
			d := r.Data(nil)
			d.SetId(aws.ToString(server.ServerId))
			d.Set(names.AttrForceDestroy, true) // In lieu of an aws_transfer_user sweeper.
			d.Set("identity_provider_type", server.IdentityProviderType)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Transfer Servers (%s): %w", region, err)
	}

	return nil
}

func sweepWorkflows(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.TransferClient(ctx)
	input := &transfer.ListWorkflowsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := transfer.NewListWorkflowsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Transfer Workflow sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Transfer Workflows (%s): %w", region, err)
		}

		for _, server := range page.Workflows {
			r := resourceWorkflow()
			d := r.Data(nil)
			d.SetId(aws.ToString(server.WorkflowId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Transfer Workflows (%s): %w", region, err)
	}

	return nil
}
