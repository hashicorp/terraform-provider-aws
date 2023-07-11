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
)

func FindVPCAttributeV2(ctx context.Context, conn *ec2.Client, vpcID string, attribute string) (bool, error) {
	input := &ec2.DescribeVpcAttributeInput{
		Attribute: awstypes.VpcAttributeName(attribute),
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
	case string(awstypes.VpcAttributeNameEnableDnsHostnames):
		v = output.EnableDnsHostnames
	case string(awstypes.VpcAttributeNameEnableDnsSupport):
		v = output.EnableDnsSupport
	case string(awstypes.VpcAttributeNameEnableNetworkAddressUsageMetrics):
		v = output.EnableNetworkAddressUsageMetrics
	default:
		return false, fmt.Errorf("unsupported VPC attribute: %s", attribute)
	}

	if v == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	return aws.ToBool(v.Value), nil
}

func FindVPCV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcsInput) (*awstypes.Vpc, error) {
	output, err := FindVPCsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output[0], nil
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
			return nil, fmt.Errorf("reading VPCs: %s", err)
		}

		output = append(output, page.Vpcs...)
	}

	return output, nil
}

func FindVPCByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		VpcIds: []string{id},
	}

	output, err := FindVPCV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCIPv6CIDRBlockAssociationByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpcIpv6CidrBlockAssociation, *awstypes.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: BuildAttributeFilterListV2(map[string]string{
			"ipv6-cidr-block-association.association-id": id,
		}),
	}

	vpc, err := FindVPCV2(ctx, conn, input)

	if err != nil {
		return nil, nil, err
	}

	for _, association := range vpc.Ipv6CidrBlockAssociationSet {
		if aws.ToString(association.AssociationId) == id {
			if state := string(association.Ipv6CidrBlockState.State); state == string(awstypes.VpcCidrBlockStateCodeDisassociated) {
				return nil, nil, &retry.NotFoundError{Message: state}
			}

			return &association, vpc, nil
		}
	}

	return nil, nil, &retry.NotFoundError{}
}

func FindVPCDefaultNetworkACLV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterListV2(map[string]string{
			"default": "true",
			"vpc-id":  id,
		}),
	}

	return FindNetworkACLV2(ctx, conn, input)
}

func FindNetworkACLV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkAclsInput) (*awstypes.NetworkAcl, error) {
	output, err := FindNetworkACLsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output[0], nil
}

func FindNetworkACLsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkAclsInput) ([]awstypes.NetworkAcl, error) {
	var output []awstypes.NetworkAcl

	pages := ec2.NewDescribeNetworkAclsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("reading Network ACLs: %s", err)
		}

		output = append(output, page.NetworkAcls...)
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindVPCDefaultSecurityGroupV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: BuildAttributeFilterListV2(map[string]string{
			"group-name": DefaultSecurityGroupName,
			"vpc-id":     id,
		}),
	}

	return FindSecurityGroupV2(ctx, conn, input)
}

func FindVPCMainRouteTableV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterListV2(map[string]string{
			"association.main": "true",
			"vpc-id":           id,
		}),
	}

	return FindRouteTableV2(ctx, conn, input)
}

func FindRouteTableV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeRouteTablesInput) (*awstypes.RouteTable, error) {
	output, err := FindRouteTablesV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output[0], nil
}

func FindRouteTablesV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeRouteTablesInput) ([]awstypes.RouteTable, error) {
	var output []awstypes.RouteTable

	pages := ec2.NewDescribeRouteTablesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("reading Route Tables: %s", err)
		}

		output = append(output, page.RouteTables...)
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// FindSecurityGroupV2 looks up a security group using an ec2.DescribeSecurityGroupsInput. Returns a retry.NotFoundError if not found.
func FindSecurityGroupV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupsInput) (*awstypes.SecurityGroup, error) {
	output, err := FindSecurityGroupsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output[0], nil
}

func FindSecurityGroupsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupsInput) ([]awstypes.SecurityGroup, error) {
	var output []awstypes.SecurityGroup

	pages := ec2.NewDescribeSecurityGroupsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("reading Security Groups: %s", err)
		}

		output = append(output, page.SecurityGroups...)
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindIPAMPoolAllocationsV2(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolAllocationsInput) ([]awstypes.IpamPoolAllocation, error) {
	var output []awstypes.IpamPoolAllocation

	pages := ec2.NewGetIpamPoolAllocationsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("reading IPAM Pool Allocations: %s", err)
		}

		output = append(output, page.IpamPoolAllocations...)
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
