// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package elb

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_elb", &resource.Sweeper{
		Name: "aws_elb",
		F:    sweepLoadBalancers,
	})
}

func sweepLoadBalancers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ELBConn(ctx)
	input := &elb.DescribeLoadBalancersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeLoadBalancersPagesWithContext(ctx, input, func(page *elb.DescribeLoadBalancersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LoadBalancerDescriptions {
			r := ResourceLoadBalancer()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.LoadBalancerName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ELB Classic Load Balancer sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing ELB Classic Load Balancers (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ELB Classic Load Balancers (%s): %w", region, err)
	}

	return nil
}
