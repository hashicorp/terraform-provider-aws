//go:build sweep
// +build sweep

package globalaccelerator

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
}

func sweepAccelerators(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GlobalAcceleratorConn()
	input := &globalaccelerator.ListAcceleratorsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListAcceleratorsPagesWithContext(ctx, input, func(page *globalaccelerator.ListAcceleratorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Accelerators {
			r := ResourceAccelerator()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.AcceleratorArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Global Accelerator Accelerator sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Global Accelerator Accelerators (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Global Accelerator Accelerators (%s): %w", region, err)
	}

	return nil
}

func sweepEndpointGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GlobalAcceleratorConn()
	input := &globalaccelerator.ListAcceleratorsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListAcceleratorsPagesWithContext(ctx, input, func(page *globalaccelerator.ListAcceleratorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Accelerators {
			input := &globalaccelerator.ListListenersInput{
				AcceleratorArn: v.AcceleratorArn,
			}

			err := conn.ListListenersPagesWithContext(ctx, input, func(page *globalaccelerator.ListListenersOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Listeners {
					input := &globalaccelerator.ListEndpointGroupsInput{
						ListenerArn: v.ListenerArn,
					}

					err := conn.ListEndpointGroupsPagesWithContext(ctx, input, func(page *globalaccelerator.ListEndpointGroupsOutput, lastPage bool) bool {
						if page == nil {
							return !lastPage
						}

						for _, v := range page.EndpointGroups {
							r := ResourceEndpointGroup()
							d := r.Data(nil)
							d.SetId(aws.StringValue(v.EndpointGroupArn))

							sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
						}

						return !lastPage
					})

					if sweep.SkipSweepError(err) {
						continue
					}

					if err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Global Accelerator Endpoint Groups (%s): %w", region, err))
					}
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Global Accelerator Listeners (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Global Accelerator Endpoint Group sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Global Accelerator Accelerators (%s): %w", region, err))
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Global Accelerator Endpoint Groups (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepListeners(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).GlobalAcceleratorConn()
	input := &globalaccelerator.ListAcceleratorsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListAcceleratorsPagesWithContext(ctx, input, func(page *globalaccelerator.ListAcceleratorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Accelerators {
			input := &globalaccelerator.ListListenersInput{
				AcceleratorArn: v.AcceleratorArn,
			}

			err := conn.ListListenersPagesWithContext(ctx, input, func(page *globalaccelerator.ListListenersOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Listeners {
					r := ResourceListener()
					d := r.Data(nil)
					d.SetId(aws.StringValue(v.ListenerArn))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Global Accelerator Listeners (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Global Accelerator Listener sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Global Accelerator Accelerators (%s): %w", region, err))
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Global Accelerator Listeners (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
