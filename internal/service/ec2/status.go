// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

const (
	launchTemplateFoundStatus = "Found"
)

func statusAvailabilityZoneGroupOptInStatus(conn *ec2.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findAvailabilityZoneGroupByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.OptInStatus), nil
	}
}

func statusCapacityReservation(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findCapacityReservationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusCarrierGateway(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findCarrierGatewayByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusFleet(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findFleet(ctx, conn, &ec2.DescribeFleetsInput{
			FleetIds: []string{id},
		})

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.FleetState), nil
	}
}

func statusHost(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findHostByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusInstance(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findInstance(ctx, conn, &ec2.DescribeInstancesInput{
			InstanceIds: []string{id},
		})

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State.Name), nil
	}
}

func statusInstanceIAMInstanceProfile(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		instance, err := findInstanceByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if instance.IamInstanceProfile == nil || instance.IamInstanceProfile.Arn == nil {
			return instance, "", nil
		}

		name, err := instanceProfileARNToName(aws.ToString(instance.IamInstanceProfile.Arn))

		if err != nil {
			return instance, "", err
		}

		return instance, name, nil
	}
}

func statusInstanceCapacityReservationSpecificationEquals(conn *ec2.Client, id string, expectedValue *awstypes.CapacityReservationSpecification) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findInstanceByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.CapacityReservationSpecification, strconv.FormatBool(capacityReservationSpecificationResponsesEqual(output.CapacityReservationSpecification, expectedValue)), nil
	}
}

func statusInstanceMaintenanceOptionsAutoRecovery(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findInstanceByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if v := output.MaintenanceOptions; v != nil {
			return v, string(v.AutoRecovery), nil
		}

		return nil, "", nil
	}
}

func statusInstanceMetadataOptions(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findInstanceByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.MetadataOptions == nil {
			return nil, "", nil
		}

		return output.MetadataOptions, string(output.MetadataOptions.State), nil
	}
}

func statusInstanceRootBlockDeviceDeleteOnTermination(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findInstanceByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		for _, v := range output.BlockDeviceMappings {
			if aws.ToString(v.DeviceName) == aws.ToString(output.RootDeviceName) && v.Ebs != nil {
				return v.Ebs, strconv.FormatBool(aws.ToBool(v.Ebs.DeleteOnTermination)), nil
			}
		}

		return nil, "", nil
	}
}

func statusLaunchTemplate(conn *ec2.Client, id string, idIsName bool) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		var output *awstypes.LaunchTemplate
		var err error
		if idIsName {
			output, err = findLaunchTemplateByName(ctx, conn, id)
		} else {
			output, err = findLaunchTemplateByID(ctx, conn, id)
		}

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, launchTemplateFoundStatus, nil
	}
}

func statusLocalGatewayRoute(conn *ec2.Client, localGatewayRouteTableID, destinationCIDRBlock string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findLocalGatewayRouteByTwoPartKey(ctx, conn, localGatewayRouteTableID, destinationCIDRBlock)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusLocalGatewayRouteTableVPCAssociation(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findLocalGatewayRouteTableVPCAssociationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.State), nil
	}
}

func statusManagedPrefixListState(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findManagedPrefixListByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusPlacementGroup(conn *ec2.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findPlacementGroupByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

const (
	securityGroupStatusCreated = "Created"
)

func statusSecurityGroup(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSecurityGroupByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, securityGroupStatusCreated, nil
	}
}

func statusSecurityGroupVPCAssociation(conn *ec2.Client, groupID, vpcID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSecurityGroupVPCAssociationByTwoPartKey(ctx, conn, groupID, vpcID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusSpotFleetActivityStatus(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSpotFleetRequestByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ActivityStatus), nil
	}
}

func statusSpotFleetRequest(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSpotFleetRequest(ctx, conn, &ec2.DescribeSpotFleetRequestsInput{
			SpotFleetRequestIds: []string{id},
		})

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.SpotFleetRequestState), nil
	}
}

func statusSpotInstanceRequest(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSpotInstanceRequestByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status.Code), nil
	}
}

func statusSubnetState(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSubnetByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusSubnetIPv6CIDRBlockAssociationState(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSubnetIPv6CIDRBlockAssociationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Ipv6CidrBlockState, string(output.Ipv6CidrBlockState.State), nil
	}
}

func statusSubnetAssignIPv6AddressOnCreation(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSubnetByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.ToBool(output.AssignIpv6AddressOnCreation)), nil
	}
}

func statusSubnetEnableDNS64(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSubnetByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.ToBool(output.EnableDns64)), nil
	}
}

