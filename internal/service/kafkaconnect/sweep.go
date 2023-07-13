// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package kafkaconnect

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
}

func sweepConnectors(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.KafkaConnectConn(ctx)
	input := &kafkaconnect.ListConnectorsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListConnectorsPagesWithContext(ctx, input, func(page *kafkaconnect.ListConnectorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Connectors {
			r := ResourceConnector()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ConnectorArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MSK Connect Connector sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MSK Connect Connectors (%s): %w", region, err)
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
	conn := client.KafkaConnectConn(ctx)
	input := &kafkaconnect.ListCustomPluginsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListCustomPluginsPagesWithContext(ctx, input, func(page *kafkaconnect.ListCustomPluginsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CustomPlugins {
			r := ResourceCustomPlugin()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.CustomPluginArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MSK Connect Custom Plugin sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MSK Connect Custom Plugins (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MSK Connect Custom Plugins (%s): %w", region, err)
	}

	return nil
}
