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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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

func findVPCMainRouteTableV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"association.main": "true",
			"vpc-id":           id,
		}),
	}

	return findRouteTableV2(ctx, conn, input)
}

func findRouteTableV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeRouteTablesInput) (*awstypes.RouteTable, error) {
	output, err := findRouteTablesV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findRouteTablesV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeRouteTablesInput) ([]awstypes.RouteTable, error) {
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

func FindIPAM(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamsInput) (*awstypes.Ipam, error) {
	output, err := FindIPAMs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindIPAMs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamsInput) ([]awstypes.Ipam, error) {
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

func FindIPAMByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Ipam, error) {
	input := &ec2.DescribeIpamsInput{
		IpamIds: []string{id},
	}

	output, err := FindIPAM(ctx, conn, input)

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

func FindIPAMPool(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamPoolsInput) (*awstypes.IpamPool, error) {
	output, err := FindIPAMPools(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindIPAMPools(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamPoolsInput) ([]awstypes.IpamPool, error) {
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

func FindIPAMPoolByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.IpamPool, error) {
	input := &ec2.DescribeIpamPoolsInput{
		IpamPoolIds: []string{id},
	}

	output, err := FindIPAMPool(ctx, conn, input)

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

func FindIPAMPoolAllocation(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolAllocationsInput) (*awstypes.IpamPoolAllocation, error) {
	output, err := FindIPAMPoolAllocations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindIPAMPoolAllocations(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolAllocationsInput) ([]awstypes.IpamPoolAllocation, error) {
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

func FindIPAMPoolAllocationByTwoPartKey(ctx context.Context, conn *ec2.Client, allocationID, poolID string) (*awstypes.IpamPoolAllocation, error) {
	input := &ec2.GetIpamPoolAllocationsInput{
		IpamPoolAllocationId: aws.String(allocationID),
		IpamPoolId:           aws.String(poolID),
	}

	output, err := FindIPAMPoolAllocation(ctx, conn, input)

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

func FindIPAMPoolCIDR(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolCidrsInput) (*awstypes.IpamPoolCidr, error) {
	output, err := FindIPAMPoolCIDRs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindIPAMPoolCIDRs(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolCidrsInput) ([]awstypes.IpamPoolCidr, error) {
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

func FindIPAMPoolCIDRByTwoPartKey(ctx context.Context, conn *ec2.Client, cidrBlock, poolID string) (*awstypes.IpamPoolCidr, error) {
	input := &ec2.GetIpamPoolCidrsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"cidr": cidrBlock,
		}),
		IpamPoolId: aws.String(poolID),
	}

	output, err := FindIPAMPoolCIDR(ctx, conn, input)

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

func FindIPAMPoolCIDRByPoolCIDRId(ctx context.Context, conn *ec2.Client, poolCidrId, poolID string) (*awstypes.IpamPoolCidr, error) {
	input := &ec2.GetIpamPoolCidrsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"ipam-pool-cidr-id": poolCidrId,
		}),
		IpamPoolId: aws.String(poolID),
	}

	output, err := FindIPAMPoolCIDR(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check
	cidrBlock := aws.ToString(output.Cidr)
	if cidrBlock == "" {
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

func FindIPAMResourceDiscovery(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamResourceDiscoveriesInput) (*awstypes.IpamResourceDiscovery, error) {
	output, err := FindIPAMResourceDiscoveries(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindIPAMResourceDiscoveries(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamResourceDiscoveriesInput) ([]awstypes.IpamResourceDiscovery, error) {
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

func FindIPAMResourceDiscoveryByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.IpamResourceDiscovery, error) {
	input := &ec2.DescribeIpamResourceDiscoveriesInput{
		IpamResourceDiscoveryIds: []string{id},
	}

	output, err := FindIPAMResourceDiscovery(ctx, conn, input)

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

func FindIPAMResourceDiscoveryAssociation(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamResourceDiscoveryAssociationsInput) (*awstypes.IpamResourceDiscoveryAssociation, error) {
	output, err := FindIPAMResourceDiscoveryAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindIPAMResourceDiscoveryAssociations(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamResourceDiscoveryAssociationsInput) ([]awstypes.IpamResourceDiscoveryAssociation, error) {
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

func FindIPAMResourceDiscoveryAssociationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.IpamResourceDiscoveryAssociation, error) {
	input := &ec2.DescribeIpamResourceDiscoveryAssociationsInput{
		IpamResourceDiscoveryAssociationIds: []string{id},
	}

	output, err := FindIPAMResourceDiscoveryAssociation(ctx, conn, input)

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

func FindIPAMScope(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamScopesInput) (*awstypes.IpamScope, error) {
	output, err := FindIPAMScopes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindIPAMScopes(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamScopesInput) ([]awstypes.IpamScope, error) {
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

func FindIPAMScopeByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.IpamScope, error) {
	input := &ec2.DescribeIpamScopesInput{
		IpamScopeIds: []string{id},
	}

	output, err := FindIPAMScope(ctx, conn, input)

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

func FindVPCV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcsInput) (*awstypes.Vpc, error) {
	output, err := FindVPCsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindVPCsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcsInput) ([]awstypes.Vpc, error) {
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

func FindVPCIPv6CIDRBlockAssociationByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpcIpv6CidrBlockAssociation, *awstypes.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"ipv6-cidr-block-association.association-id": id,
		}),
	}

	vpc, err := FindVPCV2(ctx, conn, input)

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
