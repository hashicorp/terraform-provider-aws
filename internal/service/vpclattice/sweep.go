// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_vpclattice_resource_configuration", sweepResourceConfigurations, "aws_vpclattice_service_network_resource_association")
	awsv2.Register("aws_vpclattice_resource_gateway", sweepResourceGateways, "aws_vpclattice_resource_configuration")
	awsv2.Register("aws_vpclattice_service", sweepServices)
	awsv2.Register("aws_vpclattice_service_network", sweepServiceNetworks, "aws_vpclattice_service")
	awsv2.Register("aws_vpclattice_service_network_resource_association", sweepServiceNetworkResourceAssociations)
	awsv2.Register("aws_vpclattice_target_group", sweepTargetGroups)
}

func sweepResourceConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.VPCLatticeClient(ctx)
	input := &vpclattice.ListResourceConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := vpclattice.NewListResourceConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			if aws.ToBool(v.AmazonManaged) {
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceConfigurationResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id))))
		}
	}

	return sweepResources, nil
}

func sweepResourceGateways(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.VPCLatticeClient(ctx)
	input := &vpclattice.ListResourceGatewaysInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := vpclattice.NewListResourceGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceGatewayResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id))))
		}
	}

	return sweepResources, nil
}

func sweepServices(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.VPCLatticeClient(ctx)
	input := &vpclattice.ListServicesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := vpclattice.NewListServicesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			r := resourceService()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepServiceNetworks(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.VPCLatticeClient(ctx)
	input := &vpclattice.ListServiceNetworksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := vpclattice.NewListServiceNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			r := resourceServiceNetwork()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepServiceNetworkResourceAssociations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.VPCLatticeClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	input := &vpclattice.ListServiceNetworkResourceAssociationsInput{}
	pages := vpclattice.NewListServiceNetworkResourceAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			sweepResources = append(sweepResources, framework.NewSweepResource(newServiceNetworkResourceAssociationResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id))))
		}
	}

	return sweepResources, nil
}

func sweepTargetGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.VPCLatticeClient(ctx)
	input := &vpclattice.ListTargetGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := vpclattice.NewListTargetGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			r := resourceTargetGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
