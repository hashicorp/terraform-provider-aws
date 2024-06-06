// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func findAvailabilityZonesV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeAvailabilityZonesInput) ([]awstypes.AvailabilityZone, error) {
	output, err := conn.DescribeAvailabilityZones(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AvailabilityZones, nil
}

func findVPCAttributeV2(ctx context.Context, conn *ec2.Client, vpcID string, attribute awstypes.VpcAttributeName) (bool, error) {
	input := &ec2.DescribeVpcAttributeInput{
		Attribute: attribute,
		VpcId:     aws.String(vpcID),
	}

	output, err := conn.DescribeVpcAttribute(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
		return false, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return false, err
	}

	if output == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	var v *awstypes.AttributeBooleanValue
	switch attribute {
	case awstypes.VpcAttributeNameEnableDnsHostnames:
		v = output.EnableDnsHostnames
	case awstypes.VpcAttributeNameEnableDnsSupport:
		v = output.EnableDnsSupport
	case awstypes.VpcAttributeNameEnableNetworkAddressUsageMetrics:
		v = output.EnableNetworkAddressUsageMetrics
	default:
		return false, fmt.Errorf("unsupported VPC attribute: %s", attribute)
	}

	if v == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	return aws.ToBool(v.Value), nil
}

func findVPCV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcsInput) (*awstypes.Vpc, error) {
	output, err := findVPCsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcsInput) ([]awstypes.Vpc, error) {
	var output []awstypes.Vpc

	pages := ec2.NewDescribeVpcsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Vpcs...)
	}

	return output, nil
}

func findVPCByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		VpcIds: []string{id},
	}

	return findVPCV2(ctx, conn, input)
}

func findVPCIPv6CIDRBlockAssociationByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpcIpv6CidrBlockAssociation, *awstypes.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"ipv6-cidr-block-association.association-id": id,
		}),
	}

	vpc, err := findVPCV2(ctx, conn, input)

	if err != nil {
		return nil, nil, err
	}

	for _, association := range vpc.Ipv6CidrBlockAssociationSet {
		if aws.ToString(association.AssociationId) == id {
			if state := association.Ipv6CidrBlockState.State; state == awstypes.VpcCidrBlockStateCodeDisassociated {
				return nil, nil, &retry.NotFoundError{Message: string(state)}
			}

			return &association, vpc, nil
		}
	}

	return nil, nil, &retry.NotFoundError{}
}

func findVPCDefaultNetworkACLV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"default": "true",
			"vpc-id":  id,
		}),
	}

	return findNetworkACLV2(ctx, conn, input)
}

func findNetworkACLByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		NetworkAclIds: []string{id},
	}

	output, err := findNetworkACLV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.NetworkAclId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findNetworkACLV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkAclsInput) (*awstypes.NetworkAcl, error) {
	output, err := findNetworkACLsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findNetworkACLsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkAclsInput) ([]awstypes.NetworkAcl, error) {
	var output []awstypes.NetworkAcl

	pages := ec2.NewDescribeNetworkAclsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkACLIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.NetworkAcls...)
	}

	return output, nil
}

func findVPCDefaultSecurityGroupV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"group-name": DefaultSecurityGroupName,
			"vpc-id":     id,
		}),
	}

	return findSecurityGroupV2(ctx, conn, input)
}

func findVPCMainRouteTable(ctx context.Context, conn *ec2.Client, id string) (*awstypes.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"association.main": "true",
			"vpc-id":           id,
		}),
	}

	return findRouteTable(ctx, conn, input)
}

func findRouteTable(ctx context.Context, conn *ec2.Client, input *ec2.DescribeRouteTablesInput) (*awstypes.RouteTable, error) {
	output, err := findRouteTables(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findRouteTables(ctx context.Context, conn *ec2.Client, input *ec2.DescribeRouteTablesInput) ([]awstypes.RouteTable, error) {
	var output []awstypes.RouteTable

	pages := ec2.NewDescribeRouteTablesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.RouteTables...)
	}

	return output, nil
}

func findSecurityGroupV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupsInput) (*awstypes.SecurityGroup, error) {
	output, err := findSecurityGroupsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSecurityGroupsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupsInput) ([]awstypes.SecurityGroup, error) {
	var output []awstypes.SecurityGroup

	pages := ec2.NewDescribeSecurityGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound, errCodeInvalidSecurityGroupIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.SecurityGroups...)
	}

	return output, nil
}

// FindSecurityGroupByNameAndVPCIDV2 looks up a security group by name, VPC ID. Returns a retry.NotFoundError if not found.
func FindSecurityGroupByNameAndVPCIDV2(ctx context.Context, conn *ec2.Client, name, vpcID string) (*awstypes.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterListV2(
			map[string]string{
				"group-name": name,
				"vpc-id":     vpcID,
			},
		),
	}
	return findSecurityGroupV2(ctx, conn, input)
}

func findIPAMPoolAllocationsV2(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolAllocationsInput) ([]awstypes.IpamPoolAllocation, error) {
	var output []awstypes.IpamPoolAllocation

	pages := ec2.NewGetIpamPoolAllocationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolAllocationIdNotFound, errCodeInvalidIPAMPoolIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpamPoolAllocations...)
	}

	return output, nil
}

func findNetworkInterfacesV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkInterfacesInput) ([]awstypes.NetworkInterface, error) {
	var output []awstypes.NetworkInterface

	pages := ec2.NewDescribeNetworkInterfacesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInterfaceIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.NetworkInterfaces...)
	}

	return output, nil
}

func findNetworkInterfaceV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkInterfacesInput) (*awstypes.NetworkInterface, error) {
	output, err := findNetworkInterfacesV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findNetworkInterfaceByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{id},
	}

	output, err := findNetworkInterfaceV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.NetworkInterfaceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, err
}

func findNetworkInterfaceAttachmentByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkInterfaceAttachment, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"attachment.attachment-id": id,
		}),
	}

	networkInterface, err := findNetworkInterfaceV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if networkInterface.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return networkInterface.Attachment, nil
}

/*
	func findNetworkInterfaceByAttachmentIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkInterface, error) {
		input := &ec2.DescribeNetworkInterfacesInput{
			Filters: newAttributeFilterListV2(map[string]string{
				"attachment.attachment-id": id,
			}),
		}

		networkInterface, err := findNetworkInterfaceV2(ctx, conn, input)

		if err != nil {
			return nil, err
		}

		if networkInterface == nil {
			return nil, tfresource.NewEmptyResultError(input)
		}

		return networkInterface, nil
	}
*/