func statusSubnetEnableLniAtDeviceIndex(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSubnetByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatInt(int64(aws.ToInt32(output.EnableLniAtDeviceIndex)), 10), nil
	}
}

func statusSubnetEnableResourceNameDNSAAAARecordOnLaunch(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSubnetByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.ToBool(output.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsAAAARecord)), nil
	}
}

func statusSubnetEnableResourceNameDNSARecordOnLaunch(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSubnetByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.ToBool(output.PrivateDnsNameOptionsOnLaunch.EnableResourceNameDnsARecord)), nil
	}
}

func statusSubnetMapCustomerOwnedIPOnLaunch(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSubnetByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.ToBool(output.MapCustomerOwnedIpOnLaunch)), nil
	}
}

func statusSubnetMapPublicIPOnLaunch(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSubnetByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(aws.ToBool(output.MapPublicIpOnLaunch)), nil
	}
}

func statusSubnetPrivateDNSHostnameTypeOnLaunch(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSubnetByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.PrivateDnsNameOptionsOnLaunch.HostnameType), nil
	}
}

func statusVolume(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findEBSVolumeByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVolumeAttachment(conn *ec2.Client, volumeID, instanceID, deviceName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVolumeAttachment(ctx, conn, volumeID, instanceID, deviceName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVolumeAttachmentInstanceState(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findInstance(ctx, conn, &ec2.DescribeInstancesInput{
			InstanceIds: []string{id},
		})

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State.Name), nil
	}
}

func statusVolumeModification(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVolumeModificationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ModificationState), nil
	}
}

func statusVPC(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPCByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVPCCIDRBlockAssociationState(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, _, err := findVPCCIDRBlockAssociationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.CidrBlockState, string(output.CidrBlockState.State), nil
	}
}

func statusVPCIPv6CIDRBlockAssociation(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, _, err := findVPCIPv6CIDRBlockAssociationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Ipv6CidrBlockState, string(output.Ipv6CidrBlockState.State), nil
	}
}

func statusNetworkInterface(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findNetworkInterfaceByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusNetworkInterfaceAttachment(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findNetworkInterfaceAttachmentByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusNetworkInterfacePermission(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findNetworkInterfacePermissionByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.PermissionState.State), nil
	}
}

func statusVPCEndpoint(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPCEndpointByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strings.ToLower(string(output.State)), nil
	}
}

const (
	routeStatusReady = "ready"
)

func statusRoute(conn *ec2.Client, routeFinder routeFinder, routeTableID, destination string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := routeFinder(ctx, conn, routeTableID, destination)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, routeStatusReady, nil
	}
}

const (
	routeTableStatusReady = "ready"
)

func statusRouteTable(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findRouteTableByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, routeTableStatusReady, nil
	}
}

func statusRouteTableAssociation(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findRouteTableAssociationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.AssociationState == nil {
			// In ISO partitions AssociationState can be nil.
			// If the association has been found then we assume it's associated.
			state := awstypes.RouteTableAssociationStateCodeAssociated

			return &awstypes.RouteTableAssociationState{State: state}, string(state), nil
		}

		return output.AssociationState, string(output.AssociationState.State), nil
	}
}

func statusVPCEndpointServiceAvailable(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPCEndpointServiceConfiguration(ctx, conn, &ec2.DescribeVpcEndpointServiceConfigurationsInput{
			ServiceIds: []string{id},
		})

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ServiceState), nil
	}
}

func fetchVPCEndpointServiceDeletionStatus(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPCEndpointServiceConfigurationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ServiceState), nil
	}
}

const (
	vpcEndpointRouteTableAssociationStatusReady = "ready"
)

func statusVPCEndpointRouteTableAssociation(conn *ec2.Client, vpcEndpointID, routeTableID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		err := findVPCEndpointRouteTableAssociationExists(ctx, conn, vpcEndpointID, routeTableID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return true, vpcEndpointRouteTableAssociationStatusReady, nil
	}
}

func statusVPCEndpointConnectionVPCEndpoint(conn *ec2.Client, serviceID, vpcEndpointID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPCEndpointConnectionByServiceIDAndVPCEndpointID(ctx, conn, serviceID, vpcEndpointID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strings.ToLower(string(output.VpcEndpointState)), nil
	}
}

func statusVPCEndpointServicePrivateDNSNameConfiguration(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPCEndpointServicePrivateDNSNameConfigurationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVPCPeeringConnectionActive(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPCPeeringConnection(ctx, conn, &ec2.DescribeVpcPeeringConnectionsInput{
			VpcPeeringConnectionIds: []string{id},
		})

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.Code), nil
	}
}

func statusVPCPeeringConnectionDeleted(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPCPeeringConnectionByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.Code), nil
	}
}

