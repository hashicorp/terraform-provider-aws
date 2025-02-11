// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emrcontainers

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emrcontainers"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emrcontainers/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
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
	conn := client.EMRContainersClient(ctx)
	input := &emrcontainers.ListVirtualClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := emrcontainers.NewListVirtualClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EMR Containers Virtual Cluster sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EMR Containers Virtual Clusters (%s): %w", region, err)
		}

		for _, v := range page.VirtualClusters {
			if v.State == awstypes.VirtualClusterStateTerminated {
				continue
			}

			r := resourceVirtualCluster()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
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
	conn := client.EMRContainersClient(ctx)
	input := &emrcontainers.ListJobTemplatesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := emrcontainers.NewListJobTemplatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EMR Containers Job Template sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EMR Containers Job Templates (%s): %w", region, err)
		}

		for _, v := range page.Templates {
			r := resourceJobTemplate()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EMR Containers Job Templates (%s): %w", region, err)
	}

	return nil
}
