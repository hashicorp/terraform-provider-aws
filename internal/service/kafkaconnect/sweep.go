// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_mskconnect_connector", &resource.Sweeper{
		Name: "aws_mskconnect_connector",
		F:    sweepConnectors,
	})

	resource.AddTestSweepers("aws_mskconnect_custom_plugin", &resource.Sweeper{
		Name: "aws_mskconnect_custom_plugin",
		F:    sweepCustomPlugins,
		Dependencies: []string{
			"aws_mskconnect_connector",
		},
	})

	resource.AddTestSweepers("aws_mskconnect_worker_configuration", &resource.Sweeper{
		Name: "aws_mskconnect_worker_configuration",
		F:    sweepWorkerConfigurations,
		Dependencies: []string{
			"aws_mskconnect_connector",
		},
	})
}

func sweepConnectors(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.KafkaConnectClient(ctx)
	input := &kafkaconnect.ListConnectorsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kafkaconnect.NewListConnectorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MSK Connect Connector sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing MSK Connect Connectors (%s): %w", region, err)
		}

		for _, v := range page.Connectors {
			r := resourceConnector()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ConnectorArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MSK Connect Connectors (%s): %w", region, err)
	}

	return nil
}

func sweepCustomPlugins(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.KafkaConnectClient(ctx)
	input := &kafkaconnect.ListCustomPluginsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kafkaconnect.NewListCustomPluginsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MSK Connect Custom Plugin sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing MSK Connect Custom Plugins (%s): %w", region, err)
		}

		for _, v := range page.CustomPlugins {
			r := resourceCustomPlugin()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.CustomPluginArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MSK Connect Custom Plugins (%s): %w", region, err)
	}

	return nil
}

func sweepWorkerConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.KafkaConnectClient(ctx)
	input := &kafkaconnect.ListWorkerConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kafkaconnect.NewListWorkerConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MSK Connect Worker Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing MSK Connect Worker Configurations (%s): %w", region, err)
		}

		for _, v := range page.WorkerConfigurations {
			r := resourceWorkerConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.WorkerConfigurationArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MSK Connect Worker Configurations (%s): %w", region, err)
	}

	return nil
}
