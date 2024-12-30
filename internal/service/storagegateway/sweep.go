// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_storagegateway_gateway", &resource.Sweeper{
		Name: "aws_storagegateway_gateway",
		F:    sweepGateways,
		Dependencies: []string{
			"aws_storagegateway_file_system_association",
		},
	})

	resource.AddTestSweepers("aws_storagegateway_tape_pool", &resource.Sweeper{
		Name: "aws_storagegateway_tape_pool",
		F:    sweepTapePools,
	})

	resource.AddTestSweepers("aws_storagegateway_file_system_association", &resource.Sweeper{
		Name: "aws_storagegateway_file_system_association",
		F:    sweepFileSystemAssociations,
	})
}

func sweepGateways(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.StorageGatewayClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := storagegateway.NewListGatewaysPaginator(conn, &storagegateway.ListGatewaysInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Storage Gateway Gateway sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Storage Gateway Gateways (%s): %w", region, err)
		}

		for _, gateway := range page.Gateways {
			r := resourceGateway()
			d := r.Data(nil)
			d.SetId(aws.ToString(gateway.GatewayARN))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Storage Gateway Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepTapePools(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.StorageGatewayClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := storagegateway.NewListTapePoolsPaginator(conn, &storagegateway.ListTapePoolsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Storage Gateway Tape Pool sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Storage Gateway Tape Pools (%s): %w", region, err)
		}

		for _, pool := range page.PoolInfos {
			r := resourceTapePool()
			d := r.Data(nil)
			d.SetId(aws.ToString(pool.PoolARN))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Storage Gateway Gateways (%s): %w", region, err)
	}

	return nil
}

func sweepFileSystemAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.StorageGatewayClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := storagegateway.NewListFileSystemAssociationsPaginator(conn, &storagegateway.ListFileSystemAssociationsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Storage Gateway File System Association sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Storage Gateway File System Associations (%s): %w", region, err)
		}

		for _, assoc := range page.FileSystemAssociationSummaryList {
			r := resourceFileSystemAssociation()
			d := r.Data(nil)
			d.SetId(aws.ToString(assoc.FileSystemAssociationARN))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Storage Gateway File System Associations (%s): %w", region, err)
	}

	return nil
}
