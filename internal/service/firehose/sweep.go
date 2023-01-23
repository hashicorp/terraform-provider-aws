//go:build sweep
// +build sweep

package firehose

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
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
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).FirehoseConn()
	input := &firehose.ListDeliveryStreamsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listDeliveryStreamsPages(ctx, conn, input, func(page *firehose.ListDeliveryStreamsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DeliveryStreamNames {
			r := ResourceDeliveryStream()
			d := r.Data(nil)
			name := aws.StringValue(v)
			arn := arn.ARN{
				Partition: client.(*conns.AWSClient).Partition,
				Service:   firehose.ServiceName,
				Region:    client.(*conns.AWSClient).Region,
				AccountID: client.(*conns.AWSClient).AccountID,
				Resource:  fmt.Sprintf("deliverystream/%s", name),
			}.String()
			d.SetId(arn)
			d.Set("name", name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Kinesis Firehose Delivery Stream sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Kinesis Firehose Delivery Streams (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Kinesis Firehose Delivery Streams (%s): %w", region, err)
	}

	return nil
}
