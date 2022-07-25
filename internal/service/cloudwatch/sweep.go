//go:build sweep
// +build sweep

package cloudwatch

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_composite_alarm", &resource.Sweeper{
		Name: "aws_cloudwatch_composite_alarm",
		F:    sweepCompositeAlarms,
	})
}

func sweepCompositeAlarms(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).CloudWatchConn
	ctx := context.Background()

	input := &cloudwatch.DescribeAlarmsInput{
		AlarmTypes: aws.StringSlice([]string{cloudwatch.AlarmTypeCompositeAlarm}),
	}

	var sweeperErrs *multierror.Error

	err = conn.DescribeAlarmsPagesWithContext(ctx, input, func(page *cloudwatch.DescribeAlarmsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, compositeAlarm := range page.CompositeAlarms {
			if compositeAlarm == nil {
				continue
			}

			name := aws.StringValue(compositeAlarm.AlarmName)

			log.Printf("[INFO] Deleting CloudWatch Composite Alarm: %s", name)

			r := ResourceCompositeAlarm()
			d := r.Data(nil)
			d.SetId(name)

			diags := r.DeleteContext(ctx, d, client)

			for i := range diags {
				if diags[i].Severity == diag.Error {
					log.Printf("[ERROR] %s", diags[i].Summary)
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf(diags[i].Summary))
					continue
				}
			}
		}

		return !isLast
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudWatch Composite Alarms sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CloudWatch Composite Alarms: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