func findNetworkInterfacesByAttachmentInstanceOwnerIDAndDescriptionV2(ctx context.Context, conn *ec2.Client, attachmentInstanceOwnerID, description string) ([]awstypes.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"attachment.instance-owner-id": attachmentInstanceOwnerID,
			names.AttrDescription:          description,
		}),
	}

	return findNetworkInterfacesV2(ctx, conn, input)
}

func findEBSVolumesV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVolumesInput) ([]awstypes.Volume, error) {
	var output []awstypes.Volume

	pages := ec2.NewDescribeVolumesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidVolumeNotFound) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.Volumes...)
	}

	return output, nil
}

func FindEBSVolumeV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVolumesInput) (*awstypes.Volume, error) {
	output, err := findEBSVolumesV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPrefixListV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribePrefixListsInput) (*awstypes.PrefixList, error) {
	output, err := findPrefixListsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPrefixListsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribePrefixListsInput) ([]awstypes.PrefixList, error) {
	var output []awstypes.PrefixList

	pages := ec2.NewDescribePrefixListsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidPrefixListIdNotFound) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.PrefixLists...)
	}

	return output, nil
}

func findVPCEndpointByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpcEndpoint, error) {
	input := &ec2.DescribeVpcEndpointsInput{
		VpcEndpointIds: []string{id},
	}

	output, err := findVPCEndpointV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if output.State == awstypes.StateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(output.State),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.VpcEndpointId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPCEndpointV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointsInput) (*awstypes.VpcEndpoint, error) {
	output, err := findVPCEndpointsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCEndpointsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointsInput) ([]awstypes.VpcEndpoint, error) {
	var output []awstypes.VpcEndpoint

	pages := ec2.NewDescribeVpcEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointIdNotFound) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.VpcEndpoints...)
	}

	return output, nil
}

func findPrefixListByNameV2(ctx context.Context, conn *ec2.Client, name string) (*awstypes.PrefixList, error) {
	input := &ec2.DescribePrefixListsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"prefix-list-name": name,
		}),
	}

	return findPrefixListV2(ctx, conn, input)
}

func findVPCEndpointServiceConfigurationByServiceNameV2(ctx context.Context, conn *ec2.Client, name string) (*awstypes.ServiceConfiguration, error) {
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"service-name": name,
		}),
	}

	return findVPCEndpointServiceConfigurationV2(ctx, conn, input)
}

func findVPCEndpointServiceConfigurationV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointServiceConfigurationsInput) (*awstypes.ServiceConfiguration, error) {
	output, err := findVPCEndpointServiceConfigurationsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCEndpointServiceConfigurationsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointServiceConfigurationsInput) ([]awstypes.ServiceConfiguration, error) {
	var output []awstypes.ServiceConfiguration

	pages := ec2.NewDescribeVpcEndpointServiceConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointServiceIdNotFound) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.ServiceConfigurations...)
	}

	return output, nil
}

// findRouteTableByID returns the route table corresponding to the specified identifier.
// Returns NotFoundError if no route table is found.
func findRouteTableByID(ctx context.Context, conn *ec2.Client, routeTableID string) (*awstypes.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		RouteTableIds: []string{routeTableID},
	}

	return findRouteTable(ctx, conn, input)
}

// routeFinder returns the route corresponding to the specified destination.
// Returns NotFoundError if no route is found.
type routeFinder func(context.Context, *ec2.Client, string, string) (*awstypes.Route, error)

// findRouteByIPv4Destination returns the route corresponding to the specified IPv4 destination.
// Returns NotFoundError if no route is found.
func findRouteByIPv4Destination(ctx context.Context, conn *ec2.Client, routeTableID, destinationCidr string) (*awstypes.Route, error) {
	routeTable, err := findRouteTableByID(ctx, conn, routeTableID)

	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if types.CIDRBlocksEqual(aws.ToString(route.DestinationCidrBlock), destinationCidr) {
			return &route, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with IPv4 destination (%s) not found", routeTableID, destinationCidr),
	}
}

// findRouteByIPv6Destination returns the route corresponding to the specified IPv6 destination.
// Returns NotFoundError if no route is found.
func findRouteByIPv6Destination(ctx context.Context, conn *ec2.Client, routeTableID, destinationIpv6Cidr string) (*awstypes.Route, error) {
	routeTable, err := findRouteTableByID(ctx, conn, routeTableID)

	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if types.CIDRBlocksEqual(aws.ToString(route.DestinationIpv6CidrBlock), destinationIpv6Cidr) {
			return &route, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with IPv6 destination (%s) not found", routeTableID, destinationIpv6Cidr),
	}
}

// findRouteByPrefixListIDDestination returns the route corresponding to the specified prefix list destination.
// Returns NotFoundError if no route is found.
func findRouteByPrefixListIDDestination(ctx context.Context, conn *ec2.Client, routeTableID, prefixListID string) (*awstypes.Route, error) {
	routeTable, err := findRouteTableByID(ctx, conn, routeTableID)
	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if aws.ToString(route.DestinationPrefixListId) == prefixListID {
			return &route, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with Prefix List ID destination (%s) not found", routeTableID, prefixListID),
	}
}

// findMainRouteTableAssociationByID returns the main route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func findMainRouteTableAssociationByID(ctx context.Context, conn *ec2.Client, associationID string) (*awstypes.RouteTableAssociation, error) {
	association, err := findRouteTableAssociationByID(ctx, conn, associationID)

	if err != nil {
		return nil, err
	}

	if !aws.ToBool(association.Main) {
		return nil, &retry.NotFoundError{
			Message: fmt.Sprintf("%s is not the association with the main route table", associationID),
		}
	}

	return association, err
}

