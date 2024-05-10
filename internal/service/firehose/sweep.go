// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package firehose

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_kinesis_firehose_delivery_stream", &resource.Sweeper{
		Name: "aws_kinesis_firehose_delivery_stream",
		F:    sweepDeliveryStreams,
	})
}

func sweepDeliveryStreams(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.FirehoseClient(ctx)
	input := &firehose.ListDeliveryStreamsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listDeliveryStreamsPages(ctx, conn, input, func(page *firehose.ListDeliveryStreamsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DeliveryStreamNames {
			r := resourceDeliveryStream()
			d := r.Data(nil)
			name := v
			arn := arn.ARN{
				Partition: client.Partition,
				Service:   "firehose",
				Region:    client.Region,
				AccountID: client.AccountID,
				Resource:  fmt.Sprintf("deliverystream/%s", name),
			}.String()
			d.SetId(arn)
			d.Set(names.AttrName, name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Kinesis Firehose Delivery Stream sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Kinesis Firehose Delivery Streams (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Kinesis Firehose Delivery Streams (%s): %w", region, err)
	}

	return nil
}
