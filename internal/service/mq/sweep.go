// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mq

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_mq_configuration", sweepConfigurations, "aws_mq_broker")
	awsv2.Register("aws_mq_broker", sweepBrokers)
}

func sweepConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.MQClient(ctx)
	var input mq.ListConfigurationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := listConfigurationsPages(ctx, conn, &input, func(page *mq.ListConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Configurations {
			r := resourceConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepBrokers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.MQClient(ctx)
	input := mq.ListBrokersInput{MaxResults: aws.Int32(100)}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := mq.NewListBrokersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.BrokerSummaries {
			r := resourceBroker()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.BrokerId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