// findMainRouteTableAssociationByVPCID returns the main route table association for the specified VPC.
// Returns NotFoundError if no route table association is found.
func findMainRouteTableAssociationByVPCID(ctx context.Context, conn *ec2.Client, vpcID string) (*awstypes.RouteTableAssociation, error) {
	routeTable, err := findMainRouteTableByVPCID(ctx, conn, vpcID)

	if err != nil {
		return nil, err
	}

	for _, association := range routeTable.Associations {
		if aws.ToBool(association.Main) {
			if association.AssociationState != nil {
				if state := association.AssociationState.State; state == awstypes.RouteTableAssociationStateCodeDisassociated {
					continue
				}
			}

			return &association, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

// findRouteTableAssociationByID returns the route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func findRouteTableAssociationByID(ctx context.Context, conn *ec2.Client, associationID string) (*awstypes.RouteTableAssociation, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"association.route-table-association-id": associationID,
		}),
	}

	routeTable, err := findRouteTable(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, association := range routeTable.Associations {
		if aws.ToString(association.RouteTableAssociationId) == associationID {
			if association.AssociationState != nil {
				if state := association.AssociationState.State; state == awstypes.RouteTableAssociationStateCodeDisassociated {
					return nil, &retry.NotFoundError{Message: string(state)}
				}
			}

			return &association, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

// findMainRouteTableByVPCID returns the main route table for the specified VPC.
// Returns NotFoundError if no route table is found.
func findMainRouteTableByVPCID(ctx context.Context, conn *ec2.Client, vpcID string) (*awstypes.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"association.main": "true",
			"vpc-id":           vpcID,
		}),
	}

	return findRouteTable(ctx, conn, input)
}

// findVPNGatewayRoutePropagationExists returns NotFoundError if no route propagation for the specified VPN gateway is found.
func findVPNGatewayRoutePropagationExists(ctx context.Context, conn *ec2.Client, routeTableID, gatewayID string) error {
	routeTable, err := findRouteTableByID(ctx, conn, routeTableID)

	if err != nil {
		return err
	}

	for _, v := range routeTable.PropagatingVgws {
		if aws.ToString(v.GatewayId) == gatewayID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("Route Table (%s) VPN Gateway (%s) route propagation not found", routeTableID, gatewayID),
	}
}

func findVPCEndpointServiceConfigurationByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ServiceConfiguration, error) {
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		ServiceIds: []string{id},
	}

	output, err := findVPCEndpointServiceConfigurationV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.ServiceState; state == awstypes.ServiceStateDeleted || state == awstypes.ServiceStateFailed {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.ServiceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPCEndpointServicePrivateDNSNameConfigurationByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.PrivateDnsNameConfiguration, error) {
	out, err := findVPCEndpointServiceConfigurationByIDV2(ctx, conn, id)
	if err != nil {
		return nil, err
	}

	return out.PrivateDnsNameConfiguration, nil
}

func findVPCEndpointServicePermissionsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointServicePermissionsInput) ([]awstypes.AllowedPrincipal, error) {
	var output []awstypes.AllowedPrincipal

	pages := ec2.NewDescribeVpcEndpointServicePermissionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointServiceIdNotFound) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.AllowedPrincipals...)
	}

	return output, nil
}

func findVPCEndpointServicePermissionsByServiceIDV2(ctx context.Context, conn *ec2.Client, id string) ([]awstypes.AllowedPrincipal, error) {
	input := &ec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(id),
	}

	return findVPCEndpointServicePermissionsV2(ctx, conn, input)
}

func findVPCEndpointServicesV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointServicesInput) ([]awstypes.ServiceDetail, []string, error) {
	var serviceDetails []awstypes.ServiceDetail
	var serviceNames []string

	err := describeVPCEndpointServicesPagesV2(ctx, conn, input, func(page *ec2.DescribeVpcEndpointServicesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		serviceDetails = append(serviceDetails, page.ServiceDetails...)
		serviceNames = append(serviceNames, page.ServiceNames...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidServiceName) {
		return nil, nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, nil, err
	}

	return serviceDetails, serviceNames, nil
}

// findVPCEndpointRouteTableAssociationExistsV2 returns NotFoundError if no association for the specified VPC endpoint and route table IDs is found.
func findVPCEndpointRouteTableAssociationExistsV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, routeTableID string) error {
	vpcEndpoint, err := findVPCEndpointByIDV2(ctx, conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointRouteTableID := range vpcEndpoint.RouteTableIds {
		if vpcEndpointRouteTableID == routeTableID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Route Table (%s) Association not found", vpcEndpointID, routeTableID),
	}
}

// findVPCEndpointSecurityGroupAssociationExistsV2 returns NotFoundError if no association for the specified VPC endpoint and security group IDs is found.
func findVPCEndpointSecurityGroupAssociationExistsV2(ctx context.Context, conn *ec2.Client, vpcEndpointID, securityGroupID string) error {
	vpcEndpoint, err := findVPCEndpointByIDV2(ctx, conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, group := range vpcEndpoint.Groups {
		if aws.ToString(group.GroupId) == securityGroupID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Security Group (%s) Association not found", vpcEndpointID, securityGroupID),
	}
}

// findVPCEndpointSubnetAssociationExistsV2 returns NotFoundError if no association for the specified VPC endpoint and subnet IDs is found.
func findVPCEndpointSubnetAssociationExistsV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, subnetID string) error {
	vpcEndpoint, err := findVPCEndpointByIDV2(ctx, conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointSubnetID := range vpcEndpoint.SubnetIds {
		if vpcEndpointSubnetID == subnetID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Subnet (%s) Association not found", vpcEndpointID, subnetID),
	}
}

func findVPCEndpointConnectionByServiceIDAndVPCEndpointIDV2(ctx context.Context, conn *ec2.Client, serviceID, vpcEndpointID string) (*awstypes.VpcEndpointConnection, error) {
	input := &ec2.DescribeVpcEndpointConnectionsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"service-id": serviceID,
			// "InvalidFilter: The filter vpc-endpoint-id  is invalid"
			// "vpc-endpoint-id ": vpcEndpointID,
		}),
	}

	var output *awstypes.VpcEndpointConnection

	pages := ec2.NewDescribeVpcEndpointConnectionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.VpcEndpointConnections {
			v := v
			if aws.ToString(v.VpcEndpointId) == vpcEndpointID {
				output = &v
				break
			}
		}
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if vpcEndpointState := string(output.VpcEndpointState); vpcEndpointState == vpcEndpointStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     vpcEndpointState,
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPCEndpointConnectionNotificationV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointConnectionNotificationsInput) (*awstypes.ConnectionNotification, error) {
	output, err := findVPCEndpointConnectionNotificationsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCEndpointConnectionNotificationsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointConnectionNotificationsInput) ([]awstypes.ConnectionNotification, error) {
	var output []awstypes.ConnectionNotification

	pages := ec2.NewDescribeVpcEndpointConnectionNotificationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidConnectionNotification) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.ConnectionNotificationSet...)
	}

	return output, nil
}

func findVPCEndpointConnectionNotificationByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ConnectionNotification, error) {
	input := &ec2.DescribeVpcEndpointConnectionNotificationsInput{
		ConnectionNotificationId: aws.String(id),
	}

	output, err := findVPCEndpointConnectionNotificationV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.ConnectionNotificationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPCEndpointServicePermissionV2(ctx context.Context, conn *ec2.Client, serviceID, principalARN string) (*awstypes.AllowedPrincipal, error) {
	// Applying a server-side filter on "principal" can lead to errors like
	// "An error occurred (InvalidFilter) when calling the DescribeVpcEndpointServicePermissions operation: The filter value arn:aws:iam::123456789012:role/developer contains unsupported characters".
	// Apply the filter client-side.
	input := &ec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(serviceID),
	}

	allowedPrincipals, err := findVPCEndpointServicePermissionsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	allowedPrincipals = tfslices.Filter(allowedPrincipals, func(v awstypes.AllowedPrincipal) bool {
		return aws.ToString(v.Principal) == principalARN
	})

	return tfresource.AssertSingleValueResult(allowedPrincipals)
}

func findClientVPNEndpoint(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnEndpointsInput) (*awstypes.ClientVpnEndpoint, error) {
	output, err := findClientVPNEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClientVPNEndpoints(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnEndpointsInput) ([]awstypes.ClientVpnEndpoint, error) {
	var output []awstypes.ClientVpnEndpoint

	pages := ec2.NewDescribeClientVpnEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ClientVpnEndpoints...)
	}

	return output, nil
}

func findClientVPNEndpointByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ClientVpnEndpoint, error) {
	input := &ec2.DescribeClientVpnEndpointsInput{
		ClientVpnEndpointIds: []string{id},
	}

	output, err := findClientVPNEndpoint(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.Status.Code; state == awstypes.ClientVpnEndpointStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.ClientVpnEndpointId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findClientVPNEndpointClientConnectResponseOptionsByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ClientConnectResponseOptions, error) {
	output, err := findClientVPNEndpointByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	if output.ClientConnectOptions == nil || output.ClientConnectOptions.Status == nil {
		return nil, tfresource.NewEmptyResultError(id)
	}

	return output.ClientConnectOptions, nil
}

func findClientVPNAuthorizationRule(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnAuthorizationRulesInput) (*awstypes.AuthorizationRule, error) {
	output, err := findClientVPNAuthorizationRules(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClientVPNAuthorizationRules(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnAuthorizationRulesInput) ([]awstypes.AuthorizationRule, error) {
	var output []awstypes.AuthorizationRule

	pages := ec2.NewDescribeClientVpnAuthorizationRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.AuthorizationRules...)
	}

	return output, nil
}

func findClientVPNAuthorizationRuleByThreePartKey(ctx context.Context, conn *ec2.Client, endpointID, targetNetworkCIDR, accessGroupID string) (*awstypes.AuthorizationRule, error) {
	filters := map[string]string{
		"destination-cidr": targetNetworkCIDR,
	}
	if accessGroupID != "" {
		filters["group-id"] = accessGroupID
	}
	input := &ec2.DescribeClientVpnAuthorizationRulesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters:             newAttributeFilterListV2(filters),
	}

	return findClientVPNAuthorizationRule(ctx, conn, input)
}

func findClientVPNNetworkAssociation(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnTargetNetworksInput) (*awstypes.TargetNetwork, error) {
	output, err := findClientVPNNetworkAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClientVPNNetworkAssociations(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnTargetNetworksInput) ([]awstypes.TargetNetwork, error) {
	var output []awstypes.TargetNetwork

	pages := ec2.NewDescribeClientVpnTargetNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound, errCodeInvalidClientVPNAssociationIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ClientVpnTargetNetworks...)
	}

	return output, nil
}

func findClientVPNNetworkAssociationByTwoPartKey(ctx context.Context, conn *ec2.Client, associationID, endpointID string) (*awstypes.TargetNetwork, error) {
	input := &ec2.DescribeClientVpnTargetNetworksInput{
		AssociationIds:      []string{associationID},
		ClientVpnEndpointId: aws.String(endpointID),
	}

	output, err := findClientVPNNetworkAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.Status.Code; state == awstypes.AssociationStatusCodeDisassociated {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.ClientVpnEndpointId) != endpointID || aws.ToString(output.AssociationId) != associationID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findClientVPNRoute(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnRoutesInput) (*awstypes.ClientVpnRoute, error) {
	output, err := findClientVPNRoutes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClientVPNRoutes(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnRoutesInput) ([]awstypes.ClientVpnRoute, error) {
	var output []awstypes.ClientVpnRoute

	pages := ec2.NewDescribeClientVpnRoutesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Routes...)
	}

	return output, nil
}

func findClientVPNRouteByThreePartKey(ctx context.Context, conn *ec2.Client, endpointID, targetSubnetID, destinationCIDR string) (*awstypes.ClientVpnRoute, error) {
	input := &ec2.DescribeClientVpnRoutesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters: newAttributeFilterListV2(map[string]string{
			"destination-cidr": destinationCIDR,
			"target-subnet":    targetSubnetID,
		}),
	}

	return findClientVPNRoute(ctx, conn, input)
}

func findCarrierGateway(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCarrierGatewaysInput) (*awstypes.CarrierGateway, error) {
	output, err := findCarrierGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCarrierGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCarrierGatewaysInput) ([]awstypes.CarrierGateway, error) {
	var output []awstypes.CarrierGateway

	pages := ec2.NewDescribeCarrierGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidCarrierGatewayIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.CarrierGateways...)
	}

	return output, nil
}

func findCarrierGatewayByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.CarrierGateway, error) {
	input := &ec2.DescribeCarrierGatewaysInput{
		CarrierGatewayIds: []string{id},
	}

	output, err := findCarrierGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.CarrierGatewayStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.CarrierGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPNConnection(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpnConnectionsInput) (*awstypes.VpnConnection, error) {
	output, err := findVPNConnections(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPNConnections(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpnConnectionsInput) ([]awstypes.VpnConnection, error) {
	output, err := conn.DescribeVpnConnections(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNConnectionIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output.VpnConnections, nil
}

func findVPNConnectionByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpnConnection, error) {
	input := &ec2.DescribeVpnConnectionsInput{
		VpnConnectionIds: []string{id},
	}

	output, err := findVPNConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.VpnStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.VpnConnectionId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPNConnectionRouteByTwoPartKey(ctx context.Context, conn *ec2.Client, vpnConnectionID, cidrBlock string) (*awstypes.VpnStaticRoute, error) {
	input := &ec2.DescribeVpnConnectionsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"route.destination-cidr-block": cidrBlock,
			"vpn-connection-id":            vpnConnectionID,
		}),
	}

	output, err := findVPNConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Routes {
		if aws.ToString(v.DestinationCidrBlock) == cidrBlock && v.State != awstypes.VpnStateDeleted {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError: fmt.Errorf("EC2 VPN Connection (%s) Route (%s) not found", vpnConnectionID, cidrBlock),
	}
}

func findVPNGatewayVPCAttachmentByTwoPartKey(ctx context.Context, conn *ec2.Client, vpnGatewayID, vpcID string) (*awstypes.VpcAttachment, error) {
	vpnGateway, err := findVPNGatewayByID(ctx, conn, vpnGatewayID)

	if err != nil {
		return nil, err
	}

	for _, vpcAttachment := range vpnGateway.VpcAttachments {
		if aws.ToString(vpcAttachment.VpcId) == vpcID {
			if state := vpcAttachment.State; state == awstypes.AttachmentStatusDetached {
				return nil, &retry.NotFoundError{
					Message:     string(state),
					LastRequest: vpcID,
				}
			}

			return &vpcAttachment, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(vpcID)
}

func findVPNGateway(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpnGatewaysInput) (*awstypes.VpnGateway, error) {
	output, err := findVPNGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPNGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpnGatewaysInput) ([]awstypes.VpnGateway, error) {
	output, err := conn.DescribeVpnGateways(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNGatewayIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VpnGateways, nil
}

func findVPNGatewayByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpnGateway, error) {
	input := &ec2.DescribeVpnGatewaysInput{
		VpnGatewayIds: []string{id},
	}

	output, err := findVPNGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.VpnStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.VpnGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayAttachmentV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayAttachmentsInput) (*awstypes.TransitGatewayAttachment, error) {
	output, err := findTransitGatewayAttachmentsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayAttachmentsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayAttachmentsInput) ([]awstypes.TransitGatewayAttachment, error) {
	var output []awstypes.TransitGatewayAttachment

	pages := ec2.NewDescribeTransitGatewayAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayAttachments...)
	}

	return output, nil
}

func findCustomerGateway(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCustomerGatewaysInput) (*awstypes.CustomerGateway, error) {
	output, err := findCustomerGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCustomerGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCustomerGatewaysInput) ([]awstypes.CustomerGateway, error) {
	output, err := conn.DescribeCustomerGateways(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCustomerGatewayIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.CustomerGateways, nil
}

func findCustomerGatewayByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.CustomerGateway, error) {
	input := &ec2.DescribeCustomerGatewaysInput{
		CustomerGatewayIds: []string{id},
	}

	output, err := findCustomerGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.ToString(output.State); state == CustomerGatewayStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.CustomerGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAM(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamsInput) (*awstypes.Ipam, error) {
	output, err := findIPAMs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamsInput) ([]awstypes.Ipam, error) {
	var output []awstypes.Ipam

	pages := ec2.NewDescribeIpamsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Ipams...)
	}

	return output, nil
}

func findIPAMByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Ipam, error) {
	input := &ec2.DescribeIpamsInput{
		IpamIds: []string{id},
	}

	output, err := findIPAM(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.IpamStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.IpamId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMPool(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamPoolsInput) (*awstypes.IpamPool, error) {
	output, err := findIPAMPools(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMPools(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamPoolsInput) ([]awstypes.IpamPool, error) {
	var output []awstypes.IpamPool

	pages := ec2.NewDescribeIpamPoolsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpamPools...)
	}

	return output, nil
}

func findIPAMPoolByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.IpamPool, error) {
	input := &ec2.DescribeIpamPoolsInput{
		IpamPoolIds: []string{id},
	}

	output, err := findIPAMPool(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.IpamPoolStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.IpamPoolId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMPoolAllocation(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolAllocationsInput) (*awstypes.IpamPoolAllocation, error) {
	output, err := findIPAMPoolAllocations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMPoolAllocations(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolAllocationsInput) ([]awstypes.IpamPoolAllocation, error) {
	var output []awstypes.IpamPoolAllocation

	pages := ec2.NewGetIpamPoolAllocationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolAllocationIdNotFound, errCodeInvalidIPAMPoolIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpamPoolAllocations...)
	}

	return output, nil
}

func findIPAMPoolAllocationByTwoPartKey(ctx context.Context, conn *ec2.Client, allocationID, poolID string) (*awstypes.IpamPoolAllocation, error) {
	input := &ec2.GetIpamPoolAllocationsInput{
		IpamPoolAllocationId: aws.String(allocationID),
		IpamPoolId:           aws.String(poolID),
	}

	output, err := findIPAMPoolAllocation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.IpamPoolAllocationId) != allocationID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMPoolCIDR(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolCidrsInput) (*awstypes.IpamPoolCidr, error) {
	output, err := findIPAMPoolCIDRs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMPoolCIDRs(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolCidrsInput) ([]awstypes.IpamPoolCidr, error) {
	var output []awstypes.IpamPoolCidr

	pages := ec2.NewGetIpamPoolCidrsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpamPoolCidrs...)
	}

	return output, nil
}

func findIPAMPoolCIDRByTwoPartKey(ctx context.Context, conn *ec2.Client, cidrBlock, poolID string) (*awstypes.IpamPoolCidr, error) {
	input := &ec2.GetIpamPoolCidrsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"cidr": cidrBlock,
		}),
		IpamPoolId: aws.String(poolID),
	}

	output, err := findIPAMPoolCIDR(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.IpamPoolCidrStateDeprovisioned {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.Cidr) != cidrBlock {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMPoolCIDRByPoolCIDRIDAndPoolID(ctx context.Context, conn *ec2.Client, poolCIDRID, poolID string) (*awstypes.IpamPoolCidr, error) {
	input := &ec2.GetIpamPoolCidrsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"ipam-pool-cidr-id": poolCIDRID,
		}),
		IpamPoolId: aws.String(poolID),
	}

	output, err := findIPAMPoolCIDR(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check
	if aws.ToString(output.Cidr) == "" {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	if state := output.State; state == awstypes.IpamPoolCidrStateDeprovisioned {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMResourceDiscovery(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamResourceDiscoveriesInput) (*awstypes.IpamResourceDiscovery, error) {
	output, err := findIPAMResourceDiscoveries(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMResourceDiscoveries(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamResourceDiscoveriesInput) ([]awstypes.IpamResourceDiscovery, error) {
	var output []awstypes.IpamResourceDiscovery

	pages := ec2.NewDescribeIpamResourceDiscoveriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMResourceDiscoveryIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpamResourceDiscoveries...)
	}

	return output, nil
}

func findIPAMResourceDiscoveryByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.IpamResourceDiscovery, error) {
	input := &ec2.DescribeIpamResourceDiscoveriesInput{
		IpamResourceDiscoveryIds: []string{id},
	}

	output, err := findIPAMResourceDiscovery(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.IpamResourceDiscoveryStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.IpamResourceDiscoveryId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMResourceDiscoveryAssociation(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamResourceDiscoveryAssociationsInput) (*awstypes.IpamResourceDiscoveryAssociation, error) {
	output, err := findIPAMResourceDiscoveryAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMResourceDiscoveryAssociations(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamResourceDiscoveryAssociationsInput) ([]awstypes.IpamResourceDiscoveryAssociation, error) {
	var output []awstypes.IpamResourceDiscoveryAssociation

	pages := ec2.NewDescribeIpamResourceDiscoveryAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMResourceDiscoveryAssociationIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpamResourceDiscoveryAssociations...)
	}

	return output, nil
}

func findIPAMResourceDiscoveryAssociationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.IpamResourceDiscoveryAssociation, error) {
	input := &ec2.DescribeIpamResourceDiscoveryAssociationsInput{
		IpamResourceDiscoveryAssociationIds: []string{id},
	}

	output, err := findIPAMResourceDiscoveryAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.IpamResourceDiscoveryAssociationStateDisassociateComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.IpamResourceDiscoveryAssociationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMScope(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamScopesInput) (*awstypes.IpamScope, error) {
	output, err := findIPAMScopes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMScopes(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamScopesInput) ([]awstypes.IpamScope, error) {
	var output []awstypes.IpamScope

	pages := ec2.NewDescribeIpamScopesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMScopeIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpamScopes...)
	}

	return output, nil
}

func findIPAMScopeByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.IpamScope, error) {
	input := &ec2.DescribeIpamScopesInput{
		IpamScopeIds: []string{id},
	}

	output, err := findIPAMScope(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.IpamScopeStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.IpamScopeId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGateway(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewaysInput) (*awstypes.TransitGateway, error) {
	output, err := findTransitGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.TransitGateway) bool { return v.Options != nil })
}

func findTransitGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewaysInput) ([]awstypes.TransitGateway, error) {
	var output []awstypes.TransitGateway

	pages := ec2.NewDescribeTransitGatewaysPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGateways...)
	}

	return output, nil
}

func findTransitGatewayByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGateway, error) {
	input := &ec2.DescribeTransitGatewaysInput{
		TransitGatewayIds: []string{id},
	}

	output, err := findTransitGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayAttachment(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayAttachmentsInput) (*awstypes.TransitGatewayAttachment, error) {
	output, err := findTransitGatewayAttachments(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayAttachments(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayAttachmentsInput) ([]awstypes.TransitGatewayAttachment, error) {
	var output []awstypes.TransitGatewayAttachment

	pages := ec2.NewDescribeTransitGatewayAttachmentsPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayAttachments...)
	}

	return output, nil
}

func findTransitGatewayAttachmentByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayAttachment, error) {
	input := &ec2.DescribeTransitGatewayAttachmentsInput{
		TransitGatewayAttachmentIds: []string{id},
	}

	output, err := findTransitGatewayAttachment(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayConnect(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayConnectsInput) (*awstypes.TransitGatewayConnect, error) {
	output, err := findTransitGatewayConnects(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.TransitGatewayConnect) bool { return v.Options != nil })
}

func findTransitGatewayConnects(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayConnectsInput) ([]awstypes.TransitGatewayConnect, error) {
	var output []awstypes.TransitGatewayConnect

	pages := ec2.NewDescribeTransitGatewayConnectsPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayConnects...)
	}

	return output, nil
}

func findTransitGatewayConnectByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayConnect, error) {
	input := &ec2.DescribeTransitGatewayConnectsInput{
		TransitGatewayAttachmentIds: []string{id},
	}

	output, err := findTransitGatewayConnect(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayAttachmentStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayConnectPeer(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayConnectPeersInput) (*awstypes.TransitGatewayConnectPeer, error) {
	output, err := findTransitGatewayConnectPeers(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output,
		func(v *awstypes.TransitGatewayConnectPeer) bool { return v.ConnectPeerConfiguration != nil },
		func(v *awstypes.TransitGatewayConnectPeer) bool {
			return len(v.ConnectPeerConfiguration.BgpConfigurations) > 0
		},
	)
}

func findTransitGatewayConnectPeers(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayConnectPeersInput) ([]awstypes.TransitGatewayConnectPeer, error) {
	var output []awstypes.TransitGatewayConnectPeer

	pages := ec2.NewDescribeTransitGatewayConnectPeersPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayConnectPeerIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayConnectPeers...)
	}

	return output, nil
}

func findTransitGatewayConnectPeerByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayConnectPeer, error) {
	input := &ec2.DescribeTransitGatewayConnectPeersInput{
		TransitGatewayConnectPeerIds: []string{id},
	}

	output, err := findTransitGatewayConnectPeer(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayConnectPeerStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayConnectPeerId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayMulticastDomain(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayMulticastDomainsInput) (*awstypes.TransitGatewayMulticastDomain, error) {
	output, err := findTransitGatewayMulticastDomains(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.TransitGatewayMulticastDomain) bool { return v.Options != nil })
}

func findTransitGatewayMulticastDomains(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayMulticastDomainsInput) ([]awstypes.TransitGatewayMulticastDomain, error) {
	var output []awstypes.TransitGatewayMulticastDomain

	pages := ec2.NewDescribeTransitGatewayMulticastDomainsPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayMulticastDomains...)
	}

	return output, nil
}

func findTransitGatewayMulticastDomainByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayMulticastDomain, error) {
	input := &ec2.DescribeTransitGatewayMulticastDomainsInput{
		TransitGatewayMulticastDomainIds: []string{id},
	}

	output, err := findTransitGatewayMulticastDomain(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayMulticastDomainStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayMulticastDomainId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayMulticastDomainAssociation(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayMulticastDomainAssociationsInput) (*awstypes.TransitGatewayMulticastDomainAssociation, error) {
	output, err := findTransitGatewayMulticastDomainAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.TransitGatewayMulticastDomainAssociation) bool { return v.Subnet != nil })
}

func findTransitGatewayMulticastDomainAssociations(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayMulticastDomainAssociationsInput) ([]awstypes.TransitGatewayMulticastDomainAssociation, error) {
	var output []awstypes.TransitGatewayMulticastDomainAssociation

	pages := ec2.NewGetTransitGatewayMulticastDomainAssociationsPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.MulticastDomainAssociations...)
	}

	return output, nil
}

func findTransitGatewayMulticastDomainAssociationByThreePartKey(ctx context.Context, conn *ec2.Client, multicastDomainID, attachmentID, subnetID string) (*awstypes.TransitGatewayMulticastDomainAssociation, error) {
	input := &ec2.GetTransitGatewayMulticastDomainAssociationsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"subnet-id":                     subnetID,
			"transit-gateway-attachment-id": attachmentID,
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := findTransitGatewayMulticastDomainAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.Subnet.State; state == awstypes.TransitGatewayMulitcastDomainAssociationStateDisassociated {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != attachmentID || aws.ToString(output.Subnet.SubnetId) != subnetID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayMulticastGroups(ctx context.Context, conn *ec2.Client, input *ec2.SearchTransitGatewayMulticastGroupsInput) ([]awstypes.TransitGatewayMulticastGroup, error) {
	var output []awstypes.TransitGatewayMulticastGroup

	pages := ec2.NewSearchTransitGatewayMulticastGroupsPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.MulticastGroups...)
	}

	return output, nil
}

