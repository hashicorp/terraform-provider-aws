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
			"aws_networkmanager_core_network",
			"aws_networkmanager_site",
		},
	})

	resource.AddTestSweepers("aws_networkmanager_core_network", &resource.Sweeper{
		Name: "aws_networkmanager_core_network",
		F:    sweepCoreNetworks,
		Dependencies: []string{
			"aws_networkmanager_transit_gateway_peering",
			"aws_networkmanager_vpc_attachment",
		},
	})

	resource.AddTestSweepers("aws_networkmanager_transit_gateway_peering", &resource.Sweeper{
		Name: "aws_networkmanager_transit_gateway_peering",
		F:    sweepTransitGatewayPeerings,
		Dependencies: []string{
			"aws_networkmanager_transit_gateway_route_table_attachment",
		},
	})

	resource.AddTestSweepers("aws_networkmanager_transit_gateway_route_table_attachment", &resource.Sweeper{
		Name: "aws_networkmanager_transit_gateway_route_table_attachment",
		F:    sweepTransitGatewayRouteTableAttachments,
	})

	resource.AddTestSweepers("aws_networkmanager_vpc_attachment", &resource.Sweeper{
		Name: "aws_networkmanager_vpc_attachment",
		F:    sweepVPCAttachments,
	})

	resource.AddTestSweepers("aws_networkmanager_site", &resource.Sweeper{
		Name: "aws_networkmanager_site",
		F:    sweepSites,
		Dependencies: []string{
			"aws_networkmanager_device",
			"aws_networkmanager_link",
		},
	})

	resource.AddTestSweepers("aws_networkmanager_device", &resource.Sweeper{
		Name: "aws_networkmanager_device",
		F:    sweepDevices,
		Dependencies: []string{
			"aws_networkmanager_link_association",
		},
	})

	resource.AddTestSweepers("aws_networkmanager_link", &resource.Sweeper{
		Name: "aws_networkmanager_link",
		F:    sweepLinks,
		Dependencies: []string{
			"aws_networkmanager_link_association",
		},
	})

	resource.AddTestSweepers("aws_networkmanager_link_association", &resource.Sweeper{
		Name: "aws_networkmanager_link_association",
		F:    sweepLinkAssociations,
		Dependencies: []string{
			"aws_networkmanager_connection",
		},
	})

	resource.AddTestSweepers("aws_networkmanager_connection", &resource.Sweeper{
		Name: "aws_networkmanager_connection",
		F:    sweepConnections,
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

func sweepCoreNetworks(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).NetworkManagerConn
	input := &networkmanager.ListCoreNetworksInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListCoreNetworksPages(input, func(page *networkmanager.ListCoreNetworksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CoreNetworks {
			r := ResourceCoreNetwork()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.CoreNetworkId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Core Network sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Network Manager Core Networks (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager Core Networks (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGatewayPeerings(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).NetworkManagerConn
	input := &networkmanager.ListPeeringsInput{
		PeeringType: aws.String(networkmanager.PeeringTypeTransitGateway),
	}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListPeeringsPages(input, func(page *networkmanager.ListPeeringsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Peerings {
			r := ResourceTransitGatewayPeering()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.PeeringId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Transit Gateway Peering sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Network Manager Transit Gateway Peerings (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager Transit Gateway Peerings (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGatewayRouteTableAttachments(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).NetworkManagerConn
	input := &networkmanager.ListAttachmentsInput{
		AttachmentType: aws.String(networkmanager.AttachmentTypeTransitGatewayRouteTable),
	}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListAttachmentsPages(input, func(page *networkmanager.ListAttachmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Attachments {
			r := ResourceTransitGatewayRouteTableAttachment()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.AttachmentId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Transit Gateway Route Table Attachment sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Network Manager Transit Gateway Route Table Attachments (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager Transit Gateway Route Table Attachments (%s): %w", region, err)
	}

	return nil
}

func sweepVPCAttachments(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).NetworkManagerConn
	input := &networkmanager.ListAttachmentsInput{
		AttachmentType: aws.String(networkmanager.AttachmentTypeVpc),
	}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListAttachmentsPages(input, func(page *networkmanager.ListAttachmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Attachments {
			r := ResourceVPCAttachment()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.AttachmentId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager VPC Attachment sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Network Manager VPC Attachments (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager VPC Attachments (%s): %w", region, err)
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

func sweepDevices(region string) error {
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
			input := &networkmanager.GetDevicesInput{
				GlobalNetworkId: v.GlobalNetworkId,
			}

			err := conn.GetDevicesPages(input, func(page *networkmanager.GetDevicesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Devices {
					r := ResourceDevice()
					d := r.Data(nil)
					d.SetId(aws.StringValue(v.DeviceId))
					d.Set("global_network_id", v.GlobalNetworkId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Devices (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Device sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Global Networks (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Network Manager Devices (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepLinks(region string) error {
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
			input := &networkmanager.GetLinksInput{
				GlobalNetworkId: v.GlobalNetworkId,
			}

			err := conn.GetLinksPages(input, func(page *networkmanager.GetLinksOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Links {
					r := ResourceLink()
					d := r.Data(nil)
					d.SetId(aws.StringValue(v.LinkId))
					d.Set("global_network_id", v.GlobalNetworkId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Links (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Link sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Global Networks (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Network Manager Links (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepLinkAssociations(region string) error {
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
			input := &networkmanager.GetLinkAssociationsInput{
				GlobalNetworkId: v.GlobalNetworkId,
			}

			err := conn.GetLinkAssociationsPages(input, func(page *networkmanager.GetLinkAssociationsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.LinkAssociations {
					r := ResourceLinkAssociation()
					d := r.Data(nil)
					d.SetId(LinkAssociationCreateResourceID(aws.StringValue(v.GlobalNetworkId), aws.StringValue(v.LinkId), aws.StringValue(v.DeviceId)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Link Associations (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Link Association sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Global Networks (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Network Manager Link Associations (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepConnections(region string) error {
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
			input := &networkmanager.GetConnectionsInput{
				GlobalNetworkId: v.GlobalNetworkId,
			}

			err := conn.GetConnectionsPages(input, func(page *networkmanager.GetConnectionsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Connections {
					r := ResourceConnection()
					d := r.Data(nil)
					d.SetId(aws.StringValue(v.ConnectionId))
					d.Set("global_network_id", v.GlobalNetworkId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Connections (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Connection sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Global Networks (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Network Manager Connections (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
