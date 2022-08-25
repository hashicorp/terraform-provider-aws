//go:build sweep
// +build sweep

package elbv2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_lb", &resource.Sweeper{
		Name: "aws_lb",
		F:    sweepLoadBalancers,
		Dependencies: []string{
			"aws_api_gateway_vpc_link",
			"aws_vpc_endpoint_service",
		},
	})

	resource.AddTestSweepers("aws_lb_target_group", &resource.Sweeper{
		Name: "aws_lb_target_group",
		F:    sweepTargetGroups,
		Dependencies: []string{
			"aws_lb",
		},
	})
}

func sweepLoadBalancers(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ELBV2Conn

	var sweeperErrs *multierror.Error
	err = conn.DescribeLoadBalancersPages(&elbv2.DescribeLoadBalancersInput{}, func(page *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
		if page == nil || len(page.LoadBalancers) == 0 {
			log.Print("[DEBUG] No LBs to sweep")
			return false
		}

		for _, loadBalancer := range page.LoadBalancers {
			name := aws.StringValue(loadBalancer.LoadBalancerName)

			log.Printf("[INFO] Deleting LB: %s", name)
			_, err := conn.DeleteLoadBalancer(&elbv2.DeleteLoadBalancerInput{
				LoadBalancerArn: loadBalancer.LoadBalancerArn,
			})
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("failed to delete LB (%s): %w", name, err))
				continue
			}
		}
		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping LB sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving LBs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepTargetGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ELBV2Conn

	err = conn.DescribeTargetGroupsPages(&elbv2.DescribeTargetGroupsInput{}, func(page *elbv2.DescribeTargetGroupsOutput, lastPage bool) bool {
		if page == nil || len(page.TargetGroups) == 0 {
			log.Print("[DEBUG] No LB Target Groups to sweep")
			return false
		}

		for _, targetGroup := range page.TargetGroups {
			name := aws.StringValue(targetGroup.TargetGroupName)

			log.Printf("[INFO] Deleting LB Target Group: %s", name)
			_, err := conn.DeleteTargetGroup(&elbv2.DeleteTargetGroupInput{
				TargetGroupArn: targetGroup.TargetGroupArn,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete LB Target Group (%s): %s", name, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping LB Target Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("retrieving LB Target Groups: %w", err)
	}
	return nil
}
