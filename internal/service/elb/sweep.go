// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
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
	conn := client.ELBClient(ctx)
	input := &elasticloadbalancing.DescribeLoadBalancersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := elasticloadbalancing.NewDescribeLoadBalancersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ELB Classic Load Balancer sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ELB Classic Load Balancers (%s): %w", region, err)
		}

		for _, v := range page.LoadBalancerDescriptions {
			r := resourceLoadBalancer()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.LoadBalancerName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ELB Classic Load Balancers (%s): %w", region, err)
	}

	return nil
}
