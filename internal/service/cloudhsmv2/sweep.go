// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package cloudhsmv2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.CloudHSMV2Conn(ctx)
	input := &cloudhsmv2.DescribeClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeClustersPagesWithContext(ctx, input, func(page *cloudhsmv2.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			if cluster == nil {
				continue
			}

			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(aws.StringValue(cluster.ClusterId))
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudHSMv2 Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudHSMv2 Clusters (%s): %w", region, err)
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
	conn := client.CloudHSMV2Conn(ctx)
	input := &cloudhsmv2.DescribeClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeClustersPagesWithContext(ctx, input, func(page *cloudhsmv2.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			if cluster == nil {
				continue
			}

			for _, hsm := range cluster.Hsms {
				r := ResourceHSM()
				d := r.Data(nil)
				d.SetId(aws.StringValue(hsm.HsmId))
				d.Set("cluster_id", cluster.ClusterId)
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudHSMv2 HSM sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudHSMv2 HSMs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudHSMv2 HSMs (%s): %w", region, err)
	}

	return nil
}
