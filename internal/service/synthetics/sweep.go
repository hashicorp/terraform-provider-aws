//go:build sweep
// +build sweep

package synthetics

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_synthetics_canary", &resource.Sweeper{
		Name: "aws_synthetics_canary",
		F:    sweepCanaries,
		Dependencies: []string{
			"aws_lambda_function",
			"aws_lambda_layer",
			"aws_cloudwatch_log_group",
		},
	})
}

func sweepCanaries(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SyntheticsConn
	input := &synthetics.DescribeCanariesInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.DescribeCanaries(input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Synthetics Canary sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Synthetics Canaries: %w", err)
		}

		for _, canary := range output.Canaries {
			name := aws.StringValue(canary.Name)
			log.Printf("[INFO] Deleting Synthetics Canary: %s", name)

			r := ResourceCanary()
			d := r.Data(nil)
			d.SetId(name)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}
