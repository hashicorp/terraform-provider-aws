// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

//
// Move functions to statusv2.go as they are migrated to AWS SDK for Go v2.
//

func StatusNATGatewayState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNATGatewayByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusNATGatewayAddressByNATGatewayIDAndAllocationID(ctx context.Context, conn *ec2.EC2, natGatewayID, allocationID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNATGatewayAddressByNATGatewayIDAndAllocationID(ctx, conn, natGatewayID, allocationID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func StatusNATGatewayAddressByNATGatewayIDAndPrivateIP(ctx context.Context, conn *ec2.EC2, natGatewayID, privateIP string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNATGatewayAddressByNATGatewayIDAndPrivateIP(ctx, conn, natGatewayID, privateIP)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

const (
	SecurityGroupStatusCreated = "Created"
)

func StatusSecurityGroup(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSecurityGroupByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, SecurityGroupStatusCreated, nil
	}
}

func StatusSubnetState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusSubnetIPv6CIDRBlockAssociationState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetIPv6CIDRBlockAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Ipv6CidrBlockState, aws.StringValue(output.Ipv6CidrBlockState.State), nil
	}
}

func StatusSubnetAssignIPv6AddressOnCreation(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.AssignIpv6AddressOnCreation)), nil
	}
}

func StatusSubnetEnableDNS64(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.EnableDns64)), nil
	}
}

func StatusSubnetEnableLniAtDeviceIndex(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatInt(aws.Int64Value(output.EnableLniAtDeviceIndex), 10), nil
	}
}

func StatusSubnetEnableResourceNameDNSAAAARecordOnLaunch(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsAAAARecord)), nil
	}
}

func StatusSubnetEnableResourceNameDNSARecordOnLaunch(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsARecord)), nil
	}
}

func StatusSubnetMapCustomerOwnedIPOnLaunch(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.MapCustomerOwnedIpOnLaunch)), nil
	}
}

func StatusSubnetMapPublicIPOnLaunch(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.MapPublicIpOnLaunch)), nil
	}
}

func StatusSubnetPrivateDNSHostnameTypeOnLaunch(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.PrivateDnsNameOptionsOnLaunch.HostnameType), nil
	}
}

func StatusVPCCIDRBlockAssociationState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, _, err := FindVPCCIDRBlockAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.CidrBlockState, aws.StringValue(output.CidrBlockState.State), nil
	}
}

func StatusVPCIPv6CIDRBlockAssociationState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, _, err := FindVPCIPv6CIDRBlockAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Ipv6CidrBlockState, aws.StringValue(output.Ipv6CidrBlockState.State), nil
	}
}

func StatusVPCPeeringConnectionActive(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindVPCPeeringConnectionByID as it maps useful status codes to NotFoundError.
		output, err := FindVPCPeeringConnection(ctx, conn, &ec2.DescribeVpcPeeringConnectionsInput{
			VpcPeeringConnectionIds: aws.StringSlice([]string{id}),
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusVPCPeeringConnectionDeleted(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCPeeringConnectionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusInternetGatewayAttachmentState(ctx context.Context, conn *ec2.EC2, internetGatewayID, vpcID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInternetGatewayAttachment(ctx, conn, internetGatewayID, vpcID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusManagedPrefixListState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindManagedPrefixListByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusNetworkInterfaceStatus(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNetworkInterfaceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func StatusNetworkInterfaceAttachmentStatus(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNetworkInterfaceAttachmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
