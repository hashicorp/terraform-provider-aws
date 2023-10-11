// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package kafka

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.KafkaConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListClustersV2PagesWithContext(ctx, input, func(page *kafka.ListClustersV2Output, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ClusterInfoList {
			arn := aws.StringValue(v.ClusterArn)

			if state := aws.StringValue(v.State); state == kafka.ClusterStateDeleting {
				log.Printf("[INFO] Skipping MSK Cluster %s: State=%s", arn, state)
				continue
			}

			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MSK Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MSK Clusters (%s): %w", region, err)
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
	conn := client.KafkaConn(ctx)
	input := &kafka.ListConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListConfigurationsPagesWithContext(ctx, input, func(page *kafka.ListConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Configurations {
			arn := aws.StringValue(v.Arn)

			if state := aws.StringValue(v.State); state == kafka.ConfigurationStateDeleting {
				log.Printf("[INFO] Skipping MSK Configuration %s: State=%s", arn, state)
				continue
			}

			r := ResourceConfiguration()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MSK Configuration sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error listing MSK Configurations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MSK Configurations (%s): %w", region, err)
	}

	return nil
}
