// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
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
			"aws_networkmanager_connect_attachment",
			"aws_networkmanager_dx_gateway_attachment",
			"aws_networkmanager_site_to_site_vpn_attachment",
			"aws_networkmanager_transit_gateway_peering",
			"aws_networkmanager_vpc_attachment",
		},
	})

	resource.AddTestSweepers("aws_networkmanager_connect_attachment", &resource.Sweeper{
		Name: "aws_networkmanager_connect_attachment",
		F:    sweepConnectAttachments,
	})

	resource.AddTestSweepers("aws_networkmanager_dx_gateway_attachment", &resource.Sweeper{
		Name: "aws_networkmanager_dx_gateway_attachment",
		F:    sweepDirectConnectGatewayAttachments,
	})

	resource.AddTestSweepers("aws_networkmanager_site_to_site_vpn_attachment", &resource.Sweeper{
		Name: "aws_networkmanager_site_to_site_vpn_attachment",
		F:    sweepSiteToSiteVPNAttachments,
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
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.DescribeGlobalNetworksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewDescribeGlobalNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager Global Network sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Network Manager Global Networks (%s): %w", region, err)
		}

		for _, v := range page.GlobalNetworks {
			r := resourceGlobalNetwork()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.GlobalNetworkId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager Global Networks (%s): %w", region, err)
	}

	return nil
}

func sweepCoreNetworks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.ListCoreNetworksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewListCoreNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager Core Network sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Network Manager Core Networks (%s): %w", region, err)
		}

		for _, v := range page.CoreNetworks {
			r := resourceCoreNetwork()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.CoreNetworkId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager Core Networks (%s): %w", region, err)
	}

	return nil
}

func sweepConnectAttachments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.ListAttachmentsInput{
		AttachmentType: awstypes.AttachmentTypeConnect,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewListAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager Connect Attachment sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Network Manager Connect Attachments (%s): %w", region, err)
		}

		for _, v := range page.Attachments {
			r := resourceConnectAttachment()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AttachmentId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager Connect Attachments (%s): %w", region, err)
	}

	return nil
}

func sweepDirectConnectGatewayAttachments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.ListAttachmentsInput{
		AttachmentType: awstypes.AttachmentTypeDirectConnectGateway,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewListAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager Direct Connect Gateway Attachment sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Network Manager Direct Connect Gateway Attachments (%s): %w", region, err)
		}

		for _, v := range page.Attachments {
			sweepResources = append(sweepResources, framework.NewSweepResource(newDirectConnectGatewayAttachmentResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.AttachmentId))))

			r := resourceConnectAttachment()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AttachmentId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager Direct Connect Gateway Attachments (%s): %w", region, err)
	}

	return nil
}

func sweepSiteToSiteVPNAttachments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.ListAttachmentsInput{
		AttachmentType: awstypes.AttachmentTypeSiteToSiteVpn,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewListAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager Site To Site VPN Attachment sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Network Manager Site To Site VPN Attachments (%s): %w", region, err)
		}

		for _, v := range page.Attachments {
			r := resourceSiteToSiteVPNAttachment()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AttachmentId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager Site To Site VPN Attachments (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGatewayPeerings(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.ListPeeringsInput{
		PeeringType: awstypes.PeeringTypeTransitGateway,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewListPeeringsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager Transit Gateway Peering sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Network Manager Transit Gateway Peerings (%s): %w", region, err)
		}

		for _, v := range page.Peerings {
			r := resourceTransitGatewayPeering()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.PeeringId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager Transit Gateway Peerings (%s): %w", region, err)
	}

	return nil
}

func sweepTransitGatewayRouteTableAttachments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.ListAttachmentsInput{
		AttachmentType: awstypes.AttachmentTypeTransitGatewayRouteTable,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewListAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager Transit Gateway Route Table Attachment sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Network Manager Transit Gateway Route Table Attachments (%s): %w", region, err)
		}

		for _, v := range page.Attachments {
			r := resourceTransitGatewayRouteTableAttachment()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AttachmentId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager Transit Gateway Route Table Attachments (%s): %w", region, err)
	}

	return nil
}

func sweepVPCAttachments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.ListAttachmentsInput{
		AttachmentType: awstypes.AttachmentTypeVpc,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewListAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager VPC Attachment sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Network Manager VPC Attachments (%s): %w", region, err)
		}

		for _, v := range page.Attachments {
			r := resourceVPCAttachment()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AttachmentId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Manager VPC Attachments (%s): %w", region, err)
	}

	return nil
}

func sweepSites(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.DescribeGlobalNetworksInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewDescribeGlobalNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager Site sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Global Networks (%s): %w", region, err))
		}

		for _, v := range page.GlobalNetworks {
			input := &networkmanager.GetSitesInput{
				GlobalNetworkId: v.GlobalNetworkId,
			}

			pages := networkmanager.NewGetSitesPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Sites (%s): %w", region, err))
				}

				for _, v := range page.Sites {
					r := resourceSite()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.SiteId))
					d.Set("global_network_id", v.GlobalNetworkId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Network Manager Sites (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepDevices(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.DescribeGlobalNetworksInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewDescribeGlobalNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager Device sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Global Networks (%s): %w", region, err))
		}

		for _, v := range page.GlobalNetworks {
			input := &networkmanager.GetDevicesInput{
				GlobalNetworkId: v.GlobalNetworkId,
			}

			pages := networkmanager.NewGetDevicesPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Devices (%s): %w", region, err))
				}

				for _, v := range page.Devices {
					r := resourceDevice()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.DeviceId))
					d.Set("global_network_id", v.GlobalNetworkId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Network Manager Devices (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepLinks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.DescribeGlobalNetworksInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewDescribeGlobalNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager Link sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Global Networks (%s): %w", region, err))
		}

		for _, v := range page.GlobalNetworks {
			input := &networkmanager.GetLinksInput{
				GlobalNetworkId: v.GlobalNetworkId,
			}

			pages := networkmanager.NewGetLinksPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Links (%s): %w", region, err))
				}

				for _, v := range page.Links {
					r := resourceLink()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.LinkId))
					d.Set("global_network_id", v.GlobalNetworkId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Network Manager Links (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepLinkAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.DescribeGlobalNetworksInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewDescribeGlobalNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager Link Association sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Global Networks (%s): %w", region, err))
		}

		for _, v := range page.GlobalNetworks {
			input := &networkmanager.GetLinkAssociationsInput{
				GlobalNetworkId: v.GlobalNetworkId,
			}

			pages := networkmanager.NewGetLinkAssociationsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Link Associations (%s): %w", region, err))
				}

				for _, v := range page.LinkAssociations {
					r := resourceLinkAssociation()
					d := r.Data(nil)
					d.SetId(linkAssociationCreateResourceID(aws.ToString(v.GlobalNetworkId), aws.ToString(v.LinkId), aws.ToString(v.DeviceId)))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Network Manager Link Associations (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepConnections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NetworkManagerClient(ctx)
	input := &networkmanager.DescribeGlobalNetworksInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkmanager.NewDescribeGlobalNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Manager Connection sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Global Networks (%s): %w", region, err))
		}

		for _, v := range page.GlobalNetworks {
			input := &networkmanager.GetConnectionsInput{
				GlobalNetworkId: v.GlobalNetworkId,
			}

			pages := networkmanager.NewGetConnectionsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Network Manager Connections (%s): %w", region, err))
				}

				for _, v := range page.Connections {
					r := resourceConnection()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.ConnectionId))
					d.Set("global_network_id", v.GlobalNetworkId)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Network Manager Connections (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
