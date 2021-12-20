//go:build sweep
// +build sweep

package autoscaling

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
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
		Dependencies: []string{"aws_autoscaling_group"},
		F:            sweepLaunchConfigurations,
	})
}

func sweepGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AutoScalingConn

	resp, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Auto Scaling Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Auto Scaling Groups in Sweeper: %s", err)
	}

	if len(resp.AutoScalingGroups) == 0 {
		log.Print("[DEBUG] No Auto Scaling Groups to sweep")
		return nil
	}

	for _, asg := range resp.AutoScalingGroups {
		deleteopts := autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: asg.AutoScalingGroupName,
			ForceDelete:          aws.Bool(true),
		}

		err = resource.Retry(5*time.Minute, func() *resource.RetryError {
			if _, err := conn.DeleteAutoScalingGroup(&deleteopts); err != nil {
				if awserr, ok := err.(awserr.Error); ok {
					switch awserr.Code() {
					case "InvalidGroup.NotFound":
						return nil
					case "ResourceInUse", "ScalingActivityInProgress":
						return resource.RetryableError(awserr)
					}
				}

				// Didn't recognize the error, so shouldn't retry.
				return resource.NonRetryableError(err)
			}
			// Successful delete
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func sweepLaunchConfigurations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).AutoScalingConn

	resp, err := conn.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping AutoScaling Launch Configuration sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving launch configuration: %s", err)
	}

	if len(resp.LaunchConfigurations) == 0 {
		log.Print("[DEBUG] No aws launch configurations to sweep")
		return nil
	}

	for _, lc := range resp.LaunchConfigurations {
		name := aws.StringValue(lc.LaunchConfigurationName)

		log.Printf("[INFO] Deleting Launch Configuration: %s", name)
		_, err := conn.DeleteLaunchConfiguration(
			&autoscaling.DeleteLaunchConfigurationInput{
				LaunchConfigurationName: aws.String(name),
			})
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidConfiguration.NotFound", "") || tfawserr.ErrMessageContains(err, "ValidationError", "") {
				return nil
			}
			return err
		}
	}

	return nil
}
