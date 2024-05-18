// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusAvailabilityZoneGroupOptInStatus(ctx context.Context, conn *ec2.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAvailabilityZoneGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.OptInStatus), nil
	}
}

func statusCapacityReservationState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findCapacityReservationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusFleetState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindFleetByID as it maps useful status codes to NotFoundError.
		output, err := findFleet(ctx, conn, &ec2.DescribeFleetsInput{
			FleetIds: []string{id},
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.FleetState), nil
	}
}

func statusInstanceState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindInstanceByID as it maps useful status codes to NotFoundError.
		output, err := findInstance(ctx, conn, &ec2.DescribeInstancesInput{
			InstanceIds: []string{id},
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State.Name), nil
	}
}

// StatusInstanceIAMInstanceProfile fetches the Instance and its IamInstanceProfile
//
// The EC2 API accepts a name and always returns an ARN, so it is converted
// back to the name to prevent unexpected differences.
func statusInstanceIAMInstanceProfile(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
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

		name, err := InstanceProfileARNToName(aws.ToString(instance.IamInstanceProfile.Arn))

		if err != nil {
			return instance, "", err
		}

		return instance, name, nil
	}
}

func statusInstanceCapacityReservationSpecificationEquals(ctx context.Context, conn *ec2.Client, id string, expectedValue *types.CapacityReservationSpecification) retry.StateRefreshFunc {
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

func statusInstanceMaintenanceOptionsAutoRecovery(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
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

func statusInstanceMetadataOptionsState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
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

		return output.MetadataOptions, string(output.MetadataOptions.State), nil
	}
}

func statusInstanceRootBlockDeviceDeleteOnTermination(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
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

func statusSpotFleetActivityStatus(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findSpotFleetRequestByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ActivityStatus), nil
	}
}

func statusSpotFleetRequestState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindSpotFleetRequestByID as it maps useful status codes to NotFoundError.
		output, err := findSpotFleetRequest(ctx, conn, &ec2.DescribeSpotFleetRequestsInput{
			SpotFleetRequestIds: []string{id},
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.SpotFleetRequestState), nil
	}
}

func statusSpotInstanceRequest(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSpotInstanceRequestByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.Status.Code), nil
	}
}

func statusVolumeState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
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

func statusVolumeAttachmentState(ctx context.Context, conn *ec2.Client, volumeID, instanceID, deviceName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVolumeAttachment(ctx, conn, volumeID, instanceID, deviceName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVolumeModificationState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVolumeModificationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ModificationState), nil
	}
}

func statusVPCStateV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVPCIPv6CIDRBlockAssociationStateV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, _, err := findVPCIPv6CIDRBlockAssociationByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Ipv6CidrBlockState, string(output.Ipv6CidrBlockState.State), nil
	}
}

func statusVPCAttributeValueV2(ctx context.Context, conn *ec2.Client, id string, attribute awstypes.VpcAttributeName) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		attributeValue, err := findVPCAttributeV2(ctx, conn, id, attribute)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return attributeValue, strconv.FormatBool(attributeValue), nil
	}
}

func statusNetworkInterfaceV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findNetworkInterfaceByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusNetworkInterfaceAttachmentV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findNetworkInterfaceAttachmentByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusVPCEndpointStateV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCEndpointByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusRouteV2(ctx context.Context, conn *ec2.Client, routeFinder routeFinderV2, routeTableID, destination string) retry.StateRefreshFunc {
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

func statusRouteTableV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findRouteTableByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, RouteTableStatusReady, nil
	}
}

func statusRouteTableAssociationStateV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findRouteTableAssociationByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
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

func statusVPCEndpointServiceStateAvailableV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		// Don't call FindVPCEndpointServiceConfigurationByID as it maps useful status codes to NotFoundError.
		output, err := findVPCEndpointServiceConfigurationV2(ctx, conn, &ec2.DescribeVpcEndpointServiceConfigurationsInput{
			ServiceIds: []string{id},
		})

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ServiceState), nil
	}
}

func statusVPCEndpointServiceStateDeletedV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCEndpointServiceConfigurationByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ServiceState), nil
	}
}

func statusVPCEndpointRouteTableAssociationV2(ctx context.Context, conn *ec2.Client, vpcEndpointID, routeTableID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		err := findVPCEndpointRouteTableAssociationExistsV2(ctx, conn, vpcEndpointID, routeTableID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return "", VPCEndpointRouteTableAssociationStatusReady, nil
	}
}

func statusVPCEndpointConnectionVPCEndpointStateV2(ctx context.Context, conn *ec2.Client, serviceID, vpcEndpointID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCEndpointConnectionByServiceIDAndVPCEndpointIDV2(ctx, conn, serviceID, vpcEndpointID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.VpcEndpointState), nil
	}
}

func StatusImageState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindImageByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}