func statusClientVPNEndpoint(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClientVPNEndpointByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.Code), nil
	}
}

func statusClientVPNEndpointClientConnectResponseOptions(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClientVPNEndpointClientConnectResponseOptionsByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.Code), nil
	}
}

func statusClientVPNAuthorizationRule(conn *ec2.Client, endpointID, targetNetworkCIDR, accessGroupID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClientVPNAuthorizationRuleByThreePartKey(ctx, conn, endpointID, targetNetworkCIDR, accessGroupID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.Code), nil
	}
}

func statusClientVPNNetworkAssociation(conn *ec2.Client, associationID, endpointID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClientVPNNetworkAssociationByTwoPartKey(ctx, conn, associationID, endpointID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.Code), nil
	}
}

func statusClientVPNRoute(conn *ec2.Client, endpointID, targetSubnetID, destinationCIDR string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClientVPNRouteByThreePartKey(ctx, conn, endpointID, targetSubnetID, destinationCIDR)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.Code), nil
	}
}

func statusVPNConnection(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPNConnectionByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVPNConnectionRoute(conn *ec2.Client, vpnConnectionID, cidrBlock string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPNConnectionRouteByTwoPartKey(ctx, conn, vpnConnectionID, cidrBlock)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVPNGateway(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPNGatewayByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVPNGatewayVPCAttachment(conn *ec2.Client, vpnGatewayID, vpcID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPNGatewayVPCAttachmentByTwoPartKey(ctx, conn, vpnGatewayID, vpcID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVPNConcentrator(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPNConcentratorByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.State), nil
	}
}

func statusCustomerGateway(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findCustomerGatewayByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.State), nil
	}
}

func statusInternetGatewayAttachmentState(conn *ec2.Client, internetGatewayID, vpcID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findInternetGatewayAttachment(ctx, conn, internetGatewayID, vpcID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusIPAM(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findIPAMByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusIPAMPool(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findIPAMPoolByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusIPAMPoolCIDR(conn *ec2.Client, cidrBlock, poolID, poolCIDRID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		if cidrBlock == "" {
			output, err := findIPAMPoolCIDRByPoolCIDRIDAndPoolID(ctx, conn, poolCIDRID, poolID)

			if retry.NotFound(err) {
				return nil, "", nil
			}

			if err != nil {
				return nil, "", err
			}

			cidrBlock = aws.ToString(output.Cidr)
		}

		output, err := findIPAMPoolCIDRByTwoPartKey(ctx, conn, cidrBlock, poolID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusIPAMResourceDiscovery(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findIPAMResourceDiscoveryByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusIPAMResourceDiscoveryAssociation(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findIPAMResourceDiscoveryAssociationByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusIPAMScope(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findIPAMScopeByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusIPAMResourceCIDR(conn *ec2.Client, scopeID, resourceID string, addressFamily awstypes.AddressFamily) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findIPAMResourceCIDRByThreePartKey(ctx, conn, scopeID, resourceID, addressFamily)

		if retry.NotFound(err) {
			return new(awstypes.IpamResourceCidr), string(awstypes.IpamManagementStateUnmanaged), nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ManagementState), nil
	}
}

const (
	ipamPoolCIDRAllocationsExist    = "ipam-cidr-allocations-exist"
	ipamPoolCIDRAllocationsReleased = "ipam-cidr-allocations-released"
)

func statusIPAMPoolCIDRAllocationsReleased(conn *ec2.Client, poolID, cidrBlock string, optFns ...func(*ec2.Options)) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		input := ec2.GetIpamPoolAllocationsInput{
			IpamPoolId: aws.String(poolID),
		}

		allocations, err := findIPAMPoolAllocations(ctx, conn, &input, optFns...)

		if retry.NotFound(err) {
			return poolID, ipamPoolCIDRAllocationsReleased, nil
		}

		if err != nil {
			return nil, "", err
		}

		for _, allocation := range allocations {
			allocationCIDR := aws.ToString(allocation.Cidr)

			if inttypes.CIDRBlocksOverlap(cidrBlock, allocationCIDR) {
				return allocation, ipamPoolCIDRAllocationsExist, nil
			}
		}

		return poolID, ipamPoolCIDRAllocationsReleased, nil
	}
}

func statusImage(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findImageByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusImageBlockPublicAccess(conn *ec2.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findImageBlockPublicAccessState(ctx, conn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output), nil
	}
}

func statusTransitGateway(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayAttachment(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayAttachmentByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayConnect(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayConnectByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayConnectPeer(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayConnectPeerByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayMulticastDomain(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayMulticastDomainByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayMulticastDomainAssociation(conn *ec2.Client, multicastDomainID, attachmentID, subnetID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayMulticastDomainAssociationByThreePartKey(ctx, conn, multicastDomainID, attachmentID, subnetID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Subnet.State), nil
	}
}

func statusTransitGatewayPeeringAttachment(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayPeeringAttachment(ctx, conn, &ec2.DescribeTransitGatewayPeeringAttachmentsInput{
			TransitGatewayAttachmentIds: []string{id},
		})

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayPrefixListReference(conn *ec2.Client, transitGatewayRouteTableID string, prefixListID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayPrefixListReferenceByTwoPartKey(ctx, conn, transitGatewayRouteTableID, prefixListID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayStaticRoute(conn *ec2.Client, transitGatewayRouteTableID, destination string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayStaticRoute(ctx, conn, transitGatewayRouteTableID, destination)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayRouteTable(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayRouteTableByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayPolicyTable(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayPolicyTableByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayPolicyTableAssociation(conn *ec2.Client, transitGatewayPolicyTableID, transitGatewayAttachmentID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayPolicyTableAssociationByTwoPartKey(ctx, conn, transitGatewayPolicyTableID, transitGatewayAttachmentID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayRouteTableAssociation(conn *ec2.Client, transitGatewayRouteTableID, transitGatewayAttachmentID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayRouteTableAssociationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayRouteTablePropagation(conn *ec2.Client, transitGatewayRouteTableID string, transitGatewayAttachmentID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayRouteTablePropagationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusTransitGatewayVPCAttachment(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTransitGatewayVPCAttachment(ctx, conn, &ec2.DescribeTransitGatewayVpcAttachmentsInput{
			TransitGatewayAttachmentIds: []string{id},
		})

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusEIPDomainNameAttribute(conn *ec2.Client, allocationID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findEIPDomainNameAttributeByAllocationID(ctx, conn, allocationID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.PtrRecordUpdate == nil {
			return output, "", nil
		}

		return output, aws.ToString(output.PtrRecordUpdate.Status), nil
	}
}

func statusSnapshot(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSnapshotByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusSnapshotStorageTier(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSnapshotTierStatusBySnapshotID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.StorageTier), nil
	}
}

func statusInstanceConnectEndpoint(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findInstanceConnectEndpointByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVerifiedAccessEndpoint(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVerifiedAccessEndpointByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.Code), nil
	}
}

func statusFastSnapshotRestore(conn *ec2.Client, availabilityZone, snapshotID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findFastSnapshotRestoreByTwoPartKey(ctx, conn, availabilityZone, snapshotID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusSnapshotImport(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findImportSnapshotTaskByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.SnapshotTaskDetail, aws.ToString(output.SnapshotTaskDetail.Status), nil
	}
}

func statusNATGatewayState(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findNATGatewayByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusNATGatewayAddressByNATGatewayIDAndAllocationID(conn *ec2.Client, natGatewayID, allocationID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findNATGatewayAddressByNATGatewayIDAndAllocationID(ctx, conn, natGatewayID, allocationID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusNATGatewayAddressByNATGatewayIDAndPrivateIP(conn *ec2.Client, natGatewayID, privateIP string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findNATGatewayAddressByNATGatewayIDAndPrivateIP(ctx, conn, natGatewayID, privateIP)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusNATGatewayAttachedAppliances(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findNATGatewayByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		// Check if there are any attached appliances that are not in detached state
		if v := tfslices.Filter(output.AttachedAppliances, func(v awstypes.NatGatewayAttachedAppliance) bool {
			return v.AttachmentState != awstypes.NatGatewayApplianceStateDetached
		}); len(v) > 0 {
			return output, string(v[0].AttachmentState), nil
		}

		return nil, "", nil
	}
}

func statusNetworkInsightsAnalysis(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findNetworkInsightsAnalysisByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusVPCBlockPublicAccessOptions(conn *ec2.Client) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPCBlockPublicAccessOptions(ctx, conn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVPCBlockPublicAccessExclusion(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPCBlockPublicAccessExclusionByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusRouteServer(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findRouteServerByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusRouteServerAssociation(conn *ec2.Client, routeServerID, vpcID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findRouteServerAssociationByTwoPartKey(ctx, conn, routeServerID, vpcID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusRouteServerEndpoint(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findRouteServerEndpointByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusRouteServerPeer(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findRouteServerPeerByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusRouteServerPropagation(conn *ec2.Client, routeServerID, routeTableID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findRouteServerPropagationByTwoPartKey(ctx, conn, routeServerID, routeTableID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusSecondaryNetwork(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSecondaryNetworkByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusSecondarySubnet(conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSecondarySubnetByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}