func findTransitGatewayMulticastGroupMemberByThreePartKey(ctx context.Context, conn *ec2.Client, multicastDomainID, groupIPAddress, eniID string) (*awstypes.TransitGatewayMulticastGroup, error) {
	input := &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"group-ip-address": groupIPAddress,
			"is-group-member":  "true",
			"is-group-source":  "false",
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := findTransitGatewayMulticastGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, v := range output {
		if aws.ToString(v.NetworkInterfaceId) == eniID {
			// Eventual consistency check.
			if aws.ToString(v.GroupIpAddress) != groupIPAddress || !aws.ToBool(v.GroupMember) {
				return nil, &retry.NotFoundError{
					LastRequest: input,
				}
			}

			return &v, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func findTransitGatewayMulticastGroupSourceByThreePartKey(ctx context.Context, conn *ec2.Client, multicastDomainID, groupIPAddress, eniID string) (*awstypes.TransitGatewayMulticastGroup, error) {
	input := &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"group-ip-address": groupIPAddress,
			"is-group-member":  "false",
			"is-group-source":  "true",
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := findTransitGatewayMulticastGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, v := range output {
		if aws.ToString(v.NetworkInterfaceId) == eniID {
			// Eventual consistency check.
			if aws.ToString(v.GroupIpAddress) != groupIPAddress || !aws.ToBool(v.GroupSource) {
				return nil, &retry.NotFoundError{
					LastRequest: input,
				}
			}

			return &v, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func findTransitGatewayPeeringAttachment(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayPeeringAttachmentsInput) (*awstypes.TransitGatewayPeeringAttachment, error) {
	output, err := findTransitGatewayPeeringAttachments(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output,
		func(v *awstypes.TransitGatewayPeeringAttachment) bool { return v.AccepterTgwInfo != nil },
		func(v *awstypes.TransitGatewayPeeringAttachment) bool { return v.RequesterTgwInfo != nil },
	)
}

func findTransitGatewayPeeringAttachments(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayPeeringAttachmentsInput) ([]awstypes.TransitGatewayPeeringAttachment, error) {
	var output []awstypes.TransitGatewayPeeringAttachment

	pages := ec2.NewDescribeTransitGatewayPeeringAttachmentsPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayPeeringAttachments...)
	}

	return output, nil
}

func findTransitGatewayPeeringAttachmentByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayPeeringAttachment, error) {
	input := &ec2.DescribeTransitGatewayPeeringAttachmentsInput{
		TransitGatewayAttachmentIds: []string{id},
	}

	output, err := findTransitGatewayPeeringAttachment(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/vpc/latest/tgw/tgw-vpc-attachments.html#vpc-attachment-lifecycle.
	switch state := output.State; state {
	case awstypes.TransitGatewayAttachmentStateDeleted,
		awstypes.TransitGatewayAttachmentStateFailed,
		awstypes.TransitGatewayAttachmentStateRejected:
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayPrefixListReference(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayPrefixListReferencesInput) (*awstypes.TransitGatewayPrefixListReference, error) {
	output, err := findTransitGatewayPrefixListReferences(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayPrefixListReferences(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayPrefixListReferencesInput) ([]awstypes.TransitGatewayPrefixListReference, error) {
	var output []awstypes.TransitGatewayPrefixListReference

	pages := ec2.NewGetTransitGatewayPrefixListReferencesPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayPrefixListReferences...)
	}

	return output, nil
}

func findTransitGatewayPrefixListReferenceByTwoPartKey(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, prefixListID string) (*awstypes.TransitGatewayPrefixListReference, error) {
	input := &ec2.GetTransitGatewayPrefixListReferencesInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"prefix-list-id": prefixListID,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := findTransitGatewayPrefixListReference(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.PrefixListId) != prefixListID || aws.ToString(output.TransitGatewayRouteTableId) != transitGatewayRouteTableID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayStaticRoute(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, destination string) (*awstypes.TransitGatewayRoute, error) {
	input := &ec2.SearchTransitGatewayRoutesInput{
		Filters: newAttributeFilterListV2(map[string]string{
			names.AttrType:             string(awstypes.TransitGatewayRouteTypeStatic),
			"route-search.exact-match": destination,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := findTransitGatewayRoutes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, route := range output {
		if v := aws.ToString(route.DestinationCidrBlock); types.CIDRBlocksEqual(v, destination) {
			if state := route.State; state == awstypes.TransitGatewayRouteStateDeleted {
				return nil, &retry.NotFoundError{
					Message:     string(state),
					LastRequest: input,
				}
			}

			route.DestinationCidrBlock = aws.String(types.CanonicalCIDRBlock(v))

			return &route, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func findTransitGatewayRoutes(ctx context.Context, conn *ec2.Client, input *ec2.SearchTransitGatewayRoutesInput) ([]awstypes.TransitGatewayRoute, error) {
	output, err := conn.SearchTransitGatewayRoutes(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Routes, err
}

func findTransitGatewayPolicyTable(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayPolicyTablesInput) (*awstypes.TransitGatewayPolicyTable, error) {
	output, err := findTransitGatewayPolicyTables(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayRouteTable(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayRouteTablesInput) (*awstypes.TransitGatewayRouteTable, error) {
	output, err := findTransitGatewayRouteTables(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayPolicyTables(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayPolicyTablesInput) ([]awstypes.TransitGatewayPolicyTable, error) {
	var output []awstypes.TransitGatewayPolicyTable

	pages := ec2.NewDescribeTransitGatewayPolicyTablesPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayPolicyTables...)
	}

	return output, nil
}

func findTransitGatewayRouteTables(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayRouteTablesInput) ([]awstypes.TransitGatewayRouteTable, error) {
	var output []awstypes.TransitGatewayRouteTable

	pages := ec2.NewDescribeTransitGatewayRouteTablesPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayRouteTables...)
	}

	return output, nil
}

func findTransitGatewayPolicyTableByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayPolicyTable, error) {
	input := &ec2.DescribeTransitGatewayPolicyTablesInput{
		TransitGatewayPolicyTableIds: []string{id},
	}

	output, err := findTransitGatewayPolicyTable(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayPolicyTableId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayRouteTableByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayRouteTable, error) {
	input := &ec2.DescribeTransitGatewayRouteTablesInput{
		TransitGatewayRouteTableIds: []string{id},
	}

	output, err := findTransitGatewayRouteTable(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayRouteTableStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayRouteTableId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayPolicyTableAssociationByTwoPartKey(ctx context.Context, conn *ec2.Client, transitGatewayPolicyTableID, transitGatewayAttachmentID string) (*awstypes.TransitGatewayPolicyTableAssociation, error) {
	input := &ec2.GetTransitGatewayPolicyTableAssociationsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"transit-gateway-attachment-id": transitGatewayAttachmentID,
		}),
		TransitGatewayPolicyTableId: aws.String(transitGatewayPolicyTableID),
	}

	output, err := findTransitGatewayPolicyTableAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayAssociationStateDisassociated {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != transitGatewayAttachmentID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, err
}

func findTransitGatewayRouteTableAssociationByTwoPartKey(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, transitGatewayAttachmentID string) (*awstypes.TransitGatewayRouteTableAssociation, error) {
	input := &ec2.GetTransitGatewayRouteTableAssociationsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"transit-gateway-attachment-id": transitGatewayAttachmentID,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := findTransitGatewayRouteTableAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayAssociationStateDisassociated {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != transitGatewayAttachmentID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, err
}

func findTransitGatewayRouteTableAssociation(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayRouteTableAssociationsInput) (*awstypes.TransitGatewayRouteTableAssociation, error) {
	output, err := findTransitGatewayRouteTableAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayPolicyTableAssociations(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayPolicyTableAssociationsInput) ([]awstypes.TransitGatewayPolicyTableAssociation, error) {
	var output []awstypes.TransitGatewayPolicyTableAssociation

	pages := ec2.NewGetTransitGatewayPolicyTableAssociationsPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Associations...)
	}

	return output, nil
}

func findTransitGatewayPolicyTableAssociation(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayPolicyTableAssociationsInput) (*awstypes.TransitGatewayPolicyTableAssociation, error) {
	output, err := findTransitGatewayPolicyTableAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayRouteTableAssociations(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayRouteTableAssociationsInput) ([]awstypes.TransitGatewayRouteTableAssociation, error) {
	var output []awstypes.TransitGatewayRouteTableAssociation

	pages := ec2.NewGetTransitGatewayRouteTableAssociationsPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Associations...)
	}

	return output, nil
}

func findTransitGatewayRouteTablePropagationByTwoPartKey(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID string, transitGatewayAttachmentID string) (*awstypes.TransitGatewayRouteTablePropagation, error) {
	input := &ec2.GetTransitGatewayRouteTablePropagationsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"transit-gateway-attachment-id": transitGatewayAttachmentID,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := findTransitGatewayRouteTablePropagation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayPropagationStateDisabled {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != transitGatewayAttachmentID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, err
}

func findTransitGatewayRouteTablePropagation(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayRouteTablePropagationsInput) (*awstypes.TransitGatewayRouteTablePropagation, error) {
	output, err := findTransitGatewayRouteTablePropagations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayRouteTablePropagations(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayRouteTablePropagationsInput) ([]awstypes.TransitGatewayRouteTablePropagation, error) {
	var output []awstypes.TransitGatewayRouteTablePropagation

	pages := ec2.NewGetTransitGatewayRouteTablePropagationsPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayRouteTablePropagations...)
	}

	return output, nil
}

func findTransitGatewayVPCAttachment(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayVpcAttachmentsInput) (*awstypes.TransitGatewayVpcAttachment, error) {
	output, err := findTransitGatewayVPCAttachments(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.TransitGatewayVpcAttachment) bool { return v.Options != nil })
}

func findTransitGatewayVPCAttachments(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayVpcAttachmentsInput) ([]awstypes.TransitGatewayVpcAttachment, error) {
	var output []awstypes.TransitGatewayVpcAttachment

	pages := ec2.NewDescribeTransitGatewayVpcAttachmentsPaginator(conn, input)
	if pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayVpcAttachments...)
	}

	return output, nil
}

func findTransitGatewayVPCAttachmentByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayVpcAttachment, error) {
	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{
		TransitGatewayAttachmentIds: []string{id},
	}

	output, err := findTransitGatewayVPCAttachment(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/vpc/latest/tgw/tgw-vpc-attachments.html#vpc-attachment-lifecycle.
	switch state := output.State; state {
	case awstypes.TransitGatewayAttachmentStateDeleted,
		awstypes.TransitGatewayAttachmentStateFailed,
		awstypes.TransitGatewayAttachmentStateRejected:
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}
