package ec2

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func StatusAvailabilityZoneGroupOptInStatus(conn *ec2.EC2, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAvailabilityZoneGroupByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.OptInStatus), nil
	}
}

func StatusCapacityReservationState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCapacityReservationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusCarrierGatewayState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCarrierGatewayByID(conn, id)

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
func StatusLocalGatewayRouteTableVPCAssociationState(conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{
			LocalGatewayRouteTableVpcAssociationIds: aws.StringSlice([]string{localGatewayRouteTableVpcAssociationID}),
		}

		output, err := conn.DescribeLocalGatewayRouteTableVpcAssociations(input)

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

func StatusClientVPNEndpointState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClientVPNEndpointByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusClientVPNEndpointClientConnectResponseOptionsState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClientVPNEndpointClientConnectResponseOptionsByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusClientVPNAuthorizationRule(conn *ec2.EC2, endpointID, targetNetworkCIDR, accessGroupID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClientVPNAuthorizationRuleByThreePartKey(conn, endpointID, targetNetworkCIDR, accessGroupID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusClientVPNNetworkAssociation(conn *ec2.EC2, associationID, endpointID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClientVPNNetworkAssociationByIDs(conn, associationID, endpointID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusClientVPNRoute(conn *ec2.EC2, endpointID, targetSubnetID, destinationCIDR string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClientVPNRouteByThreePartKey(conn, endpointID, targetSubnetID, destinationCIDR)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusFleetState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindFleetByID as it maps useful status codes to NotFoundError.
		output, err := FindFleet(conn, &ec2.DescribeFleetsInput{
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

func StatusImageState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindImageByID(conn, id)

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
func StatusInstanceIAMInstanceProfile(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instance, err := FindInstanceByID(conn, id)

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

func StatusInstanceState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindInstanceByID as it maps useful status codes to NotFoundError.
		output, err := FindInstance(conn, &ec2.DescribeInstancesInput{
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

func StatusInstanceCapacityReservationSpecificationEquals(conn *ec2.EC2, id string, expectedValue *ec2.CapacityReservationSpecification) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInstanceByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.CapacityReservationSpecification, strconv.FormatBool(capacityReservationSpecificationResponsesEqual(output.CapacityReservationSpecification, expectedValue)), nil
	}
}

func StatusInstanceMaintenanceOptionsAutoRecovery(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInstanceByID(conn, id)

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

func StatusInstanceMetadataOptionsState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInstanceByID(conn, id)

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

func StatusInstanceRootBlockDeviceDeleteOnTermination(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInstanceByID(conn, id)

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

func StatusNATGatewayState(ctx context.Context, conn *ec2.EC2, id string) resource.StateRefreshFunc {
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

const (
	RouteStatusReady = "ready"
)

func StatusRoute(conn *ec2.EC2, routeFinder RouteFinder, routeTableID, destination string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := routeFinder(conn, routeTableID, destination)

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

func StatusRouteTable(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRouteTableByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, RouteTableStatusReady, nil
	}
}

func StatusRouteTableAssociationState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRouteTableAssociationByID(conn, id)

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

func StatusSecurityGroup(ctx context.Context, conn *ec2.EC2, id string) resource.StateRefreshFunc {
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

func StatusSpotFleetActivityStatus(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSpotFleetRequestByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ActivityStatus), nil
	}
}

func StatusSpotFleetRequestState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindSpotFleetRequestByID as it maps useful status codes to NotFoundError.
		output, err := FindSpotFleetRequest(conn, &ec2.DescribeSpotFleetRequestsInput{
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

func StatusSubnetState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusSubnetIPv6CIDRBlockAssociationState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetIPv6CIDRBlockAssociationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Ipv6CidrBlockState, aws.StringValue(output.Ipv6CidrBlockState.State), nil
	}
}

func StatusSubnetAssignIPv6AddressOnCreation(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.AssignIpv6AddressOnCreation)), nil
	}
}

func StatusSubnetEnableDNS64(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.EnableDns64)), nil
	}
}

func StatusSubnetEnableResourceNameDNSAAAARecordOnLaunch(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsAAAARecord)), nil
	}
}

func StatusSubnetEnableResourceNameDNSARecordOnLaunch(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsARecord)), nil
	}
}

func StatusSubnetMapCustomerOwnedIPOnLaunch(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.MapCustomerOwnedIpOnLaunch)), nil
	}
}

func StatusSubnetMapPublicIPOnLaunch(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.BoolValue(output.MapPublicIpOnLaunch)), nil
	}
}

func StatusSubnetPrivateDNSHostnameTypeOnLaunch(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSubnetByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.PrivateDnsNameOptionsOnLaunch.HostnameType), nil
	}
}

func StatusTransitGatewayState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayConnectState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayConnectByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayConnectPeerState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayConnectPeerByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayMulticastDomainState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayMulticastDomainByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayMulticastDomainAssociationState(conn *ec2.EC2, multicastDomainID, attachmentID, subnetID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayMulticastDomainAssociationByThreePartKey(conn, multicastDomainID, attachmentID, subnetID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Subnet.State), nil
	}
}

