// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

//
// Move functions to findv2.go as they are migrated to AWS SDK for Go v2.
//

func FindNetworkACL(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkAclsInput) (*ec2.NetworkAcl, error) {
	output, err := FindNetworkACLs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindNetworkACLs(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkAclsInput) ([]*ec2.NetworkAcl, error) {
	var output []*ec2.NetworkAcl

	err := conn.DescribeNetworkAclsPagesWithContext(ctx, input, func(page *ec2.DescribeNetworkAclsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NetworkAcls {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkACLIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNetworkACLByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		NetworkAclIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkACL(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.NetworkAclId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNetworkACLAssociationByID(ctx context.Context, conn *ec2.EC2, associationID string) (*ec2.NetworkAclAssociation, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"association.association-id": associationID,
		}),
	}

	output, err := FindNetworkACL(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Associations {
		if aws.StringValue(v.NetworkAclAssociationId) == associationID {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindNetworkACLAssociationBySubnetID(ctx context.Context, conn *ec2.EC2, subnetID string) (*ec2.NetworkAclAssociation, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"association.subnet-id": subnetID,
		}),
	}

	output, err := FindNetworkACL(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Associations {
		if aws.StringValue(v.SubnetId) == subnetID {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindNetworkACLEntryByThreePartKey(ctx context.Context, conn *ec2.EC2, naclID string, egress bool, ruleNumber int) (*ec2.NetworkAclEntry, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"entry.egress":      strconv.FormatBool(egress),
			"entry.rule-number": strconv.Itoa(ruleNumber),
		}),
		NetworkAclIds: aws.StringSlice([]string{naclID}),
	}

	output, err := FindNetworkACL(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Entries {
		if aws.BoolValue(v.Egress) == egress && aws.Int64Value(v.RuleNumber) == int64(ruleNumber) {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindNetworkInterface(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInterfacesInput) (*ec2.NetworkInterface, error) {
	output, err := FindNetworkInterfaces(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindNetworkInterfaces(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInterfacesInput) ([]*ec2.NetworkInterface, error) {
	var output []*ec2.NetworkInterface

	err := conn.DescribeNetworkInterfacesPagesWithContext(ctx, input, func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NetworkInterfaces {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInterfaceIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNetworkInterfaceByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkInterface(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.NetworkInterfaceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNetworkInterfaceByAttachmentID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: newAttributeFilterList(map[string]string{
			"attachment.attachment-id": id,
		}),
	}

	networkInterface, err := FindNetworkInterface(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if networkInterface == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return networkInterface, nil
}

func FindNetworkInterfaceAttachmentByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkInterfaceAttachment, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: newAttributeFilterList(map[string]string{
			"attachment.attachment-id": id,
		}),
	}

	networkInterface, err := FindNetworkInterface(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if networkInterface.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return networkInterface.Attachment, nil
}

func FindNetworkInterfaceSecurityGroup(ctx context.Context, conn *ec2.EC2, networkInterfaceID string, securityGroupID string) (*ec2.GroupIdentifier, error) {
	networkInterface, err := FindNetworkInterfaceByID(ctx, conn, networkInterfaceID)

	if err != nil {
		return nil, err
	}

	for _, groupIdentifier := range networkInterface.Groups {
		if aws.StringValue(groupIdentifier.GroupId) == securityGroupID {
			return groupIdentifier, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError: fmt.Errorf("Network Interface (%s) Security Group (%s) not found", networkInterfaceID, securityGroupID),
	}
}

func FindSecurityGroupByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSecurityGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.GroupId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findSecurityGroupByDescriptionAndVPCID(ctx context.Context, conn *ec2.EC2, description, vpcID string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"description": description, // nosemgrep:ci.literal-description-string-constant
				"vpc-id":      vpcID,
			},
		),
	}
	return FindSecurityGroup(ctx, conn, input)
}

func findSecurityGroupByNameAndVPCIDAndOwnerID(ctx context.Context, conn *ec2.EC2, name, vpcID, ownerID string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"group-name": name,
				"vpc-id":     vpcID,
				"owner-id":   ownerID,
			},
		),
	}
	return FindSecurityGroup(ctx, conn, input)
}

// FindSecurityGroup looks up a security group using an ec2.DescribeSecurityGroupsInput. Returns a retry.NotFoundError if not found.
func FindSecurityGroup(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSecurityGroupsInput) (*ec2.SecurityGroup, error) {
	output, err := FindSecurityGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindSecurityGroups(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSecurityGroupsInput) ([]*ec2.SecurityGroup, error) {
	var output []*ec2.SecurityGroup

	err := conn.DescribeSecurityGroupsPagesWithContext(ctx, input, func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SecurityGroups {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound, errCodeInvalidSecurityGroupIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSecurityGroupRule(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSecurityGroupRulesInput) (*ec2.SecurityGroupRule, error) {
	output, err := FindSecurityGroupRules(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindSecurityGroupRules(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSecurityGroupRulesInput) ([]*ec2.SecurityGroupRule, error) {
	var output []*ec2.SecurityGroupRule

	err := conn.DescribeSecurityGroupRulesPagesWithContext(ctx, input, func(page *ec2.DescribeSecurityGroupRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SecurityGroupRules {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSecurityGroupRuleIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSecurityGroupRuleByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SecurityGroupRule, error) {
	input := &ec2.DescribeSecurityGroupRulesInput{
		SecurityGroupRuleIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSecurityGroupRule(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.SecurityGroupRuleId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSecurityGroupEgressRuleByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SecurityGroupRule, error) {
	output, err := FindSecurityGroupRuleByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	if !aws.BoolValue(output.IsEgress) {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func FindSecurityGroupIngressRuleByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SecurityGroupRule, error) {
	output, err := FindSecurityGroupRuleByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	if aws.BoolValue(output.IsEgress) {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func FindSecurityGroupRulesBySecurityGroupID(ctx context.Context, conn *ec2.EC2, id string) ([]*ec2.SecurityGroupRule, error) {
	input := &ec2.DescribeSecurityGroupRulesInput{
		Filters: newAttributeFilterList(map[string]string{
			"group-id": id,
		}),
	}

	return FindSecurityGroupRules(ctx, conn, input)
}

func FindSubnetByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Subnet, error) {
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSubnet(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.SubnetId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSubnet(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSubnetsInput) (*ec2.Subnet, error) {
	output, err := FindSubnets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindSubnets(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSubnetsInput) ([]*ec2.Subnet, error) {
	var output []*ec2.Subnet

	err := conn.DescribeSubnetsPagesWithContext(ctx, input, func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Subnets {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSubnetIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSubnetCIDRReservationBySubnetIDAndReservationID(ctx context.Context, conn *ec2.EC2, subnetID, reservationID string) (*ec2.SubnetCidrReservation, error) {
	input := &ec2.GetSubnetCidrReservationsInput{
		SubnetId: aws.String(subnetID),
	}

	output, err := conn.GetSubnetCidrReservationsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSubnetIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || (len(output.SubnetIpv4CidrReservations) == 0 && len(output.SubnetIpv6CidrReservations) == 0) {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, r := range output.SubnetIpv4CidrReservations {
		if aws.StringValue(r.SubnetCidrReservationId) == reservationID {
			return r, nil
		}
	}
	for _, r := range output.SubnetIpv6CidrReservations {
		if aws.StringValue(r.SubnetCidrReservationId) == reservationID {
			return r, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError:   err,
		LastRequest: input,
	}
}

func FindSubnetIPv6CIDRBlockAssociationByID(ctx context.Context, conn *ec2.EC2, associationID string) (*ec2.SubnetIpv6CidrBlockAssociation, error) {
	input := &ec2.DescribeSubnetsInput{
		Filters: newAttributeFilterList(map[string]string{
			"ipv6-cidr-block-association.association-id": associationID,
		}),
	}

	output, err := FindSubnet(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, association := range output.Ipv6CidrBlockAssociationSet {
		if aws.StringValue(association.AssociationId) == associationID {
			if state := aws.StringValue(association.Ipv6CidrBlockState.State); state == ec2.SubnetCidrBlockStateCodeDisassociated {
				return nil, &retry.NotFoundError{Message: state}
			}

			return association, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindVPC(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcsInput) (*ec2.Vpc, error) {
	output, err := FindVPCs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindVPCs(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcsInput) ([]*ec2.Vpc, error) {
	var output []*ec2.Vpc

	err := conn.DescribeVpcsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Vpcs {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		VpcIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPC(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpcId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPCDHCPOptionsAssociation(ctx context.Context, conn *ec2.EC2, vpcID string, dhcpOptionsID string) error {
	vpc, err := FindVPCByID(ctx, conn, vpcID)

	if err != nil {
		return err
	}

	if aws.StringValue(vpc.DhcpOptionsId) != dhcpOptionsID {
		return &retry.NotFoundError{
			LastError: fmt.Errorf("EC2 VPC (%s) DHCP Options Set (%s) Association not found", vpcID, dhcpOptionsID),
		}
	}

	return nil
}

func FindVPCCIDRBlockAssociationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VpcCidrBlockAssociation, *ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: newAttributeFilterList(map[string]string{
			"cidr-block-association.association-id": id,
		}),
	}

	vpc, err := FindVPC(ctx, conn, input)

	if err != nil {
		return nil, nil, err
	}

	for _, association := range vpc.CidrBlockAssociationSet {
		if aws.StringValue(association.AssociationId) == id {
			if state := aws.StringValue(association.CidrBlockState.State); state == ec2.VpcCidrBlockStateCodeDisassociated {
				return nil, nil, &retry.NotFoundError{Message: state}
			}

			return association, vpc, nil
		}
	}

	return nil, nil, &retry.NotFoundError{}
}

func FindVPCIPv6CIDRBlockAssociationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VpcIpv6CidrBlockAssociation, *ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: newAttributeFilterList(map[string]string{
			"ipv6-cidr-block-association.association-id": id,
		}),
	}

	vpc, err := FindVPC(ctx, conn, input)

	if err != nil {
		return nil, nil, err
	}

	for _, association := range vpc.Ipv6CidrBlockAssociationSet {
		if aws.StringValue(association.AssociationId) == id {
			if state := aws.StringValue(association.Ipv6CidrBlockState.State); state == ec2.VpcCidrBlockStateCodeDisassociated {
				return nil, nil, &retry.NotFoundError{Message: state}
			}

			return association, vpc, nil
		}
	}

	return nil, nil, &retry.NotFoundError{}
}

func FindVPCDefaultNetworkACL(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"default": "true",
			"vpc-id":  id,
		}),
	}

	return FindNetworkACL(ctx, conn, input)
}

func FindVPCDefaultSecurityGroup(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterList(map[string]string{
			"group-name": DefaultSecurityGroupName,
			"vpc-id":     id,
		}),
	}

	return FindSecurityGroup(ctx, conn, input)
}

func FindVPCPeeringConnection(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcPeeringConnectionsInput) (*ec2.VpcPeeringConnection, error) {
	output, err := FindVPCPeeringConnections(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output, func(v *ec2.VpcPeeringConnection) bool { return v.Status != nil })
}

func FindVPCPeeringConnections(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcPeeringConnectionsInput) ([]*ec2.VpcPeeringConnection, error) {
	var output []*ec2.VpcPeeringConnection

	err := conn.DescribeVpcPeeringConnectionsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcPeeringConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VpcPeeringConnections {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCPeeringConnectionIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCPeeringConnectionByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VpcPeeringConnection, error) {
	input := &ec2.DescribeVpcPeeringConnectionsInput{
		VpcPeeringConnectionIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPCPeeringConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/vpc/latest/peering/vpc-peering-basics.html#vpc-peering-lifecycle.
	switch statusCode := aws.StringValue(output.Status.Code); statusCode {
	case ec2.VpcPeeringConnectionStateReasonCodeDeleted,
		ec2.VpcPeeringConnectionStateReasonCodeExpired,
		ec2.VpcPeeringConnectionStateReasonCodeFailed,
		ec2.VpcPeeringConnectionStateReasonCodeRejected:
		return nil, &retry.NotFoundError{
			Message:     statusCode,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpcPeeringConnectionId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindDHCPOptions(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeDhcpOptionsInput) (*ec2.DhcpOptions, error) {
	output, err := FindDHCPOptionses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindDHCPOptionses(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeDhcpOptionsInput) ([]*ec2.DhcpOptions, error) {
	var output []*ec2.DhcpOptions

	err := conn.DescribeDhcpOptionsPagesWithContext(ctx, input, func(page *ec2.DescribeDhcpOptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DhcpOptions {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidDHCPOptionIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindDHCPOptionsByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.DhcpOptions, error) {
	input := &ec2.DescribeDhcpOptionsInput{
		DhcpOptionsIds: aws.StringSlice([]string{id}),
	}

	output, err := FindDHCPOptions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.DhcpOptionsId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindEgressOnlyInternetGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeEgressOnlyInternetGatewaysInput) (*ec2.EgressOnlyInternetGateway, error) {
	output, err := FindEgressOnlyInternetGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindEgressOnlyInternetGateways(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeEgressOnlyInternetGatewaysInput) ([]*ec2.EgressOnlyInternetGateway, error) {
	var output []*ec2.EgressOnlyInternetGateway

	err := conn.DescribeEgressOnlyInternetGatewaysPagesWithContext(ctx, input, func(page *ec2.DescribeEgressOnlyInternetGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.EgressOnlyInternetGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindEgressOnlyInternetGatewayByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.EgressOnlyInternetGateway, error) {
	input := &ec2.DescribeEgressOnlyInternetGatewaysInput{
		EgressOnlyInternetGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindEgressOnlyInternetGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.EgressOnlyInternetGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindFlowLogByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.FlowLog, error) {
	input := &ec2.DescribeFlowLogsInput{
		FlowLogIds: aws.StringSlice([]string{id}),
	}

	output, err := FindFlowLog(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.FlowLogId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindFlowLogs(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeFlowLogsInput) ([]*ec2.FlowLog, error) {
	var output []*ec2.FlowLog

	err := conn.DescribeFlowLogsPagesWithContext(ctx, input, func(page *ec2.DescribeFlowLogsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FlowLogs {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindFlowLog(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeFlowLogsInput) (*ec2.FlowLog, error) {
	output, err := FindFlowLogs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindInternetGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInternetGatewaysInput) (*ec2.InternetGateway, error) {
	output, err := FindInternetGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindInternetGateways(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInternetGatewaysInput) ([]*ec2.InternetGateway, error) {
	var output []*ec2.InternetGateway

	err := conn.DescribeInternetGatewaysPagesWithContext(ctx, input, func(page *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InternetGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidInternetGatewayIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindInternetGatewayByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.InternetGateway, error) {
	input := &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindInternetGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.InternetGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindInternetGatewayAttachment(ctx context.Context, conn *ec2.EC2, internetGatewayID, vpcID string) (*ec2.InternetGatewayAttachment, error) {
	internetGateway, err := FindInternetGatewayByID(ctx, conn, internetGatewayID)

	if err != nil {
		return nil, err
	}

	if len(internetGateway.Attachments) == 0 || internetGateway.Attachments[0] == nil {
		return nil, tfresource.NewEmptyResultError(internetGatewayID)
	}

	if count := len(internetGateway.Attachments); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, internetGatewayID)
	}

	attachment := internetGateway.Attachments[0]

	if aws.StringValue(attachment.VpcId) != vpcID {
		return nil, tfresource.NewEmptyResultError(vpcID)
	}

	return attachment, nil
}

func FindManagedPrefixList(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeManagedPrefixListsInput) (*ec2.ManagedPrefixList, error) {
	output, err := FindManagedPrefixLists(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindManagedPrefixLists(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeManagedPrefixListsInput) ([]*ec2.ManagedPrefixList, error) {
	var output []*ec2.ManagedPrefixList

	err := conn.DescribeManagedPrefixListsPagesWithContext(ctx, input, func(page *ec2.DescribeManagedPrefixListsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.PrefixLists {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPrefixListIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindManagedPrefixListByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	input := &ec2.DescribeManagedPrefixListsInput{
		PrefixListIds: aws.StringSlice([]string{id}),
	}

	output, err := FindManagedPrefixList(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.PrefixListStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.PrefixListId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindManagedPrefixListEntries(ctx context.Context, conn *ec2.EC2, input *ec2.GetManagedPrefixListEntriesInput) ([]*ec2.PrefixListEntry, error) {
	var output []*ec2.PrefixListEntry

	err := conn.GetManagedPrefixListEntriesPagesWithContext(ctx, input, func(page *ec2.GetManagedPrefixListEntriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Entries {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPrefixListIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindManagedPrefixListEntriesByID(ctx context.Context, conn *ec2.EC2, id string) ([]*ec2.PrefixListEntry, error) {
	input := &ec2.GetManagedPrefixListEntriesInput{
		PrefixListId: aws.String(id),
	}

	return FindManagedPrefixListEntries(ctx, conn, input)
}

func FindManagedPrefixListEntryByIDAndCIDR(ctx context.Context, conn *ec2.EC2, id, cidr string) (*ec2.PrefixListEntry, error) {
	prefixListEntries, err := FindManagedPrefixListEntriesByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	for _, v := range prefixListEntries {
		if aws.StringValue(v.Cidr) == cidr {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindNATGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNatGatewaysInput) (*ec2.NatGateway, error) {
	output, err := FindNATGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindNATGateways(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNatGatewaysInput) ([]*ec2.NatGateway, error) {
	var output []*ec2.NatGateway

	err := conn.DescribeNatGatewaysPagesWithContext(ctx, input, func(page *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NatGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeNatGatewayNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNATGatewayByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NatGateway, error) {
	input := &ec2.DescribeNatGatewaysInput{
		NatGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNATGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.NatGatewayStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.NatGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNATGatewayAddressByNATGatewayIDAndAllocationID(ctx context.Context, conn *ec2.EC2, natGatewayID, allocationID string) (*ec2.NatGatewayAddress, error) {
	output, err := FindNATGatewayByID(ctx, conn, natGatewayID)

	if err != nil {
		return nil, err
	}

	for _, v := range output.NatGatewayAddresses {
		if aws.StringValue(v.AllocationId) == allocationID {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindNATGatewayAddressByNATGatewayIDAndPrivateIP(ctx context.Context, conn *ec2.EC2, natGatewayID, privateIP string) (*ec2.NatGatewayAddress, error) {
	output, err := FindNATGatewayByID(ctx, conn, natGatewayID)

	if err != nil {
		return nil, err
	}

	for _, v := range output.NatGatewayAddresses {
		if aws.StringValue(v.PrivateIp) == privateIP {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindPrefixList(ctx context.Context, conn *ec2.EC2, input *ec2.DescribePrefixListsInput) (*ec2.PrefixList, error) {
	output, err := FindPrefixLists(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func FindPrefixLists(ctx context.Context, conn *ec2.EC2, input *ec2.DescribePrefixListsInput) ([]*ec2.PrefixList, error) {
	var output []*ec2.PrefixList

	err := conn.DescribePrefixListsPagesWithContext(ctx, input, func(page *ec2.DescribePrefixListsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.PrefixLists {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPrefixListIdNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindPrefixListByName(ctx context.Context, conn *ec2.EC2, name string) (*ec2.PrefixList, error) {
	input := &ec2.DescribePrefixListsInput{
		Filters: newAttributeFilterList(map[string]string{
			"prefix-list-name": name,
		}),
	}

	return FindPrefixList(ctx, conn, input)
}
