// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mq

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mq"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_mq_configuration", &resource.Sweeper{
		Name: "aws_mq_configuration",
		F:    sweepConfigurations,
	})

	resource.AddTestSweepers("aws_mq_broker", &resource.Sweeper{
		Name: "aws_mq_broker",
		F:    sweepBrokers,
	})
}

func sweepConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	input := &mq.ListConfigurationsInput{}
	conn := client.MQClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.ListConfigurations(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MQ Configuration sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MQ Configurations (%s): %w", region, err)
	}

	for _, v := range output.Configurations {
		r := resourceConfiguration()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.Id))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MQ Configurations (%s): %w", region, err)
	}

	return nil
}

func sweepBrokers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	input := &mq.ListBrokersInput{MaxResults: aws.Int32(100)}
	conn := client.MQClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := mq.NewListBrokersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping MQ Broker sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing MQ Brokers (%s): %w", region, err)
		}

		for _, v := range page.BrokerSummaries {
			r := resourceBroker()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.BrokerId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MQ Brokers (%s): %w", region, err)
	}

	return nil
}
