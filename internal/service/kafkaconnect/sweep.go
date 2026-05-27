// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafkaconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_mskconnect_connector", sweepConnectors)
	awsv2.Register("aws_mskconnect_custom_plugin", sweepCustomPlugins, "aws_mskconnect_connector")
	awsv2.Register("aws_mskconnect_worker_configuration", sweepWorkerConfigurations, "aws_mskconnect_connector")
}

func sweepConnectors(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.KafkaConnectClient(ctx)
	var input kafkaconnect.ListConnectorsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kafkaconnect.NewListConnectorsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Connectors {
			r := resourceConnector()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ConnectorArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepCustomPlugins(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.KafkaConnectClient(ctx)
	var input kafkaconnect.ListCustomPluginsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kafkaconnect.NewListCustomPluginsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.CustomPlugins {
			r := resourceCustomPlugin()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.CustomPluginArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepWorkerConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.KafkaConnectClient(ctx)
	var input kafkaconnect.ListWorkerConfigurationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kafkaconnect.NewListWorkerConfigurationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.WorkerConfigurations {
			r := resourceWorkerConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.WorkerConfigurationArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
