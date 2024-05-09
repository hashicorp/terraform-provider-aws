// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"strconv"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	ec2_sdkv2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func StatusAvailabilityZoneGroupOptInStatus(ctx context.Context, conn *ec2.EC2, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAvailabilityZoneGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.OptInStatus), nil
	}
}

func StatusCapacityReservationState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCapacityReservationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusCarrierGatewayState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCarrierGatewayByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

// StatusLocalGatewayRouteTableVPCAssociationState fetches the LocalGatewayRouteTableVpcAssociation and its State
func StatusLocalGatewayRouteTableVPCAssociationState(ctx context.Context, conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{
			LocalGatewayRouteTableVpcAssociationIds: aws.StringSlice([]string{localGatewayRouteTableVpcAssociationID}),
		}

		output, err := conn.DescribeLocalGatewayRouteTableVpcAssociationsWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		var association *ec2.LocalGatewayRouteTableVpcAssociation

		for _, outputAssociation := range output.LocalGatewayRouteTableVpcAssociations {
			if outputAssociation == nil {
				continue
			}

			if aws.StringValue(outputAssociation.LocalGatewayRouteTableVpcAssociationId) == localGatewayRouteTableVpcAssociationID {
				association = outputAssociation
				break
			}
		}

		if association == nil {
			return association, ec2.RouteTableAssociationStateCodeDisassociated, nil
		}

		return association, aws.StringValue(association.State), nil
	}
}

func StatusClientVPNEndpointState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClientVPNEndpointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusClientVPNEndpointClientConnectResponseOptionsState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClientVPNEndpointClientConnectResponseOptionsByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusClientVPNAuthorizationRule(ctx context.Context, conn *ec2.EC2, endpointID, targetNetworkCIDR, accessGroupID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClientVPNAuthorizationRuleByThreePartKey(ctx, conn, endpointID, targetNetworkCIDR, accessGroupID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusClientVPNNetworkAssociation(ctx context.Context, conn *ec2.EC2, associationID, endpointID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClientVPNNetworkAssociationByIDs(ctx, conn, associationID, endpointID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusClientVPNRoute(ctx context.Context, conn *ec2.EC2, endpointID, targetSubnetID, destinationCIDR string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClientVPNRouteByThreePartKey(ctx, conn, endpointID, targetSubnetID, destinationCIDR)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusFleetState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindFleetByID as it maps useful status codes to NotFoundError.
		output, err := FindFleet(ctx, conn, &ec2.DescribeFleetsInput{
			FleetIds: aws.StringSlice([]string{id}),
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.FleetState), nil
	}
}

func StatusImageState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindImageByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

// StatusInstanceIAMInstanceProfile fetches the Instance and its IamInstanceProfile
//
// The EC2 API accepts a name and always returns an ARN, so it is converted
// back to the name to prevent unexpected differences.
func StatusInstanceIAMInstanceProfile(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instance, err := FindInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if instance.IamInstanceProfile == nil || instance.IamInstanceProfile.Arn == nil {
			return instance, "", nil
		}

		name, err := InstanceProfileARNToName(aws.StringValue(instance.IamInstanceProfile.Arn))

		if err != nil {
			return instance, "", err
		}

		return instance, name, nil
	}
}

func StatusInstanceState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindInstanceByID as it maps useful status codes to NotFoundError.
		output, err := FindInstance(ctx, conn, &ec2.DescribeInstancesInput{
			InstanceIds: aws.StringSlice([]string{id}),
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State.Name), nil
	}
}

func StatusInstanceCapacityReservationSpecificationEquals(ctx context.Context, conn *ec2.EC2, id string, expectedValue *ec2.CapacityReservationSpecification) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.CapacityReservationSpecification, strconv.FormatBool(capacityReservationSpecificationResponsesEqual(output.CapacityReservationSpecification, expectedValue)), nil
	}
}

func StatusInstanceMaintenanceOptionsAutoRecovery(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if v := output.MaintenanceOptions; v != nil {
			return v, aws.StringValue(v.AutoRecovery), nil
		}

		return nil, "", nil
	}
}

func StatusInstanceMetadataOptionsState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.MetadataOptions == nil {
			return nil, "", nil
		}

		return output.MetadataOptions, aws.StringValue(output.MetadataOptions.State), nil
	}
}

