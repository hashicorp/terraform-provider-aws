//go:build sweep
// +build sweep

package networkmanager

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_networkmanager_global_network", &resource.Sweeper{
		Name: "aws_networkmanager_global_network",
		F:    sweepGlobalNetworks,
		Dependencies: []string{
			"aws_networkmanager_site",
		},
	})

	resource.AddTestSweepers("aws_networkmanager_site", &resource.Sweeper{
		Name: "aws_networkmanager_site",
		F:    sweepSites,
	})
}

func sweepGlobalNetworks(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).NetworkManagerConn
	input := &networkmanager.DescribeGlobalNetworksInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeGlobalNetworksPages(input, func(page *networkmanager.DescribeGlobalNetworksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.GlobalNetworks {
			r := ResourceGlobalNetwork()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.GlobalNetworkId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Global Network sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Network Manager Global Networks (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager Global Networks (%s): %w", region, err)
	}

	return nil
}

func sweepSites(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).NetworkManagerConn
	input := &networkmanager.DescribeGlobalNetworksInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeGlobalNetworksPages(input, func(page *networkmanager.DescribeGlobalNetworksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.GlobalNetworks {
			input := &networkmanager.GetSitesInput{
				GlobalNetworkId: v.GlobalNetworkId,
			}

			err := conn.GetSitesPages(input, func(page *networkmanager.GetSitesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Sites {
					r := ResourceSite()
					d := r.Data(nil)
					d.SetId(aws.StringValue(v.SiteId))
					d.Set("global_network_id", v.GlobalNetworkId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Sites (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Site sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Global Networks (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Network Manager Sites (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
