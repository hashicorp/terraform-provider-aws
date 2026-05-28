// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_msk_cluster", sweepClusters, "aws_mskconnect_connector")
	awsv2.Register("aws_msk_configuration", sweepConfigurations, "aws_msk_cluster")
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.KafkaClient(ctx)
	var input kafka.ListClustersV2Input
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kafka.NewListClustersV2Paginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
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

	return sweepResources, nil
}

func sweepConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.KafkaClient(ctx)
	var input kafka.ListConfigurationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kafka.NewListConfigurationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
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

	return sweepResources, nil
}
