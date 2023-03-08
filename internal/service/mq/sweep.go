//go:build sweep
// +build sweep

package mq

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_mq_broker", &resource.Sweeper{
		Name: "aws_mq_broker",
		F:    sweepBrokers,
	})
}

func sweepBrokers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &mq.ListBrokersInput{MaxResults: aws.Int64(100)}
	conn := client.(*conns.AWSClient).MQConn()
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListBrokersPagesWithContext(ctx, input, func(page *mq.ListBrokersResponse, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.BrokerSummaries {
			r := ResourceBroker()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.BrokerId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MQ Broker sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MQ Brokers (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MQ Brokers (%s): %w", region, err)
	}

	return nil
}
