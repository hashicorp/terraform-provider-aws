//go:build sweep
// +build sweep

package autoscaling

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_autoscaling_group", &resource.Sweeper{
		Name: "aws_autoscaling_group",
		F:    sweepGroups,
	})

	resource.AddTestSweepers("aws_launch_configuration", &resource.Sweeper{
		Name:         "aws_launch_configuration",
		F:            sweepLaunchConfigurations,
		Dependencies: []string{"aws_autoscaling_group"},
	})
}

func sweepGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AutoScalingConn
	input := &autoscaling.DescribeAutoScalingGroupsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeAutoScalingGroupsPages(input, func(page *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AutoScalingGroups {
			r := ResourceGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.AutoScalingGroupName))
			d.Set("force_delete", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Auto Scaling Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Auto Scaling Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Auto Scaling Groups (%s): %w", region, err)
	}

	return nil
}

func sweepLaunchConfigurations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AutoScalingConn
	input := &autoscaling.DescribeLaunchConfigurationsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeLaunchConfigurationsPages(input, func(page *autoscaling.DescribeLaunchConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LaunchConfigurations {
			r := ResourceLaunchConfiguration()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.LaunchConfigurationName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Auto Scaling Launch Configuration sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Auto Scaling Launch Configurations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Auto Scaling Launch Configurations (%s): %w", region, err)
	}

	return nil
}
