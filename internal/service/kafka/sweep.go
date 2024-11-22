// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_msk_cluster", &resource.Sweeper{
		Name: "aws_msk_cluster",
		F:    sweepClusters,
		Dependencies: []string{
			"aws_mskconnect_connector",
		},
	})

	resource.AddTestSweepers("aws_msk_configuration", &resource.Sweeper{
		Name: "aws_msk_configuration",
		F:    sweepConfigurations,
		Dependencies: []string{
			"aws_msk_cluster",
		},
	})
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &kafka.ListClustersV2Input{}
	conn := client.KafkaClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kafka.NewListClustersV2Paginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MSK Cluster sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing MSK Clusters (%s): %w", region, err)
		}

		for _, v := range page.ClusterInfoList {
			arn := aws.ToString(v.ClusterArn)

			if state := v.State; state == types.ClusterStateDeleting {
				log.Printf("[INFO] Skipping MSK Cluster %s: State=%s", arn, state)
				continue
			}

			r := resourceCluster()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MSK Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.KafkaClient(ctx)
	input := &kafka.ListConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kafka.NewListConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MSK Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing MSK Configurations (%s): %w", region, err)
		}

		for _, v := range page.Configurations {
			arn := aws.ToString(v.Arn)

			if state := v.State; state == types.ConfigurationStateDeleting {
				log.Printf("[INFO] Skipping MSK Configuration %s: State=%s", arn, state)
				continue
			}

			r := resourceConfiguration()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MSK Configurations (%s): %w", region, err)
	}

	return nil
}
