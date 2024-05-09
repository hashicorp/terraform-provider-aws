// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_globalaccelerator_accelerator", &resource.Sweeper{
		Name: "aws_globalaccelerator_accelerator",
		F:    sweepAccelerators,
		Dependencies: []string{
			"aws_globalaccelerator_listener",
		},
	})

	resource.AddTestSweepers("aws_globalaccelerator_listener", &resource.Sweeper{
		Name: "aws_globalaccelerator_listener",
		F:    sweepListeners,
		Dependencies: []string{
			"aws_globalaccelerator_endpoint_group",
		},
	})

	resource.AddTestSweepers("aws_globalaccelerator_endpoint_group", &resource.Sweeper{
		Name: "aws_globalaccelerator_endpoint_group",
		F:    sweepEndpointGroups,
	})

	resource.AddTestSweepers("aws_globalaccelerator_custom_routing_accelerator", &resource.Sweeper{
		Name: "aws_globalaccelerator_custom_routing_accelerator",
		F:    sweepCustomRoutingAccelerators,
		Dependencies: []string{
			"aws_globalaccelerator_custom_routing_listener",
		},
	})

	resource.AddTestSweepers("aws_globalaccelerator_custom_routing_listener", &resource.Sweeper{
		Name: "aws_globalaccelerator_custom_routing_listener",
		F:    sweepCustomRoutingListeners,
		Dependencies: []string{
			"aws_globalaccelerator_custom_routing_endpoint_group",
		},
	})

	resource.AddTestSweepers("aws_globalaccelerator_custom_routing_endpoint_group", &resource.Sweeper{
		Name: "aws_globalaccelerator_custom_routing_endpoint_group",
		F:    sweepCustomRoutingEndpointGroups,
	})
}

func sweepAccelerators(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlobalAcceleratorClient(ctx)
	input := &globalaccelerator.ListAcceleratorsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := globalaccelerator.NewListAcceleratorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Global Accelerator Accelerator sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Global Accelerator Accelerators (%s): %w", region, err)
		}

		for _, v := range page.Accelerators {
			r := resourceAccelerator()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AcceleratorArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Global Accelerator Accelerators (%s): %w", region, err)
	}

	return nil
}

func sweepEndpointGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlobalAcceleratorClient(ctx)
	input := &globalaccelerator.ListAcceleratorsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := globalaccelerator.NewListAcceleratorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Global Accelerator Endpoint Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Global Accelerator Accelerators (%s): %w", region, err)
		}

		for _, v := range page.Accelerators {
			input := &globalaccelerator.ListListenersInput{
				AcceleratorArn: v.AcceleratorArn,
			}

			pages := globalaccelerator.NewListListenersPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.Listeners {
					input := &globalaccelerator.ListEndpointGroupsInput{
						ListenerArn: v.ListenerArn,
					}

					pages := globalaccelerator.NewListEndpointGroupsPaginator(conn, input)
					for pages.HasMorePages() {
						page, err := pages.NextPage(ctx)

						if err != nil {
							continue
						}

						for _, v := range page.EndpointGroups {
							r := resourceEndpointGroup()
							d := r.Data(nil)
							d.SetId(aws.ToString(v.EndpointGroupArn))

							sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
						}
					}
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Global Accelerator Endpoint Groups (%s): %w", region, err)
	}

	return nil
}

func sweepListeners(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlobalAcceleratorClient(ctx)
	input := &globalaccelerator.ListAcceleratorsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := globalaccelerator.NewListAcceleratorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Global Accelerator Endpoint Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Global Accelerator Accelerators (%s): %w", region, err)
		}

		for _, v := range page.Accelerators {
			input := &globalaccelerator.ListListenersInput{
				AcceleratorArn: v.AcceleratorArn,
			}

			pages := globalaccelerator.NewListListenersPaginator(conn, input)
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
		return fmt.Errorf("error sweeping Global Accelerator Listeners (%s): %w", region, err)
	}

	return nil
}

func sweepCustomRoutingAccelerators(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlobalAcceleratorClient(ctx)
	input := &globalaccelerator.ListCustomRoutingAcceleratorsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := globalaccelerator.NewListCustomRoutingAcceleratorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Global Accelerator Custom Routing Accelerator sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Global Accelerator Custom Routing Accelerators (%s): %w", region, err)
		}

		for _, v := range page.Accelerators {
			r := resourceCustomRoutingAccelerator()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AcceleratorArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Global Accelerator Custom Routing Accelerators (%s): %w", region, err)
	}

	return nil
}

func sweepCustomRoutingEndpointGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlobalAcceleratorClient(ctx)
	input := &globalaccelerator.ListCustomRoutingAcceleratorsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := globalaccelerator.NewListCustomRoutingAcceleratorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Global Accelerator Custom Routing Accelerator sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Global Accelerator Custom Routing Accelerators (%s): %w", region, err)
		}

		for _, v := range page.Accelerators {
			input := &globalaccelerator.ListCustomRoutingListenersInput{
				AcceleratorArn: v.AcceleratorArn,
			}

			pages := globalaccelerator.NewListCustomRoutingListenersPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.Listeners {
					input := &globalaccelerator.ListCustomRoutingEndpointGroupsInput{
						ListenerArn: v.ListenerArn,
					}

					pages := globalaccelerator.NewListCustomRoutingEndpointGroupsPaginator(conn, input)
					for pages.HasMorePages() {
						page, err := pages.NextPage(ctx)

						if err != nil {
							continue
						}

						for _, v := range page.EndpointGroups {
							r := resourceCustomRoutingEndpointGroup()
							d := r.Data(nil)
							d.SetId(aws.ToString(v.EndpointGroupArn))

							sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
						}
					}
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Global Accelerator Custom Routing Endpoint Groups (%s): %w", region, err)
	}

	return nil
}

func sweepCustomRoutingListeners(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GlobalAcceleratorClient(ctx)
	input := &globalaccelerator.ListCustomRoutingAcceleratorsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := globalaccelerator.NewListCustomRoutingAcceleratorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Global Accelerator Custom Routing Accelerator sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Global Accelerator Custom Routing Accelerators (%s): %w", region, err)
		}

		for _, v := range page.Accelerators {
			input := &globalaccelerator.ListCustomRoutingListenersInput{
				AcceleratorArn: v.AcceleratorArn,
			}

			pages := globalaccelerator.NewListCustomRoutingListenersPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.Listeners {
					r := resourceCustomRoutingListener()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.ListenerArn))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Global Accelerator Custom Routing Listeners (%s): %w", region, err)
	}

	return nil
}
