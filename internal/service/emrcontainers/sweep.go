// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package emrcontainers

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_emrcontainers_virtual_cluster", &resource.Sweeper{
		Name: "aws_emrcontainers_virtual_cluster",
		F:    sweepVirtualClusters,
	})

	resource.AddTestSweepers("aws_emrcontainers_job_template", &resource.Sweeper{
		Name: "aws_emrcontainers_job_template",
		F:    sweepJobTemplates,
	})
}

func sweepVirtualClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EMRContainersConn(ctx)
	input := &emrcontainers.ListVirtualClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListVirtualClustersPagesWithContext(ctx, input, func(page *emrcontainers.ListVirtualClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VirtualClusters {
			if aws.StringValue(v.State) == emrcontainers.VirtualClusterStateTerminated {
				continue
			}

			r := ResourceVirtualCluster()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EMR Containers Virtual Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EMR Containers Virtual Clusters (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EMR Containers Virtual Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepJobTemplates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EMRContainersConn(ctx)
	input := &emrcontainers.ListJobTemplatesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListJobTemplatesPagesWithContext(ctx, input, func(page *emrcontainers.ListJobTemplatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Templates {
			r := ResourceJobTemplate()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EMR Containers Job Template sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EMR Containers Job Templates (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EMR Containers Job Templates (%s): %w", region, err)
	}

	return nil
}
