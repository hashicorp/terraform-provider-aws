//go:build sweep
// +build sweep

package firehose

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_kinesis_firehose_delivery_stream", &resource.Sweeper{
		Name: "aws_kinesis_firehose_delivery_stream",
		F:    sweepDeliveryStreams,
	})
}

func sweepDeliveryStreams(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).FirehoseConn
	input := &firehose.ListDeliveryStreamsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	for {
		page, err := conn.ListDeliveryStreams(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Kinesis Firehose Delivery Streams sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Kinesis Firehose Delivery Streams: %w", err)
		}

		for _, sn := range page.DeliveryStreamNames {
			r := ResourceDeliveryStream()
			d := r.Data(nil)
			d.SetId("???")
			d.Set("name", sn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if !aws.BoolValue(page.HasMoreDeliveryStreams) {
			break
		}
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Kinesis Firehose Delivery Streams (%s): %w", region, err)
	}

	return nil
}
