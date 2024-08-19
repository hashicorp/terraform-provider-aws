// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
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
	input := &elasticloadbalancingv2.DescribeLoadBalancersInput{}
	conn := client.ELBV2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ELBv2 Load Balancer sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ELBv2 Load Balancers (%s): %w", region, err)
		}

		for _, v := range page.LoadBalancers {
			r := resourceLoadBalancer()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.LoadBalancerArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ELBv2 Load Balancers (%s): %w", region, err)
	}

	return nil
}

func sweepTargetGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	input := &elasticloadbalancingv2.DescribeTargetGroupsInput{}
	conn := client.ELBV2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := elasticloadbalancingv2.NewDescribeTargetGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ELBv2 Target Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ELBv2 Target Groups (%s): %w", region, err)
		}

		for _, v := range page.TargetGroups {
			r := resourceTargetGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TargetGroupArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ELBv2 Target Groups (%s): %w", region, err)
	}

	return nil
}

func sweepListeners(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	input := &elasticloadbalancingv2.DescribeLoadBalancersInput{}
	conn := client.ELBV2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ELBv2 Listener sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ELBv2 Load Balancers (%s): %w", region, err)
		}

		for _, v := range page.LoadBalancers {
			input := &elasticloadbalancingv2.DescribeListenersInput{
				LoadBalancerArn: v.LoadBalancerArn,
			}

			pages := elasticloadbalancingv2.NewDescribeListenersPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.Listeners {
					r := resourceListener()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.ListenerArn))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ELBv2 Listeners (%s): %w", region, err)
	}

	return nil
}