func StatusInstanceRootBlockDeviceDeleteOnTermination(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		for _, v := range output.BlockDeviceMappings {
			if aws.StringValue(v.DeviceName) == aws.StringValue(output.RootDeviceName) && v.Ebs != nil {
				return v.Ebs, strconv.FormatBool(aws.BoolValue(v.Ebs.DeleteOnTermination)), nil
			}
		}

		return nil, "", nil
	}
}

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
	RouteStatusReady = "ready"
)

func StatusRoute(ctx context.Context, conn *ec2.EC2, routeFinder RouteFinder, routeTableID, destination string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := routeFinder(ctx, conn, routeTableID, destination)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, RouteStatusReady, nil
	}
}

const (
	RouteTableStatusReady = "ready"
)

func StatusRouteTable(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRouteTableByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, RouteTableStatusReady, nil
	}
}

func StatusRouteTableAssociationState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRouteTableAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.AssociationState == nil {
			// In ISO partitions AssociationState can be nil.
			// If the association has been found then we assume it's associated.
			state := ec2.RouteTableAssociationStateCodeAssociated

			return &ec2.RouteTableAssociationState{State: aws.String(state)}, state, nil
		}

		return output.AssociationState, aws.StringValue(output.AssociationState.State), nil
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

func StatusSpotFleetActivityStatus(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSpotFleetRequestByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ActivityStatus), nil
	}
}

func StatusSpotFleetRequestState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindSpotFleetRequestByID as it maps useful status codes to NotFoundError.
		output, err := FindSpotFleetRequest(ctx, conn, &ec2.DescribeSpotFleetRequestsInput{
			SpotFleetRequestIds: aws.StringSlice([]string{id}),
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.SpotFleetRequestState), nil
	}
}

func StatusSpotInstanceRequest(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSpotInstanceRequestByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
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

func StatusTransitGatewayState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayConnectState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayConnectByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayConnectPeerState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayConnectPeerByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayMulticastDomainState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayMulticastDomainByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayMulticastDomainAssociationState(ctx context.Context, conn *ec2.EC2, multicastDomainID, attachmentID, subnetID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayMulticastDomainAssociationByThreePartKey(ctx, conn, multicastDomainID, attachmentID, subnetID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Subnet.State), nil
	}
}