func StatusTransitGatewayPeeringAttachmentState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindTransitGatewayPeeringAttachmentByID as it maps useful status codes to NotFoundError.
		output, err := FindTransitGatewayPeeringAttachment(conn, &ec2.DescribeTransitGatewayPeeringAttachmentsInput{
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

func StatusTransitGatewayPrefixListReferenceState(conn *ec2.EC2, transitGatewayRouteTableID string, prefixListID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayPrefixListReferenceByTwoPartKey(conn, transitGatewayRouteTableID, prefixListID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayRouteState(conn *ec2.EC2, transitGatewayRouteTableID, destination string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayRoute(conn, transitGatewayRouteTableID, destination)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayRouteTableState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayRouteTableByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayPolicyTableState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayPolicyTableByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayPolicyTableAssociationState(conn *ec2.EC2, transitGatewayPolicyTableID, transitGatewayAttachmentID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayPolicyTableAssociationByTwoPartKey(conn, transitGatewayPolicyTableID, transitGatewayAttachmentID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayRouteTableAssociationState(conn *ec2.EC2, transitGatewayRouteTableID, transitGatewayAttachmentID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayRouteTableAssociationByTwoPartKey(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayRouteTablePropagationState(conn *ec2.EC2, transitGatewayRouteTableID string, transitGatewayAttachmentID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTransitGatewayRouteTablePropagationByTwoPartKey(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusTransitGatewayVPCAttachmentState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindTransitGatewayVPCAttachmentByID as it maps useful status codes to NotFoundError.
		output, err := FindTransitGatewayVPCAttachment(conn, &ec2.DescribeTransitGatewayVpcAttachmentsInput{
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

func StatusVolumeState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEBSVolumeByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVolumeAttachmentState(conn *ec2.EC2, volumeID, instanceID, deviceName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEBSVolumeAttachment(conn, volumeID, instanceID, deviceName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVolumeModificationState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVolumeModificationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ModificationState), nil
	}
}

func StatusVPCState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVPCAttributeValue(conn *ec2.EC2, id string, attribute string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		attributeValue, err := FindVPCAttribute(conn, id, attribute)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return attributeValue, strconv.FormatBool(attributeValue), nil
	}
}

func StatusVPCCIDRBlockAssociationState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, _, err := FindVPCCIDRBlockAssociationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.CidrBlockState, aws.StringValue(output.CidrBlockState.State), nil
	}
}

func StatusVPCIPv6CIDRBlockAssociationState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, _, err := FindVPCIPv6CIDRBlockAssociationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Ipv6CidrBlockState, aws.StringValue(output.Ipv6CidrBlockState.State), nil
	}
}

func StatusVPCPeeringConnectionActive(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindVPCPeeringConnectionByID as it maps useful status codes to NotFoundError.
		output, err := FindVPCPeeringConnection(conn, &ec2.DescribeVpcPeeringConnectionsInput{
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

func StatusVPCPeeringConnectionDeleted(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCPeeringConnectionByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.Code), nil
	}
}

func StatusVPNGatewayVPCAttachmentState(conn *ec2.EC2, vpnGatewayID, vpcID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPNGatewayVPCAttachment(conn, vpnGatewayID, vpcID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusCustomerGatewayState(ctx context.Context, conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCustomerGatewayByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVPNConnectionState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPNConnectionByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVPNConnectionRouteState(conn *ec2.EC2, vpnConnectionID, cidrBlock string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPNConnectionRouteByVPNConnectionIDAndCIDR(conn, vpnConnectionID, cidrBlock)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusHostState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindHostByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusInternetGatewayAttachmentState(conn *ec2.EC2, internetGatewayID, vpcID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInternetGatewayAttachment(conn, internetGatewayID, vpcID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusManagedPrefixListState(ctx context.Context, conn *ec2.EC2, id string) resource.StateRefreshFunc {
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

func StatusNetworkInsightsAnalysis(ctx context.Context, conn *ec2.EC2, id string) resource.StateRefreshFunc {
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

func StatusNetworkInterfaceStatus(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNetworkInterfaceByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func StatusNetworkInterfaceAttachmentStatus(ctx context.Context, conn *ec2.EC2, id string) resource.StateRefreshFunc {
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

func StatusPlacementGroupState(conn *ec2.EC2, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindPlacementGroupByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVPCEndpointState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCEndpointByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusVPCEndpointServiceStateAvailable(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindVPCEndpointServiceConfigurationByID as it maps useful status codes to NotFoundError.
		output, err := FindVPCEndpointServiceConfiguration(conn, &ec2.DescribeVpcEndpointServiceConfigurationsInput{
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

func StatusVPCEndpointServiceStateDeleted(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCEndpointServiceConfigurationByID(conn, id)

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

func StatusVPCEndpointRouteTableAssociation(conn *ec2.EC2, vpcEndpointID, routeTableID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		err := FindVPCEndpointRouteTableAssociationExists(conn, vpcEndpointID, routeTableID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return "", VPCEndpointRouteTableAssociationStatusReady, nil
	}
}

func StatusEBSSnapshotImport(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindImportSnapshotTaskByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.SnapshotTaskDetail, aws.StringValue(output.SnapshotTaskDetail.Status), nil
	}
}

func statusVPCEndpointConnectionVPCEndpointState(conn *ec2.EC2, serviceID, vpcEndpointID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCEndpointConnectionByServiceIDAndVPCEndpointID(conn, serviceID, vpcEndpointID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.VpcEndpointState), nil
	}
}

func StatusSnapshotStorageTier(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSnapshotTierStatusBySnapshotID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.StorageTier), nil
	}
}

func StatusIPAMState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusIPAMPoolState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMPoolByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusIPAMPoolCIDRState(conn *ec2.EC2, cidrBlock, poolID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMPoolCIDRByTwoPartKey(conn, cidrBlock, poolID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func StatusIPAMScopeState(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindIPAMScopeByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}
