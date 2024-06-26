// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_lb", &resource.Sweeper{
		Name: "aws_lb",
		F:    sweepLoadBalancers,
		Dependencies: []string{
			"aws_api_gateway_vpc_link",
			"aws_vpc_endpoint_service",
			"aws_lb_listener",
		},
	})

	resource.AddTestSweepers("aws_lb_target_group", &resource.Sweeper{
		Name: "aws_lb_target_group",
		F:    sweepTargetGroups,
		Dependencies: []string{
			"aws_lb",
		},
	})

	resource.AddTestSweepers("aws_lb_listener", &resource.Sweeper{
		Name: "aws_lb_listener",
		F:    sweepListeners,
	})
}

func sweepLoadBalancers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.ELBV2Client(ctx)

	var sweeperErrs *multierror.Error

	pages := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(conn, &elasticloadbalancingv2.DescribeLoadBalancersInput{})

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping LB sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving LBs: %w", err))
		}

		for _, loadBalancer := range page.LoadBalancers {
			name := aws.ToString(loadBalancer.LoadBalancerName)

			log.Printf("[INFO] Deleting LB: %s", name)
			_, err := conn.DeleteLoadBalancer(ctx, &elasticloadbalancingv2.DeleteLoadBalancerInput{
				LoadBalancerArn: loadBalancer.LoadBalancerArn,
			})
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("failed to delete LB (%s): %w", name, err))
				continue
			}
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepTargetGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.ELBV2Client(ctx)

	pages := elasticloadbalancingv2.NewDescribeTargetGroupsPaginator(conn, &elasticloadbalancingv2.DescribeTargetGroupsInput{})

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping LB Target Group sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("retrieving LB Target Groups: %w", err)
		}

		for _, targetGroup := range page.TargetGroups {
			name := aws.ToString(targetGroup.TargetGroupName)

			log.Printf("[INFO] Deleting LB Target Group: %s", name)
			_, err := conn.DeleteTargetGroup(ctx, &elasticloadbalancingv2.DeleteTargetGroupInput{
				TargetGroupArn: targetGroup.TargetGroupArn,
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete LB Target Group (%s): %s", name, err)
			}
		}
	}

	return nil
}

func sweepListeners(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.ELBV2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(conn, &elasticloadbalancingv2.DescribeLoadBalancersInput{})

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing ELBv2 Listeners for %s: %w", region, err))
		}

		for _, loadBalancer := range page.LoadBalancers {
			pages := elasticloadbalancingv2.NewDescribeListenersPaginator(conn, &elasticloadbalancingv2.DescribeListenersInput{
				LoadBalancerArn: loadBalancer.LoadBalancerArn})

			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					errs = multierror.Append(errs, fmt.Errorf("failed to describe LB Listeners (%s): %w", region, err))
					continue
				}

				for _, listener := range page.Listeners {
					r := ResourceListener()
					d := r.Data(nil)
					d.SetId(aws.ToString(listener.ListenerArn))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping ELBv2 Listeners for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping ELBv2 Listener sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
