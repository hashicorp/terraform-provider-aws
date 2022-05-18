//go:build sweep
// +build sweep

package mq

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/go-multierror"
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).MQConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &mq.ListBrokersInput{MaxResults: aws.Int64(100)}

	err = conn.ListBrokersPages(input, func(page *mq.ListBrokersResponse, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, bs := range page.BrokerSummaries {
			r := ResourceBroker()
			d := r.Data(nil)

			id := aws.StringValue(bs.BrokerId)
			d.SetId(id)

			if err != nil {
				err := fmt.Errorf("error reading MQ Broker (%s): %w", id, err)
				log.Printf("[ERROR] %s", err)
				errs = multierror.Append(errs, err)
				continue
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing MQ Broker for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping MQ Broker for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MQ Broker sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
