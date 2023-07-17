// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package vpclattice

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	resource.AddTestSweepers("aws_vpclattice_service", &resource.Sweeper{
		Name: "aws_vpclattice_service",
		F:    sweepServices,
	})

	resource.AddTestSweepers("aws_vpclattice_service_network", &resource.Sweeper{
		Name: "aws_vpclattice_service_network",
		F:    sweepServiceNetworks,
		Dependencies: []string{
			"aws_vpclattice_service",
		},
	})
}

func sweepServices(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.VPCLatticeClient(ctx)
	input := &vpclattice.ListServicesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := vpclattice.NewListServicesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) || skipSweepErr(err) {
			log.Printf("[WARN] Skipping VPC Lattice Service sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing VPC Lattice Services (%s): %w", region, err)
		}

		for _, v := range page.Items {
			r := ResourceService()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping VPC Lattice Services (%s): %w", region, err)
	}

	return nil
}

func sweepServiceNetworks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.VPCLatticeClient(ctx)
	input := &vpclattice.ListServiceNetworksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := vpclattice.NewListServiceNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) || skipSweepErr(err) {
			log.Printf("[WARN] Skipping VPC Lattice Service Network sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing VPC Lattice Service Networks (%s): %w", region, err)
		}

		for _, v := range page.Items {
			r := ResourceServiceNetwork()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping VPC Lattice Service Networks (%s): %w", region, err)
	}

	return nil
}

func skipSweepErr(err error) bool {
	return tfawserr.ErrCodeEquals(err, "AccessDeniedException")
}