func StatusTransitGatewayPeeringAttachmentState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindTransitGatewayPeeringAttachmentByID as it maps useful status codes to NotFoundError.
		output, err := FindTransitGatewayPeeringAttachment(ctx, conn, &ec2.DescribeTransitGatewayPeeringAttachmentsInput{
			TransitGatewayAttachmentIds: aws.StringSlice([]string{id}),
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayPrefixListReferenceState(ctx context.Context, conn *ec2.EC2, transitGatewayRouteTableID string, prefixListID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayPrefixListReferenceByTwoPartKey(ctx, conn, transitGatewayRouteTableID, prefixListID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayStaticRouteState(ctx context.Context, conn *ec2.EC2, transitGatewayRouteTableID, destination string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayStaticRoute(ctx, conn, transitGatewayRouteTableID, destination)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayRouteTableState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayRouteTableByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayPolicyTableState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayPolicyTableByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayPolicyTableAssociationState(ctx context.Context, conn *ec2.EC2, transitGatewayPolicyTableID, transitGatewayAttachmentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayPolicyTableAssociationByTwoPartKey(ctx, conn, transitGatewayPolicyTableID, transitGatewayAttachmentID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayRouteTableAssociationState(ctx context.Context, conn *ec2.EC2, transitGatewayRouteTableID, transitGatewayAttachmentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayRouteTableAssociationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayRouteTablePropagationState(ctx context.Context, conn *ec2.EC2, transitGatewayRouteTableID string, transitGatewayAttachmentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayRouteTablePropagationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayVPCAttachmentState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindTransitGatewayVPCAttachmentByID as it maps useful status codes to NotFoundError.
		output, err := FindTransitGatewayVPCAttachment(ctx, conn, &ec2.DescribeTransitGatewayVpcAttachmentsInput{
			TransitGatewayAttachmentIds: aws.StringSlice([]string{id}),
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVolumeState(ctx context.Context, conn *ec2_sdkv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEBSVolumeByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func StatusVolumeAttachmentState(ctx context.Context, conn *ec2.EC2, volumeID, instanceID, deviceName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVolumeAttachment(ctx, conn, volumeID, instanceID, deviceName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVolumeModificationState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVolumeModificationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ModificationState), nil
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

func statusEIPDomainNameAttribute(ctx context.Context, conn *ec2_sdkv2.Client, allocationID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findEIPDomainNameAttributeByAllocationID(ctx, conn, allocationID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.PtrRecordUpdate == nil {
			return output, "", nil
		}

		return output, aws_sdkv2.ToString(output.PtrRecordUpdate.Status), nil
	}
}

func StatusHostState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindHostByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
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

func StatusNetworkInsightsAnalysis(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNetworkInsightsAnalysisByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
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

func StatusPlacementGroupState(ctx context.Context, conn *ec2.EC2, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindPlacementGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVPCEndpointState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCEndpointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVPCEndpointServiceStateAvailable(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindVPCEndpointServiceConfigurationByID as it maps useful status codes to NotFoundError.
		output, err := FindVPCEndpointServiceConfiguration(ctx, conn, &ec2.DescribeVpcEndpointServiceConfigurationsInput{
			ServiceIds: aws.StringSlice([]string{id}),
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ServiceState), nil
	}
}

func StatusVPCEndpointServiceStateDeleted(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCEndpointServiceConfigurationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ServiceState), nil
	}
}

const (
	VPCEndpointRouteTableAssociationStatusReady = "ready"
)

func StatusVPCEndpointRouteTableAssociation(ctx context.Context, conn *ec2.EC2, vpcEndpointID, routeTableID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		err := FindVPCEndpointRouteTableAssociationExists(ctx, conn, vpcEndpointID, routeTableID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return "", VPCEndpointRouteTableAssociationStatusReady, nil
	}
}

func StatusEBSSnapshotImport(ctx context.Context, conn *ec2_sdkv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindImportSnapshotTaskByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.SnapshotTaskDetail, aws.StringValue(output.SnapshotTaskDetail.Status), nil
	}
}

func statusVPCEndpointConnectionVPCEndpointState(ctx context.Context, conn *ec2.EC2, serviceID, vpcEndpointID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCEndpointConnectionByServiceIDAndVPCEndpointID(ctx, conn, serviceID, vpcEndpointID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.VpcEndpointState), nil
	}
}

func StatusSnapshotStorageTier(ctx context.Context, conn *ec2_sdkv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSnapshotTierStatusBySnapshotID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.StorageTier), nil
	}
}

func StatusIPAMState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusIPAMPoolState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMPoolByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusIPAMPoolCIDRState(ctx context.Context, conn *ec2.EC2, cidrBlock, poolID, poolCidrId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		if cidrBlock == "" {
			output, err := FindIPAMPoolCIDRByPoolCIDRId(ctx, conn, poolCidrId, poolID)

			if tfresource.NotFound(err) {
				return nil, "", nil
			}

			if err != nil {
				return nil, "", err
			}
			cidrBlock = aws.StringValue(output.Cidr)
		}

		output, err := FindIPAMPoolCIDRByTwoPartKey(ctx, conn, cidrBlock, poolID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

const (
	// naming mapes to the SDK constants that exist for IPAM
	IpamPoolCIDRAllocationCreateComplete = "create-complete" // nosemgrep:ci.caps2-in-const-name, ci.caps2-in-var-name, ci.caps5-in-const-name, ci.caps5-in-var-name
)

func StatusIPAMPoolCIDRAllocationState(ctx context.Context, conn *ec2.EC2, allocationID, poolID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMPoolAllocationByTwoPartKey(ctx, conn, allocationID, poolID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, IpamPoolCIDRAllocationCreateComplete, nil
	}
}

func StatusIPAMResourceDiscoveryState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMResourceDiscoveryByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusIPAMResourceDiscoveryAssociationStatus(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMResourceDiscoveryAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusIPAMScopeState(ctx context.Context, conn *ec2.EC2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMScopeByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func statusInstanceConnectEndpoint(ctx context.Context, conn *ec2_sdkv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInstanceConnectEndpointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func StatusImageBlockPublicAccessState(ctx context.Context, conn *ec2_sdkv2.Client) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindImageBlockPublicAccessState(ctx, conn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws_sdkv2.ToString(output), nil
	}
}

func StatusVerifiedAccessEndpoint(ctx context.Context, conn *ec2_sdkv2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVerifiedAccessEndpointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.Code), nil
	}
}

func statusFastSnapshotRestore(ctx context.Context, conn *ec2_sdkv2.Client, availabilityZone, snapshotID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFastSnapshotRestoreByTwoPartKey(ctx, conn, availabilityZone, snapshotID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}
