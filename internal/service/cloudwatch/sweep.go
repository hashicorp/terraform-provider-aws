//go:build sweep
// +build sweep

package cloudwatch

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
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
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudWatchConn()
	input := &cloudwatch.DescribeAlarmsInput{
		AlarmTypes: aws.StringSlice([]string{cloudwatch.AlarmTypeCompositeAlarm}),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeAlarmsPagesWithContext(ctx, input, func(page *cloudwatch.DescribeAlarmsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CompositeAlarms {
			r := ResourceCompositeAlarm()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.AlarmName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] SkippingCloudWatch Composite Alarm sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudWatch Composite Alarms (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudWatch Composite Alarms (%s): %w", region, err)
	}

	return nil
}
