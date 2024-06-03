// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudhsmv2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_cloudhsm_v2_cluster", &resource.Sweeper{
		Name:         "aws_cloudhsm_v2_cluster",
		F:            sweepClusters,
		Dependencies: []string{"aws_cloudhsm_v2_hsm"},
	})

	resource.AddTestSweepers("aws_cloudhsm_v2_hsm", &resource.Sweeper{
		Name: "aws_cloudhsm_v2_hsm",
		F:    sweepHSMs,
	})
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudHSMV2Client(ctx)
	input := &cloudhsmv2.DescribeClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudhsmv2.NewDescribeClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudHSMv2 Cluster sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudHSMv2 Clusters (%s): %w", region, err)
		}

		for _, v := range page.Clusters {
			r := resourceCluster()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ClusterId))
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudHSMv2 Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepHSMs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudHSMV2Client(ctx)
	input := &cloudhsmv2.DescribeClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudhsmv2.NewDescribeClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudHSMv2 HSM sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudHSMv2 Clusters (%s): %w", region, err)
		}

		for _, v := range page.Clusters {
			clusterID := aws.ToString(v.ClusterId)

			for _, v := range v.Hsms {
				r := resourceHSM()
				d := r.Data(nil)
				d.SetId(aws.ToString(v.HsmId))
				d.Set("cluster_id", clusterID)
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudHSMv2 HSMs (%s): %w", region, err)
	}

	return nil
}
