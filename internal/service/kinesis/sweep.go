//go:build sweep
// +build sweep

package kinesis

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_kinesis_stream", &resource.Sweeper{
		Name: "aws_kinesis_stream",
		F:    sweepStreams,
	})
}

func sweepStreams(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).KinesisConn
	input := &kinesis.ListStreamsInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListStreamsPages(input, func(page *kinesis.ListStreamsOutput, lastPage bool) bool {
		for _, streamName := range page.StreamNames {
			if streamName == nil {
				continue
			}

			r := ResourceStream()
			d := r.Data(nil)
			d.Set("name", streamName)
			d.Set("enforce_consumer_deletion", true)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Kinesis Stream (%s): %w", aws.StringValue(streamName), err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Kinesis Stream sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error listing Kinesis Streams: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}
