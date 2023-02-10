package ec2

import (
	"context"
	"fmt"
	"strconv"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	ec2_sdkv2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func FindAvailabilityZones(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeAvailabilityZonesInput) ([]*ec2.AvailabilityZone, error) {
	output, err := conn.DescribeAvailabilityZonesWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AvailabilityZones, nil
}

func FindAvailabilityZone(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeAvailabilityZonesInput) (*ec2.AvailabilityZone, error) {
	output, err := FindAvailabilityZones(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindAvailabilityZoneGroupByName(ctx context.Context, conn *ec2.EC2, name string) (*ec2.AvailabilityZone, error) {
	input := &ec2.DescribeAvailabilityZonesInput{
		AllAvailabilityZones: aws.Bool(true),
		Filters: BuildAttributeFilterList(map[string]string{
			"group-name": name,
		}),
	}

	output, err := FindAvailabilityZones(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	// An AZ group may contain more than one AZ.
	availabilityZone := output[0]

	// Eventual consistency check.
	if aws.StringValue(availabilityZone.GroupName) != name {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return availabilityZone, nil
}

func FindCapacityReservation(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeCapacityReservationsInput) (*ec2.CapacityReservation, error) {
	output, err := FindCapacityReservations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindCapacityReservations(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeCapacityReservationsInput) ([]*ec2.CapacityReservation, error) {
	var output []*ec2.CapacityReservation

	err := conn.DescribeCapacityReservationsPagesWithContext(ctx, input, func(page *ec2.DescribeCapacityReservationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CapacityReservations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCapacityReservationIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindCapacityReservationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.CapacityReservation, error) {
	input := &ec2.DescribeCapacityReservationsInput{
		CapacityReservationIds: aws.StringSlice([]string{id}),
	}

	output, err := FindCapacityReservation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/capacity-reservations-using.html#capacity-reservations-view.
	if state := aws.StringValue(output.State); state == ec2.CapacityReservationStateCancelled || state == ec2.CapacityReservationStateExpired {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.CapacityReservationId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindCarrierGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeCarrierGatewaysInput) (*ec2.CarrierGateway, error) {
	output, err := FindCarrierGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindCarrierGateways(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeCarrierGatewaysInput) ([]*ec2.CarrierGateway, error) {
	var output []*ec2.CarrierGateway

	err := conn.DescribeCarrierGatewaysPagesWithContext(ctx, input, func(page *ec2.DescribeCarrierGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CarrierGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCarrierGatewayIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindCarrierGatewayByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.CarrierGateway, error) {
	input := &ec2.DescribeCarrierGatewaysInput{
		CarrierGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindCarrierGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.CarrierGatewayStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.CarrierGatewayId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindClientVPNEndpoint(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeClientVpnEndpointsInput) (*ec2.ClientVpnEndpoint, error) {
	output, err := FindClientVPNEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindClientVPNEndpoints(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeClientVpnEndpointsInput) ([]*ec2.ClientVpnEndpoint, error) {
	var output []*ec2.ClientVpnEndpoint

	err := conn.DescribeClientVpnEndpointsPagesWithContext(ctx, input, func(page *ec2.DescribeClientVpnEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ClientVpnEndpoints {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindClientVPNEndpointByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ClientVpnEndpoint, error) {
	input := &ec2.DescribeClientVpnEndpointsInput{
		ClientVpnEndpointIds: aws.StringSlice([]string{id}),
	}

	output, err := FindClientVPNEndpoint(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.Status.Code); state == ec2.ClientVpnEndpointStatusCodeDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.ClientVpnEndpointId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindClientVPNEndpointClientConnectResponseOptionsByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ClientConnectResponseOptions, error) {
	output, err := FindClientVPNEndpointByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	if output.ClientConnectOptions == nil || output.ClientConnectOptions.Status == nil {
		return nil, tfresource.NewEmptyResultError(id)
	}

	return output.ClientConnectOptions, nil
}

func FindClientVPNAuthorizationRule(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeClientVpnAuthorizationRulesInput) (*ec2.AuthorizationRule, error) {
	output, err := FindClientVPNAuthorizationRules(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindClientVPNAuthorizationRules(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeClientVpnAuthorizationRulesInput) ([]*ec2.AuthorizationRule, error) {
	var output []*ec2.AuthorizationRule

	err := conn.DescribeClientVpnAuthorizationRulesPagesWithContext(ctx, input, func(page *ec2.DescribeClientVpnAuthorizationRulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AuthorizationRules {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindClientVPNAuthorizationRuleByThreePartKey(ctx context.Context, conn *ec2.EC2, endpointID, targetNetworkCIDR, accessGroupID string) (*ec2.AuthorizationRule, error) {
	filters := map[string]string{
		"destination-cidr": targetNetworkCIDR,
	}
	if accessGroupID != "" {
		filters["group-id"] = accessGroupID
	}
	input := &ec2.DescribeClientVpnAuthorizationRulesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters:             BuildAttributeFilterList(filters),
	}

	return FindClientVPNAuthorizationRule(ctx, conn, input)
}

func FindClientVPNNetworkAssociation(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeClientVpnTargetNetworksInput) (*ec2.TargetNetwork, error) {
	output, err := FindClientVPNNetworkAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindClientVPNNetworkAssociations(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeClientVpnTargetNetworksInput) ([]*ec2.TargetNetwork, error) {
	var output []*ec2.TargetNetwork

	err := conn.DescribeClientVpnTargetNetworksPagesWithContext(ctx, input, func(page *ec2.DescribeClientVpnTargetNetworksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ClientVpnTargetNetworks {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound, errCodeInvalidClientVPNAssociationIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindClientVPNNetworkAssociationByIDs(ctx context.Context, conn *ec2.EC2, associationID, endpointID string) (*ec2.TargetNetwork, error) {
	input := &ec2.DescribeClientVpnTargetNetworksInput{
		AssociationIds:      aws.StringSlice([]string{associationID}),
		ClientVpnEndpointId: aws.String(endpointID),
	}

	output, err := FindClientVPNNetworkAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.Status.Code); state == ec2.AssociationStatusCodeDisassociated {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.ClientVpnEndpointId) != endpointID || aws.StringValue(output.AssociationId) != associationID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindClientVPNRoute(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeClientVpnRoutesInput) (*ec2.ClientVpnRoute, error) {
	output, err := FindClientVPNRoutes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindClientVPNRoutes(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeClientVpnRoutesInput) ([]*ec2.ClientVpnRoute, error) {
	var output []*ec2.ClientVpnRoute

	err := conn.DescribeClientVpnRoutesPagesWithContext(ctx, input, func(page *ec2.DescribeClientVpnRoutesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Routes {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindClientVPNRouteByThreePartKey(ctx context.Context, conn *ec2.EC2, endpointID, targetSubnetID, destinationCIDR string) (*ec2.ClientVpnRoute, error) {
	input := &ec2.DescribeClientVpnRoutesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters: BuildAttributeFilterList(map[string]string{
			"destination-cidr": destinationCIDR,
			"target-subnet":    targetSubnetID,
		}),
	}

	return FindClientVPNRoute(ctx, conn, input)
}

func FindCOIPPools(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeCoipPoolsInput) ([]*ec2.CoipPool, error) {
	var output []*ec2.CoipPool

	err := conn.DescribeCoipPoolsPagesWithContext(ctx, input, func(page *ec2.DescribeCoipPoolsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CoipPools {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPoolIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindCOIPPool(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeCoipPoolsInput) (*ec2.CoipPool, error) {
	output, err := FindCOIPPools(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindEBSVolumes(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVolumesInput) ([]*ec2.Volume, error) {
	var output []*ec2.Volume

	err := conn.DescribeVolumesPagesWithContext(ctx, input, func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Volumes {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVolumeNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindEBSVolume(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVolumesInput) (*ec2.Volume, error) {
	output, err := FindEBSVolumes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindEBSVolumeByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Volume, error) {
	input := &ec2.DescribeVolumesInput{
		VolumeIds: aws.StringSlice([]string{id}),
	}

	output, err := FindEBSVolume(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.VolumeStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VolumeId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindEBSVolumeAttachment(ctx context.Context, conn *ec2.EC2, volumeID, instanceID, deviceName string) (*ec2.VolumeAttachment, error) {
	input := &ec2.DescribeVolumesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"attachment.device":      deviceName,
			"attachment.instance-id": instanceID,
		}),
		VolumeIds: aws.StringSlice([]string{volumeID}),
	}

	output, err := FindEBSVolume(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.VolumeStateAvailable || state == ec2.VolumeStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VolumeId) != volumeID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	for _, v := range output.Attachments {
		if aws.StringValue(v.State) == ec2.VolumeAttachmentStateDetached {
			continue
		}

		if aws.StringValue(v.Device) == deviceName && aws.StringValue(v.InstanceId) == instanceID {
			return v, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindEIPs(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeAddressesInput) ([]*ec2.Address, error) {
	var addresses []*ec2.Address

	output, err := conn.DescribeAddressesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAddressNotFound, errCodeInvalidAllocationIDNotFound) ||
		tfawserr.ErrMessageContains(err, errCodeAuthFailure, "does not belong to you") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	for _, v := range output.Addresses {
		if v != nil {
			addresses = append(addresses, v)
		}
	}

	return addresses, nil
}

func FindEIP(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeAddressesInput) (*ec2.Address, error) {
	output, err := FindEIPs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindEIPByAllocationID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Address, error) {
	input := &ec2.DescribeAddressesInput{
		AllocationIds: aws.StringSlice([]string{id}),
	}

	output, err := FindEIP(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.AllocationId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindEIPByAssociationID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Address, error) {
	input := &ec2.DescribeAddressesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association-id": id,
		}),
	}

	output, err := FindEIP(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.AssociationId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindEIPByPublicIP(ctx context.Context, conn *ec2.EC2, ip string) (*ec2.Address, error) {
	input := &ec2.DescribeAddressesInput{
		PublicIps: aws.StringSlice([]string{ip}),
	}

	output, err := FindEIP(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.PublicIp) != ip {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindHostByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Host, error) {
	input := &ec2.DescribeHostsInput{
		HostIds: aws.StringSlice([]string{id}),
	}

	return FindHost(ctx, conn, input)
}

func FindHostByIDAndFilters(ctx context.Context, conn *ec2.EC2, id string, filters []*ec2.Filter) (*ec2.Host, error) {
	input := &ec2.DescribeHostsInput{}

	if id != "" {
		input.HostIds = aws.StringSlice([]string{id})
	}

	if len(filters) > 0 {
		input.Filter = filters
	}

	return FindHost(ctx, conn, input)
}

func FindHost(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeHostsInput) (*ec2.Host, error) {
	output, err := conn.DescribeHostsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidHostIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Hosts) == 0 || output.Hosts[0] == nil || output.Hosts[0].HostProperties == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Hosts); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	host := output.Hosts[0]

	if state := aws.StringValue(host.State); state == ec2.AllocationStateReleased || state == ec2.AllocationStateReleasedPermanentFailure {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return host, nil
}

func FindImages(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeImagesInput) ([]*ec2.Image, error) {
	var output []*ec2.Image

	err := conn.DescribeImagesPagesWithContext(ctx, input, func(page *ec2.DescribeImagesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Images {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAMIIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindImage(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeImagesInput) (*ec2.Image, error) {
	output, err := FindImages(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindImageByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Image, error) {
	input := &ec2.DescribeImagesInput{
		ImageIds: aws.StringSlice([]string{id}),
	}

	output, err := FindImage(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.ImageStateDeregistered {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.ImageId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindImageAttribute(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeImageAttributeInput) (*ec2.DescribeImageAttributeOutput, error) {
	output, err := conn.DescribeImageAttributeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAMIIDNotFound, errCodeInvalidAMIIDUnavailable) {
		return nil, &resource.NotFoundError{
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

	return output, nil
}

func FindImageLaunchPermissionsByID(ctx context.Context, conn *ec2.EC2, id string) ([]*ec2.LaunchPermission, error) {
	input := &ec2.DescribeImageAttributeInput{
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		ImageId:   aws.String(id),
	}

	output, err := FindImageAttribute(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output.LaunchPermissions) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.LaunchPermissions, nil
}

func FindImageLaunchPermission(ctx context.Context, conn *ec2.EC2, imageID, accountID, group, organizationARN, organizationalUnitARN string) (*ec2.LaunchPermission, error) {
	output, err := FindImageLaunchPermissionsByID(ctx, conn, imageID)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if (accountID != "" && aws.StringValue(v.UserId) == accountID) ||
			(group != "" && aws.StringValue(v.Group) == group) ||
			(organizationARN != "" && aws.StringValue(v.OrganizationArn) == organizationARN) ||
			(organizationalUnitARN != "" && aws.StringValue(v.OrganizationalUnitArn) == organizationalUnitARN) {
			return v, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindInstances(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInstancesInput) ([]*ec2.Instance, error) {
	var output []*ec2.Instance

	err := conn.DescribeInstancesPagesWithContext(ctx, input, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Reservations {
			if v != nil {
				for _, v := range v.Instances {
					if v != nil {
						output = append(output, v)
					}
				}
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindInstance(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInstancesInput) (*ec2.Instance, error) {
	output, err := FindInstances(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].State == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindInstanceByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindInstance(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State.Name); state == ec2.InstanceStateNameTerminated {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.InstanceId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindInstanceCreditSpecifications(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInstanceCreditSpecificationsInput) ([]*ec2.InstanceCreditSpecification, error) {
	var output []*ec2.InstanceCreditSpecification

	err := conn.DescribeInstanceCreditSpecificationsPagesWithContext(ctx, input, func(page *ec2.DescribeInstanceCreditSpecificationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InstanceCreditSpecifications {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindInstanceCreditSpecification(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInstanceCreditSpecificationsInput) (*ec2.InstanceCreditSpecification, error) {
	output, err := FindInstanceCreditSpecifications(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindInstanceCreditSpecificationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.InstanceCreditSpecification, error) {
	input := &ec2.DescribeInstanceCreditSpecificationsInput{
		InstanceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindInstanceCreditSpecification(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.InstanceId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindInstanceTypes(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInstanceTypesInput) ([]*ec2.InstanceTypeInfo, error) {
	var output []*ec2.InstanceTypeInfo

	err := conn.DescribeInstanceTypesPagesWithContext(ctx, input, func(page *ec2.DescribeInstanceTypesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InstanceTypes {
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

func FindInstanceType(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInstanceTypesInput) (*ec2.InstanceTypeInfo, error) {
	output, err := FindInstanceTypes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindInstanceTypeByName(ctx context.Context, conn *ec2.EC2, name string) (*ec2.InstanceTypeInfo, error) {
	input := &ec2.DescribeInstanceTypesInput{
		InstanceTypes: aws.StringSlice([]string{name}),
	}

	output, err := FindInstanceType(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindInstanceTypeOfferings(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInstanceTypeOfferingsInput) ([]*ec2.InstanceTypeOffering, error) {
	var output []*ec2.InstanceTypeOffering

	err := conn.DescribeInstanceTypeOfferingsPagesWithContext(ctx, input, func(page *ec2.DescribeInstanceTypeOfferingsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InstanceTypeOfferings {
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

func FindLocalGatewayRouteTables(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLocalGatewayRouteTablesInput) ([]*ec2.LocalGatewayRouteTable, error) {
	var output []*ec2.LocalGatewayRouteTable

	err := conn.DescribeLocalGatewayRouteTablesPagesWithContext(ctx, input, func(page *ec2.DescribeLocalGatewayRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LocalGatewayRouteTables {
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

func FindLocalGatewayRouteTable(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLocalGatewayRouteTablesInput) (*ec2.LocalGatewayRouteTable, error) {
	output, err := FindLocalGatewayRouteTables(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindLocalGatewayVirtualInterfaceGroups(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput) ([]*ec2.LocalGatewayVirtualInterfaceGroup, error) {
	var output []*ec2.LocalGatewayVirtualInterfaceGroup

	err := conn.DescribeLocalGatewayVirtualInterfaceGroupsPagesWithContext(ctx, input, func(page *ec2.DescribeLocalGatewayVirtualInterfaceGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LocalGatewayVirtualInterfaceGroups {
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

func FindLocalGatewayVirtualInterfaceGroup(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput) (*ec2.LocalGatewayVirtualInterfaceGroup, error) {
	output, err := FindLocalGatewayVirtualInterfaceGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindLocalGateways(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLocalGatewaysInput) ([]*ec2.LocalGateway, error) {
	var output []*ec2.LocalGateway

	err := conn.DescribeLocalGatewaysPagesWithContext(ctx, input, func(page *ec2.DescribeLocalGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LocalGateways {
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

func FindLocalGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLocalGatewaysInput) (*ec2.LocalGateway, error) {
	output, err := FindLocalGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindNetworkACL(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkAclsInput) (*ec2.NetworkAcl, error) {
	output, err := FindNetworkACLs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNetworkACLAssociationByID(ctx context.Context, conn *ec2.EC2, associationID string) (*ec2.NetworkAclAssociation, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
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

	return nil, &resource.NotFoundError{}
}

func FindNetworkACLAssociationBySubnetID(ctx context.Context, conn *ec2.EC2, subnetID string) (*ec2.NetworkAclAssociation, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
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

	return nil, &resource.NotFoundError{}
}

func FindNetworkACLEntryByThreePartKey(ctx context.Context, conn *ec2.EC2, naclID string, egress bool, ruleNumber int) (*ec2.NetworkAclEntry, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
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

	return nil, &resource.NotFoundError{}
}

func FindNetworkInterface(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInterfacesInput) (*ec2.NetworkInterface, error) {
	output, err := FindNetworkInterfaces(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindLambdaNetworkInterfacesBySecurityGroupIDsAndFunctionName(ctx context.Context, conn *ec2.EC2, securityGroupIDs []string, functionName string) ([]*ec2.NetworkInterface, error) {
	// lambdaENIDescriptionPrefix is the common prefix used in the description for Lambda function
	// elastic network interfaces (ENI). This can be used with a function name to filter to only
	// ENIs associated with a single function.
	lambdaENIDescriptionPrefix := "AWS Lambda VPC ENI-"
	description := fmt.Sprintf("%s%s-*", lambdaENIDescriptionPrefix, functionName)

	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"interface-type": ec2.NetworkInterfaceTypeLambda,
			"description":    description,
		}),
	}
	input.Filters = append(input.Filters, NewFilter("group-id", securityGroupIDs))

	return FindNetworkInterfaces(ctx, conn, input)
}

func FindNetworkInterfacesByAttachmentInstanceOwnerIDAndDescription(ctx context.Context, conn *ec2.EC2, attachmentInstanceOwnerID, description string) ([]*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"attachment.instance-owner-id": attachmentInstanceOwnerID,
			"description":                  description,
		}),
	}

	return FindNetworkInterfaces(ctx, conn, input)
}

func FindNetworkInterfaceByAttachmentID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: BuildAttributeFilterList(map[string]string{
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
		Filters: BuildAttributeFilterList(map[string]string{
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

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("Network Interface (%s) Security Group (%s) not found", networkInterfaceID, securityGroupID),
	}
}

func FindNetworkInsightsAnalysis(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInsightsAnalysesInput) (*ec2.NetworkInsightsAnalysis, error) {
	output, err := FindNetworkInsightsAnalyses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindNetworkInsightsAnalyses(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInsightsAnalysesInput) ([]*ec2.NetworkInsightsAnalysis, error) {
	var output []*ec2.NetworkInsightsAnalysis

	err := conn.DescribeNetworkInsightsAnalysesPagesWithContext(ctx, input, func(page *ec2.DescribeNetworkInsightsAnalysesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NetworkInsightsAnalyses {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInsightsAnalysisIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNetworkInsightsAnalysisByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkInsightsAnalysis, error) {
	input := &ec2.DescribeNetworkInsightsAnalysesInput{
		NetworkInsightsAnalysisIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkInsightsAnalysis(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.NetworkInsightsAnalysisId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNetworkInsightsPath(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInsightsPathsInput) (*ec2.NetworkInsightsPath, error) {
	output, err := FindNetworkInsightsPaths(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindNetworkInsightsPaths(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNetworkInsightsPathsInput) ([]*ec2.NetworkInsightsPath, error) {
	var output []*ec2.NetworkInsightsPath

	err := conn.DescribeNetworkInsightsPathsPagesWithContext(ctx, input, func(page *ec2.DescribeNetworkInsightsPathsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NetworkInsightsPaths {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInsightsPathIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNetworkInsightsPathByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkInsightsPath, error) {
	input := &ec2.DescribeNetworkInsightsPathsInput{
		NetworkInsightsPathIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkInsightsPath(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.NetworkInsightsPathId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

// FindMainRouteTableAssociationByID returns the main route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func FindMainRouteTableAssociationByID(ctx context.Context, conn *ec2.EC2, associationID string) (*ec2.RouteTableAssociation, error) {
	association, err := FindRouteTableAssociationByID(ctx, conn, associationID)

	if err != nil {
		return nil, err
	}

	if !aws.BoolValue(association.Main) {
		return nil, &resource.NotFoundError{
			Message: fmt.Sprintf("%s is not the association with the main route table", associationID),
		}
	}

	return association, err
}

// FindMainRouteTableAssociationByVPCID returns the main route table association for the specified VPC.
// Returns NotFoundError if no route table association is found.
func FindMainRouteTableAssociationByVPCID(ctx context.Context, conn *ec2.EC2, vpcID string) (*ec2.RouteTableAssociation, error) {
	routeTable, err := FindMainRouteTableByVPCID(ctx, conn, vpcID)

	if err != nil {
		return nil, err
	}

	for _, association := range routeTable.Associations {
		if aws.BoolValue(association.Main) {
			if association.AssociationState != nil {
				if state := aws.StringValue(association.AssociationState.State); state == ec2.RouteTableAssociationStateCodeDisassociated {
					continue
				}
			}

			return association, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

// FindRouteTableAssociationByID returns the route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func FindRouteTableAssociationByID(ctx context.Context, conn *ec2.EC2, associationID string) (*ec2.RouteTableAssociation, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.route-table-association-id": associationID,
		}),
	}

	routeTable, err := FindRouteTable(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, association := range routeTable.Associations {
		if aws.StringValue(association.RouteTableAssociationId) == associationID {
			if association.AssociationState != nil {
				if state := aws.StringValue(association.AssociationState.State); state == ec2.RouteTableAssociationStateCodeDisassociated {
					return nil, &resource.NotFoundError{Message: state}
				}
			}

			return association, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

// FindMainRouteTableByVPCID returns the main route table for the specified VPC.
// Returns NotFoundError if no route table is found.
func FindMainRouteTableByVPCID(ctx context.Context, conn *ec2.EC2, vpcID string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.main": "true",
			"vpc-id":           vpcID,
		}),
	}

	return FindRouteTable(ctx, conn, input)
}

// FindRouteTableByID returns the route table corresponding to the specified identifier.
// Returns NotFoundError if no route table is found.
func FindRouteTableByID(ctx context.Context, conn *ec2.EC2, routeTableID string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		RouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	return FindRouteTable(ctx, conn, input)
}

func FindRouteTable(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeRouteTablesInput) (*ec2.RouteTable, error) {
	output, err := FindRouteTables(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindRouteTables(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeRouteTablesInput) ([]*ec2.RouteTable, error) {
	var output []*ec2.RouteTable

	err := conn.DescribeRouteTablesPagesWithContext(ctx, input, func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, table := range page.RouteTables {
			if table == nil {
				continue
			}

			output = append(output, table)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

// RouteFinder returns the route corresponding to the specified destination.
// Returns NotFoundError if no route is found.
type RouteFinder func(context.Context, *ec2.EC2, string, string) (*ec2.Route, error)

// FindRouteByIPv4Destination returns the route corresponding to the specified IPv4 destination.
// Returns NotFoundError if no route is found.
func FindRouteByIPv4Destination(ctx context.Context, conn *ec2.EC2, routeTableID, destinationCidr string) (*ec2.Route, error) {
	routeTable, err := FindRouteTableByID(ctx, conn, routeTableID)

	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if verify.CIDRBlocksEqual(aws.StringValue(route.DestinationCidrBlock), destinationCidr) {
			return route, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with IPv4 destination (%s) not found", routeTableID, destinationCidr),
	}
}

// FindRouteByIPv6Destination returns the route corresponding to the specified IPv6 destination.
// Returns NotFoundError if no route is found.
func FindRouteByIPv6Destination(ctx context.Context, conn *ec2.EC2, routeTableID, destinationIpv6Cidr string) (*ec2.Route, error) {
	routeTable, err := FindRouteTableByID(ctx, conn, routeTableID)

	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if verify.CIDRBlocksEqual(aws.StringValue(route.DestinationIpv6CidrBlock), destinationIpv6Cidr) {
			return route, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with IPv6 destination (%s) not found", routeTableID, destinationIpv6Cidr),
	}
}

// FindRouteByPrefixListIDDestination returns the route corresponding to the specified prefix list destination.
// Returns NotFoundError if no route is found.
func FindRouteByPrefixListIDDestination(ctx context.Context, conn *ec2.EC2, routeTableID, prefixListID string) (*ec2.Route, error) {
	routeTable, err := FindRouteTableByID(ctx, conn, routeTableID)
	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if aws.StringValue(route.DestinationPrefixListId) == prefixListID {
			return route, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with Prefix List ID destination (%s) not found", routeTableID, prefixListID),
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
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

// FindSecurityGroupByNameAndVPCID looks up a security group by name, VPC ID. Returns a resource.NotFoundError if not found.
func FindSecurityGroupByNameAndVPCID(ctx context.Context, conn *ec2.EC2, name, vpcID string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: BuildAttributeFilterList(
			map[string]string{
				"group-name": name,
				"vpc-id":     vpcID,
			},
		),
	}
	return FindSecurityGroup(ctx, conn, input)
}

// FindSecurityGroupByNameAndVPCIDAndOwnerID looks up a security group by name, VPC ID and owner ID. Returns a resource.NotFoundError if not found.
func FindSecurityGroupByNameAndVPCIDAndOwnerID(ctx context.Context, conn *ec2.EC2, name, vpcID, ownerID string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: BuildAttributeFilterList(
			map[string]string{
				"group-name": name,
				"vpc-id":     vpcID,
				"owner-id":   ownerID,
			},
		),
	}
	return FindSecurityGroup(ctx, conn, input)
}

// FindSecurityGroup looks up a security group using an ec2.DescribeSecurityGroupsInput. Returns a resource.NotFoundError if not found.
func FindSecurityGroup(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSecurityGroupsInput) (*ec2.SecurityGroup, error) {
	output, err := FindSecurityGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
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

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{}
	}

	return output, nil
}

func FindSecurityGroupIngressRuleByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SecurityGroupRule, error) {
	output, err := FindSecurityGroupRuleByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	if aws.BoolValue(output.IsEgress) {
		return nil, &resource.NotFoundError{}
	}

	return output, nil
}

func FindSecurityGroupRulesBySecurityGroupID(ctx context.Context, conn *ec2.EC2, id string) ([]*ec2.SecurityGroupRule, error) {
	input := &ec2.DescribeSecurityGroupRulesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"group-id": id,
		}),
	}

	return FindSecurityGroupRules(ctx, conn, input)
}

func FindSpotDatafeedSubscription(ctx context.Context, conn *ec2.EC2) (*ec2.SpotDatafeedSubscription, error) {
	input := &ec2.DescribeSpotDatafeedSubscriptionInput{}
	output, err := conn.DescribeSpotDatafeedSubscriptionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidSpotDatafeedNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SpotDatafeedSubscription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SpotDatafeedSubscription, nil
}

func FindSpotFleetInstances(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSpotFleetInstancesInput) ([]*ec2.ActiveInstance, error) {
	var output []*ec2.ActiveInstance

	err := describeSpotFleetInstancesPages(ctx, conn, input, func(page *ec2.DescribeSpotFleetInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ActiveInstances {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSpotFleetRequestIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSpotFleetRequests(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSpotFleetRequestsInput) ([]*ec2.SpotFleetRequestConfig, error) {
	var output []*ec2.SpotFleetRequestConfig

	err := conn.DescribeSpotFleetRequestsPagesWithContext(ctx, input, func(page *ec2.DescribeSpotFleetRequestsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SpotFleetRequestConfigs {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSpotFleetRequestIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSpotFleetRequest(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSpotFleetRequestsInput) (*ec2.SpotFleetRequestConfig, error) {
	output, err := FindSpotFleetRequests(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].SpotFleetRequestConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindSpotFleetRequestByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SpotFleetRequestConfig, error) {
	input := &ec2.DescribeSpotFleetRequestsInput{
		SpotFleetRequestIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSpotFleetRequest(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.SpotFleetRequestState); state == ec2.BatchStateCancelled || state == ec2.BatchStateCancelledRunning || state == ec2.BatchStateCancelledTerminating {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.SpotFleetRequestId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSpotFleetRequestHistoryRecords(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSpotFleetRequestHistoryInput) ([]*ec2.HistoryRecord, error) {
	var output []*ec2.HistoryRecord

	err := describeSpotFleetRequestHistoryPages(ctx, conn, input, func(page *ec2.DescribeSpotFleetRequestHistoryOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.HistoryRecords {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSpotFleetRequestIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSpotInstanceRequests(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSpotInstanceRequestsInput) ([]*ec2.SpotInstanceRequest, error) {
	var output []*ec2.SpotInstanceRequest

	err := conn.DescribeSpotInstanceRequestsPagesWithContext(ctx, input, func(page *ec2.DescribeSpotInstanceRequestsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SpotInstanceRequests {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSpotInstanceRequestIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSpotInstanceRequest(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSpotInstanceRequestsInput) (*ec2.SpotInstanceRequest, error) {
	output, err := FindSpotInstanceRequests(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].State == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindSpotInstanceRequestByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SpotInstanceRequest, error) {
	input := &ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSpotInstanceRequest(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.SpotInstanceStateCancelled || state == ec2.SpotInstanceStateClosed {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.SpotInstanceRequestId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
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
		return nil, &resource.NotFoundError{
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

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
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

	return nil, &resource.NotFoundError{
		LastError:   err,
		LastRequest: input,
	}
}

func FindSubnetIPv6CIDRBlockAssociationByID(ctx context.Context, conn *ec2.EC2, associationID string) (*ec2.SubnetIpv6CidrBlockAssociation, error) {
	input := &ec2.DescribeSubnetsInput{
		Filters: BuildAttributeFilterList(map[string]string{
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
				return nil, &resource.NotFoundError{Message: state}
			}

			return association, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindVolumeModifications(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVolumesModificationsInput) ([]*ec2.VolumeModification, error) {
	var output []*ec2.VolumeModification

	err := conn.DescribeVolumesModificationsPagesWithContext(ctx, input, func(page *ec2.DescribeVolumesModificationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VolumesModifications {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVolumeNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVolumeModification(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVolumesModificationsInput) (*ec2.VolumeModification, error) {
	output, err := FindVolumeModifications(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindVolumeModificationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VolumeModification, error) {
	input := &ec2.DescribeVolumesModificationsInput{
		VolumeIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVolumeModification(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.VolumeId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPCAttribute(ctx context.Context, conn *ec2.EC2, vpcID string, attribute string) (bool, error) {
	input := &ec2.DescribeVpcAttributeInput{
		Attribute: aws.String(attribute),
		VpcId:     aws.String(vpcID),
	}

	output, err := conn.DescribeVpcAttributeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
		return false, &resource.NotFoundError{
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

	var v *ec2.AttributeBooleanValue
	switch attribute {
	case ec2.VpcAttributeNameEnableDnsHostnames:
		v = output.EnableDnsHostnames
	case ec2.VpcAttributeNameEnableDnsSupport:
		v = output.EnableDnsSupport
	case ec2.VpcAttributeNameEnableNetworkAddressUsageMetrics:
		v = output.EnableNetworkAddressUsageMetrics
	default:
		return false, fmt.Errorf("unsupported VPC attribute: %s", attribute)
	}

	if v == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	return aws.BoolValue(v.Value), nil
}

func FindVPCClassicLinkEnabled(ctx context.Context, conn *ec2.EC2, vpcID string) (bool, error) {
	input := &ec2.DescribeVpcClassicLinkInput{
		VpcIds: aws.StringSlice([]string{vpcID}),
	}

	output, err := conn.DescribeVpcClassicLinkWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound, errCodeUnsupportedOperation) {
		return false, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return false, err
	}

	if output == nil || len(output.Vpcs) == 0 || output.Vpcs[0] == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Vpcs); count > 1 {
		return false, tfresource.NewTooManyResultsError(count, input)
	}

	vpc := output.Vpcs[0]

	// Eventual consistency check.
	if aws.StringValue(vpc.VpcId) != vpcID {
		return false, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return aws.BoolValue(vpc.ClassicLinkEnabled), nil
}

func FindVPCClassicLinkDNSSupported(ctx context.Context, conn *ec2.EC2, vpcID string) (bool, error) {
	input := &ec2.DescribeVpcClassicLinkDnsSupportInput{
		VpcIds: aws.StringSlice([]string{vpcID}),
	}

	output, err := conn.DescribeVpcClassicLinkDnsSupportWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound, errCodeUnsupportedOperation) ||
		tfawserr.ErrMessageContains(err, errCodeAuthFailure, "This request has been administratively disabled") {
		return false, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return false, err
	}

	if output == nil || len(output.Vpcs) == 0 || output.Vpcs[0] == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Vpcs); count > 1 {
		return false, tfresource.NewTooManyResultsError(count, input)
	}

	vpc := output.Vpcs[0]

	// Eventual consistency check.
	if aws.StringValue(vpc.VpcId) != vpcID {
		return false, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return aws.BoolValue(vpc.ClassicLinkDnsSupported), nil
}

func FindVPC(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcsInput) (*ec2.Vpc, error) {
	output, err := FindVPCs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
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
		return &resource.NotFoundError{
			LastError: fmt.Errorf("EC2 VPC (%s) DHCP Options Set (%s) Association not found", vpcID, dhcpOptionsID),
		}
	}

	return nil
}

func FindVPCCIDRBlockAssociationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VpcCidrBlockAssociation, *ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: BuildAttributeFilterList(map[string]string{
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
				return nil, nil, &resource.NotFoundError{Message: state}
			}

			return association, vpc, nil
		}
	}

	return nil, nil, &resource.NotFoundError{}
}

func FindVPCIPv6CIDRBlockAssociationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VpcIpv6CidrBlockAssociation, *ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: BuildAttributeFilterList(map[string]string{
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
				return nil, nil, &resource.NotFoundError{Message: state}
			}

			return association, vpc, nil
		}
	}

	return nil, nil, &resource.NotFoundError{}
}

func FindVPCDefaultNetworkACL(ctx context.Context, conn *ec2.EC2, id string) (*ec2.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"default": "true",
			"vpc-id":  id,
		}),
	}

	return FindNetworkACL(ctx, conn, input)
}

func FindVPCDefaultSecurityGroup(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"group-name": DefaultSecurityGroupName,
			"vpc-id":     id,
		}),
	}

	return FindSecurityGroup(ctx, conn, input)
}

func FindVPCMainRouteTable(ctx context.Context, conn *ec2.EC2, id string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.main": "true",
			"vpc-id":           id,
		}),
	}

	return FindRouteTable(ctx, conn, input)
}

func FindVPCEndpoint(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointsInput) (*ec2.VpcEndpoint, error) {
	output, err := FindVPCEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindVPCEndpoints(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointsInput) ([]*ec2.VpcEndpoint, error) {
	var output []*ec2.VpcEndpoint

	err := conn.DescribeVpcEndpointsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VpcEndpoints {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCEndpointByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VpcEndpoint, error) {
	input := &ec2.DescribeVpcEndpointsInput{
		VpcEndpointIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPCEndpoint(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == vpcEndpointStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpcEndpointId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPCConnectionNotification(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointConnectionNotificationsInput) (*ec2.ConnectionNotification, error) {
	output, err := FindVPCConnectionNotifications(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindVPCConnectionNotifications(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointConnectionNotificationsInput) ([]*ec2.ConnectionNotification, error) {
	var output []*ec2.ConnectionNotification

	err := conn.DescribeVpcEndpointConnectionNotificationsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcEndpointConnectionNotificationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ConnectionNotificationSet {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidConnectionNotification) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCConnectionNotificationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ConnectionNotification, error) {
	input := &ec2.DescribeVpcEndpointConnectionNotificationsInput{
		ConnectionNotificationId: aws.String(id),
	}

	output, err := FindVPCConnectionNotification(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.ConnectionNotificationId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPCEndpointServiceConfiguration(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointServiceConfigurationsInput) (*ec2.ServiceConfiguration, error) {
	output, err := FindVPCEndpointServiceConfigurations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindVPCEndpointServiceConfigurations(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointServiceConfigurationsInput) ([]*ec2.ServiceConfiguration, error) {
	var output []*ec2.ServiceConfiguration

	err := conn.DescribeVpcEndpointServiceConfigurationsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcEndpointServiceConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ServiceConfigurations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointServiceIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCEndpointServiceConfigurationByServiceName(ctx context.Context, conn *ec2.EC2, name string) (*ec2.ServiceConfiguration, error) {
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"service-name": name,
		}),
	}

	return FindVPCEndpointServiceConfiguration(ctx, conn, input)
}

func FindVPCEndpointServices(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointServicesInput) ([]*ec2.ServiceDetail, []string, error) {
	var serviceDetails []*ec2.ServiceDetail
	var serviceNames []string

	err := describeVPCEndpointServicesPages(ctx, conn, input, func(page *ec2.DescribeVpcEndpointServicesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ServiceDetails {
			if v != nil {
				serviceDetails = append(serviceDetails, v)
			}
		}

		for _, v := range page.ServiceNames {
			serviceNames = append(serviceNames, aws.StringValue(v))
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidServiceName) {
		return nil, nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, nil, err
	}

	return serviceDetails, serviceNames, nil
}

func FindVPCEndpointServiceConfigurationByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ServiceConfiguration, error) {
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		ServiceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPCEndpointServiceConfiguration(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.ServiceState); state == ec2.ServiceStateDeleted || state == ec2.ServiceStateFailed {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.ServiceId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPCEndpointServicePermissions(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcEndpointServicePermissionsInput) ([]*ec2.AllowedPrincipal, error) {
	var output []*ec2.AllowedPrincipal

	err := conn.DescribeVpcEndpointServicePermissionsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcEndpointServicePermissionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AllowedPrincipals {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointServiceIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindVPCEndpointServicePermissionsByID(ctx context.Context, conn *ec2.EC2, id string) ([]*ec2.AllowedPrincipal, error) {
	input := &ec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(id),
	}

	return FindVPCEndpointServicePermissions(ctx, conn, input)
}

func FindVPCEndpointServicePermissionExists(ctx context.Context, conn *ec2.EC2, serviceID, principalARN string) error {
	allowedPrincipals, err := FindVPCEndpointServicePermissionsByID(ctx, conn, serviceID)

	if err != nil {
		return err
	}

	for _, v := range allowedPrincipals {
		if aws.StringValue(v.Principal) == principalARN {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint Service (%s) Principal (%s) not found", serviceID, principalARN),
	}
}

// FindVPCEndpointRouteTableAssociationExists returns NotFoundError if no association for the specified VPC endpoint and route table IDs is found.
func FindVPCEndpointRouteTableAssociationExists(ctx context.Context, conn *ec2.EC2, vpcEndpointID string, routeTableID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(ctx, conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointRouteTableID := range vpcEndpoint.RouteTableIds {
		if aws.StringValue(vpcEndpointRouteTableID) == routeTableID {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Route Table (%s) Association not found", vpcEndpointID, routeTableID),
	}
}

// FindVPCEndpointSecurityGroupAssociationExists returns NotFoundError if no association for the specified VPC endpoint and security group IDs is found.
func FindVPCEndpointSecurityGroupAssociationExists(ctx context.Context, conn *ec2.EC2, vpcEndpointID, securityGroupID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(ctx, conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, group := range vpcEndpoint.Groups {
		if aws.StringValue(group.GroupId) == securityGroupID {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Security Group (%s) Association not found", vpcEndpointID, securityGroupID),
	}
}

// FindVPCEndpointSubnetAssociationExists returns NotFoundError if no association for the specified VPC endpoint and subnet IDs is found.
func FindVPCEndpointSubnetAssociationExists(ctx context.Context, conn *ec2.EC2, vpcEndpointID string, subnetID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(ctx, conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointSubnetID := range vpcEndpoint.SubnetIds {
		if aws.StringValue(vpcEndpointSubnetID) == subnetID {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Subnet (%s) Association not found", vpcEndpointID, subnetID),
	}
}

func FindVPCPeeringConnection(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpcPeeringConnectionsInput) (*ec2.VpcPeeringConnection, error) {
	output, err := FindVPCPeeringConnections(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
			Message:     statusCode,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpcPeeringConnectionId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

// FindVPNGatewayRoutePropagationExists returns NotFoundError if no route propagation for the specified VPN gateway is found.
func FindVPNGatewayRoutePropagationExists(ctx context.Context, conn *ec2.EC2, routeTableID, gatewayID string) error {
	routeTable, err := FindRouteTableByID(ctx, conn, routeTableID)

	if err != nil {
		return err
	}

	for _, v := range routeTable.PropagatingVgws {
		if aws.StringValue(v.GatewayId) == gatewayID {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("Route Table (%s) VPN Gateway (%s) route propagation not found", routeTableID, gatewayID),
	}
}

func FindVPNGatewayVPCAttachment(ctx context.Context, conn *ec2.EC2, vpnGatewayID, vpcID string) (*ec2.VpcAttachment, error) {
	vpnGateway, err := FindVPNGatewayByID(ctx, conn, vpnGatewayID)

	if err != nil {
		return nil, err
	}

	for _, vpcAttachment := range vpnGateway.VpcAttachments {
		if aws.StringValue(vpcAttachment.VpcId) == vpcID {
			if state := aws.StringValue(vpcAttachment.State); state == ec2.AttachmentStatusDetached {
				return nil, &resource.NotFoundError{
					Message:     state,
					LastRequest: vpcID,
				}
			}

			return vpcAttachment, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(vpcID)
}

func FindVPNGatewayByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VpnGateway, error) {
	input := &ec2.DescribeVpnGatewaysInput{
		VpnGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPNGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.VpnStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpnGatewayId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPNGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpnGatewaysInput) (*ec2.VpnGateway, error) {
	output, err := conn.DescribeVpnGatewaysWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNGatewayIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.VpnGateways) == 0 || output.VpnGateways[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.VpnGateways); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.VpnGateways[0], nil
}

func FindCustomerGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeCustomerGatewaysInput) (*ec2.CustomerGateway, error) {
	output, err := FindCustomerGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindCustomerGateways(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeCustomerGatewaysInput) ([]*ec2.CustomerGateway, error) {
	output, err := conn.DescribeCustomerGatewaysWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCustomerGatewayIDNotFound) {
		return nil, &resource.NotFoundError{
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

func FindCustomerGatewayByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.CustomerGateway, error) {
	input := &ec2.DescribeCustomerGatewaysInput{
		CustomerGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindCustomerGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == CustomerGatewayStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.CustomerGatewayId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPNConnectionByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.VpnConnection, error) {
	input := &ec2.DescribeVpnConnectionsInput{
		VpnConnectionIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPNConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.VpnStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpnConnectionId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPNConnection(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeVpnConnectionsInput) (*ec2.VpnConnection, error) {
	output, err := conn.DescribeVpnConnectionsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNConnectionIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.VpnConnections) == 0 || output.VpnConnections[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.VpnConnections); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.VpnConnections[0], nil
}

func FindVPNConnectionRouteByVPNConnectionIDAndCIDR(ctx context.Context, conn *ec2.EC2, vpnConnectionID, cidrBlock string) (*ec2.VpnStaticRoute, error) {
	input := &ec2.DescribeVpnConnectionsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"route.destination-cidr-block": cidrBlock,
			"vpn-connection-id":            vpnConnectionID,
		}),
	}

	output, err := FindVPNConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Routes {
		if aws.StringValue(v.DestinationCidrBlock) == cidrBlock && aws.StringValue(v.State) != ec2.VpnStateDeleted {
			return v, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("EC2 VPN Connection (%s) Route (%s) not found", vpnConnectionID, cidrBlock),
	}
}

func FindTrafficMirrorFilter(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTrafficMirrorFiltersInput) (*ec2.TrafficMirrorFilter, error) {
	output, err := FindTrafficMirrorFilters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTrafficMirrorFilters(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTrafficMirrorFiltersInput) ([]*ec2.TrafficMirrorFilter, error) {
	var output []*ec2.TrafficMirrorFilter

	err := conn.DescribeTrafficMirrorFiltersPagesWithContext(ctx, input, func(page *ec2.DescribeTrafficMirrorFiltersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficMirrorFilters {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorFilterIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTrafficMirrorFilterByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TrafficMirrorFilter, error) {
	input := &ec2.DescribeTrafficMirrorFiltersInput{
		TrafficMirrorFilterIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTrafficMirrorFilter(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TrafficMirrorFilterId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTrafficMirrorFilterRuleByTwoPartKey(ctx context.Context, conn *ec2.EC2, filterID, ruleID string) (*ec2.TrafficMirrorFilterRule, error) {
	output, err := FindTrafficMirrorFilterByID(ctx, conn, filterID)

	if err != nil {
		return nil, err
	}

	for _, v := range [][]*ec2.TrafficMirrorFilterRule{output.IngressFilterRules, output.EgressFilterRules} {
		for _, v := range v {
			if aws.StringValue(v.TrafficMirrorFilterRuleId) == ruleID {
				return v, nil
			}
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindTrafficMirrorSession(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTrafficMirrorSessionsInput) (*ec2.TrafficMirrorSession, error) {
	output, err := FindTrafficMirrorSessions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTrafficMirrorSessions(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTrafficMirrorSessionsInput) ([]*ec2.TrafficMirrorSession, error) {
	var output []*ec2.TrafficMirrorSession

	err := conn.DescribeTrafficMirrorSessionsPagesWithContext(ctx, input, func(page *ec2.DescribeTrafficMirrorSessionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficMirrorSessions {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorSessionIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTrafficMirrorSessionByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TrafficMirrorSession, error) {
	input := &ec2.DescribeTrafficMirrorSessionsInput{
		TrafficMirrorSessionIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTrafficMirrorSession(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TrafficMirrorSessionId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTrafficMirrorTarget(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTrafficMirrorTargetsInput) (*ec2.TrafficMirrorTarget, error) {
	output, err := FindTrafficMirrorTargets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTrafficMirrorTargets(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTrafficMirrorTargetsInput) ([]*ec2.TrafficMirrorTarget, error) {
	var output []*ec2.TrafficMirrorTarget

	err := conn.DescribeTrafficMirrorTargetsPagesWithContext(ctx, input, func(page *ec2.DescribeTrafficMirrorTargetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficMirrorTargets {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorTargetIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTrafficMirrorTargetByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TrafficMirrorTarget, error) {
	input := &ec2.DescribeTrafficMirrorTargetsInput{
		TrafficMirrorTargetIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTrafficMirrorTarget(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TrafficMirrorTargetId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewaysInput) (*ec2.TransitGateway, error) {
	output, err := FindTransitGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Options == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGateways(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewaysInput) ([]*ec2.TransitGateway, error) {
	var output []*ec2.TransitGateway

	err := conn.DescribeTransitGatewaysPagesWithContext(ctx, input, func(page *ec2.DescribeTransitGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TransitGateway, error) {
	input := &ec2.DescribeTransitGatewaysInput{
		TransitGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.TransitGatewayStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayAttachment(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayAttachmentsInput) (*ec2.TransitGatewayAttachment, error) {
	output, err := FindTransitGatewayAttachments(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayAttachments(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayAttachmentsInput) ([]*ec2.TransitGatewayAttachment, error) {
	var output []*ec2.TransitGatewayAttachment

	err := conn.DescribeTransitGatewayAttachmentsPagesWithContext(ctx, input, func(page *ec2.DescribeTransitGatewayAttachmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayAttachments {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayAttachmentByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TransitGatewayAttachment, error) {
	input := &ec2.DescribeTransitGatewayAttachmentsInput{
		TransitGatewayAttachmentIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayAttachment(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayAttachmentId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayConnect(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayConnectsInput) (*ec2.TransitGatewayConnect, error) {
	output, err := FindTransitGatewayConnects(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Options == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayConnects(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayConnectsInput) ([]*ec2.TransitGatewayConnect, error) {
	var output []*ec2.TransitGatewayConnect

	err := conn.DescribeTransitGatewayConnectsPagesWithContext(ctx, input, func(page *ec2.DescribeTransitGatewayConnectsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayConnects {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayConnectByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TransitGatewayConnect, error) {
	input := &ec2.DescribeTransitGatewayConnectsInput{
		TransitGatewayAttachmentIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayConnect(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.TransitGatewayAttachmentStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayAttachmentId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayConnectPeer(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayConnectPeersInput) (*ec2.TransitGatewayConnectPeer, error) {
	output, err := FindTransitGatewayConnectPeers(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].ConnectPeerConfiguration == nil ||
		len(output[0].ConnectPeerConfiguration.BgpConfigurations) == 0 || output[0].ConnectPeerConfiguration.BgpConfigurations[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayConnectPeers(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayConnectPeersInput) ([]*ec2.TransitGatewayConnectPeer, error) {
	var output []*ec2.TransitGatewayConnectPeer

	err := conn.DescribeTransitGatewayConnectPeersPagesWithContext(ctx, input, func(page *ec2.DescribeTransitGatewayConnectPeersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayConnectPeers {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayConnectPeerIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayConnectPeerByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TransitGatewayConnectPeer, error) {
	input := &ec2.DescribeTransitGatewayConnectPeersInput{
		TransitGatewayConnectPeerIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayConnectPeer(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.TransitGatewayConnectPeerStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayConnectPeerId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayMulticastDomain(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayMulticastDomainsInput) (*ec2.TransitGatewayMulticastDomain, error) {
	output, err := FindTransitGatewayMulticastDomains(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Options == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayMulticastDomains(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayMulticastDomainsInput) ([]*ec2.TransitGatewayMulticastDomain, error) {
	var output []*ec2.TransitGatewayMulticastDomain

	err := conn.DescribeTransitGatewayMulticastDomainsPagesWithContext(ctx, input, func(page *ec2.DescribeTransitGatewayMulticastDomainsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayMulticastDomains {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayMulticastDomainByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TransitGatewayMulticastDomain, error) {
	input := &ec2.DescribeTransitGatewayMulticastDomainsInput{
		TransitGatewayMulticastDomainIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayMulticastDomain(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.TransitGatewayMulticastDomainStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayMulticastDomainId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayMulticastDomainAssociation(ctx context.Context, conn *ec2.EC2, input *ec2.GetTransitGatewayMulticastDomainAssociationsInput) (*ec2.TransitGatewayMulticastDomainAssociation, error) {
	output, err := FindTransitGatewayMulticastDomainAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Subnet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayMulticastDomainAssociations(ctx context.Context, conn *ec2.EC2, input *ec2.GetTransitGatewayMulticastDomainAssociationsInput) ([]*ec2.TransitGatewayMulticastDomainAssociation, error) {
	var output []*ec2.TransitGatewayMulticastDomainAssociation

	err := conn.GetTransitGatewayMulticastDomainAssociationsPagesWithContext(ctx, input, func(page *ec2.GetTransitGatewayMulticastDomainAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.MulticastDomainAssociations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayMulticastDomainAssociationByThreePartKey(ctx context.Context, conn *ec2.EC2, multicastDomainID, attachmentID, subnetID string) (*ec2.TransitGatewayMulticastDomainAssociation, error) {
	input := &ec2.GetTransitGatewayMulticastDomainAssociationsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"subnet-id":                     subnetID,
			"transit-gateway-attachment-id": attachmentID,
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := FindTransitGatewayMulticastDomainAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.Subnet.State); state == ec2.TransitGatewayMulitcastDomainAssociationStateDisassociated {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayAttachmentId) != attachmentID || aws.StringValue(output.Subnet.SubnetId) != subnetID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayMulticastGroups(ctx context.Context, conn *ec2.EC2, input *ec2.SearchTransitGatewayMulticastGroupsInput) ([]*ec2.TransitGatewayMulticastGroup, error) {
	var output []*ec2.TransitGatewayMulticastGroup

	err := conn.SearchTransitGatewayMulticastGroupsPagesWithContext(ctx, input, func(page *ec2.SearchTransitGatewayMulticastGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.MulticastGroups {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayMulticastGroupMemberByThreePartKey(ctx context.Context, conn *ec2.EC2, multicastDomainID, groupIPAddress, eniID string) (*ec2.TransitGatewayMulticastGroup, error) {
	input := &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"group-ip-address": groupIPAddress,
			"is-group-member":  "true",
			"is-group-source":  "false",
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := FindTransitGatewayMulticastGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, v := range output {
		if aws.StringValue(v.NetworkInterfaceId) == eniID {
			// Eventual consistency check.
			if aws.StringValue(v.GroupIpAddress) != groupIPAddress || !aws.BoolValue(v.GroupMember) {
				return nil, &resource.NotFoundError{
					LastRequest: input,
				}
			}

			return v, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func FindTransitGatewayMulticastGroupSourceByThreePartKey(ctx context.Context, conn *ec2.EC2, multicastDomainID, groupIPAddress, eniID string) (*ec2.TransitGatewayMulticastGroup, error) {
	input := &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"group-ip-address": groupIPAddress,
			"is-group-member":  "false",
			"is-group-source":  "true",
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := FindTransitGatewayMulticastGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, v := range output {
		if aws.StringValue(v.NetworkInterfaceId) == eniID {
			// Eventual consistency check.
			if aws.StringValue(v.GroupIpAddress) != groupIPAddress || !aws.BoolValue(v.GroupSource) {
				return nil, &resource.NotFoundError{
					LastRequest: input,
				}
			}

			return v, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func FindTransitGatewayPeeringAttachment(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayPeeringAttachmentsInput) (*ec2.TransitGatewayPeeringAttachment, error) {
	output, err := FindTransitGatewayPeeringAttachments(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].AccepterTgwInfo == nil || output[0].RequesterTgwInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayPeeringAttachments(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayPeeringAttachmentsInput) ([]*ec2.TransitGatewayPeeringAttachment, error) {
	var output []*ec2.TransitGatewayPeeringAttachment

	err := conn.DescribeTransitGatewayPeeringAttachmentsPagesWithContext(ctx, input, func(page *ec2.DescribeTransitGatewayPeeringAttachmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayPeeringAttachments {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayPeeringAttachmentByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TransitGatewayPeeringAttachment, error) {
	input := &ec2.DescribeTransitGatewayPeeringAttachmentsInput{
		TransitGatewayAttachmentIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayPeeringAttachment(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/vpc/latest/tgw/tgw-vpc-attachments.html#vpc-attachment-lifecycle.
	switch state := aws.StringValue(output.State); state {
	case ec2.TransitGatewayAttachmentStateDeleted,
		ec2.TransitGatewayAttachmentStateFailed,
		ec2.TransitGatewayAttachmentStateRejected:
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayAttachmentId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayPrefixListReference(ctx context.Context, conn *ec2.EC2, input *ec2.GetTransitGatewayPrefixListReferencesInput) (*ec2.TransitGatewayPrefixListReference, error) {
	output, err := FindTransitGatewayPrefixListReferences(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayPrefixListReferences(ctx context.Context, conn *ec2.EC2, input *ec2.GetTransitGatewayPrefixListReferencesInput) ([]*ec2.TransitGatewayPrefixListReference, error) {
	var output []*ec2.TransitGatewayPrefixListReference

	err := conn.GetTransitGatewayPrefixListReferencesPagesWithContext(ctx, input, func(page *ec2.GetTransitGatewayPrefixListReferencesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayPrefixListReferences {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayPrefixListReferenceByTwoPartKey(ctx context.Context, conn *ec2.EC2, transitGatewayRouteTableID, prefixListID string) (*ec2.TransitGatewayPrefixListReference, error) {
	input := &ec2.GetTransitGatewayPrefixListReferencesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"prefix-list-id": prefixListID,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := FindTransitGatewayPrefixListReference(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.PrefixListId) != prefixListID || aws.StringValue(output.TransitGatewayRouteTableId) != transitGatewayRouteTableID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayRoute(ctx context.Context, conn *ec2.EC2, transitGatewayRouteTableID, destination string) (*ec2.TransitGatewayRoute, error) {
	input := &ec2.SearchTransitGatewayRoutesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"type": ec2.TransitGatewayRouteTypeStatic,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := conn.SearchTransitGatewayRoutesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Routes) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, route := range output.Routes {
		if route == nil {
			continue
		}

		if v := aws.StringValue(route.DestinationCidrBlock); verify.CIDRBlocksEqual(v, destination) {
			if state := aws.StringValue(route.State); state == ec2.TransitGatewayRouteStateDeleted {
				return nil, &resource.NotFoundError{
					Message:     state,
					LastRequest: input,
				}
			}

			route.DestinationCidrBlock = aws.String(verify.CanonicalCIDRBlock(v))

			return route, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindTransitGatewayPolicyTable(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayPolicyTablesInput) (*ec2.TransitGatewayPolicyTable, error) {
	output, err := FindTransitGatewayPolicyTables(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayRouteTable(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayRouteTablesInput) (*ec2.TransitGatewayRouteTable, error) {
	output, err := FindTransitGatewayRouteTables(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayPolicyTables(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayPolicyTablesInput) ([]*ec2.TransitGatewayPolicyTable, error) {
	var output []*ec2.TransitGatewayPolicyTable

	err := conn.DescribeTransitGatewayPolicyTablesPagesWithContext(ctx, input, func(page *ec2.DescribeTransitGatewayPolicyTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayPolicyTables {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayRouteTables(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayRouteTablesInput) ([]*ec2.TransitGatewayRouteTable, error) {
	var output []*ec2.TransitGatewayRouteTable

	err := conn.DescribeTransitGatewayRouteTablesPagesWithContext(ctx, input, func(page *ec2.DescribeTransitGatewayRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayRouteTables {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayPolicyTableByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TransitGatewayPolicyTable, error) {
	input := &ec2.DescribeTransitGatewayPolicyTablesInput{
		TransitGatewayPolicyTableIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayPolicyTable(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayPolicyTableId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayRouteTableByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TransitGatewayRouteTable, error) {
	input := &ec2.DescribeTransitGatewayRouteTablesInput{
		TransitGatewayRouteTableIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayRouteTable(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.TransitGatewayRouteTableStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayRouteTableId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindTransitGatewayPolicyTableAssociationByTwoPartKey(ctx context.Context, conn *ec2.EC2, transitGatewayPolicyTableID, transitGatewayAttachmentID string) (*ec2.TransitGatewayPolicyTableAssociation, error) {
	input := &ec2.GetTransitGatewayPolicyTableAssociationsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"transit-gateway-attachment-id": transitGatewayAttachmentID,
		}),
		TransitGatewayPolicyTableId: aws.String(transitGatewayPolicyTableID),
	}

	output, err := FindTransitGatewayPolicyTableAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.TransitGatewayAssociationStateDisassociated {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayAttachmentId) != transitGatewayAttachmentID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, err
}

func FindTransitGatewayRouteTableAssociationByTwoPartKey(ctx context.Context, conn *ec2.EC2, transitGatewayRouteTableID, transitGatewayAttachmentID string) (*ec2.TransitGatewayRouteTableAssociation, error) {
	input := &ec2.GetTransitGatewayRouteTableAssociationsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"transit-gateway-attachment-id": transitGatewayAttachmentID,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := FindTransitGatewayRouteTableAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.TransitGatewayAssociationStateDisassociated {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayAttachmentId) != transitGatewayAttachmentID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, err
}

func FindTransitGatewayRouteTableAssociation(ctx context.Context, conn *ec2.EC2, input *ec2.GetTransitGatewayRouteTableAssociationsInput) (*ec2.TransitGatewayRouteTableAssociation, error) {
	output, err := FindTransitGatewayRouteTableAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayPolicyTableAssociations(ctx context.Context, conn *ec2.EC2, input *ec2.GetTransitGatewayPolicyTableAssociationsInput) ([]*ec2.TransitGatewayPolicyTableAssociation, error) {
	var output []*ec2.TransitGatewayPolicyTableAssociation

	err := conn.GetTransitGatewayPolicyTableAssociationsPagesWithContext(ctx, input, func(page *ec2.GetTransitGatewayPolicyTableAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Associations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayPolicyTableAssociation(ctx context.Context, conn *ec2.EC2, input *ec2.GetTransitGatewayPolicyTableAssociationsInput) (*ec2.TransitGatewayPolicyTableAssociation, error) {
	output, err := FindTransitGatewayPolicyTableAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayRouteTableAssociations(ctx context.Context, conn *ec2.EC2, input *ec2.GetTransitGatewayRouteTableAssociationsInput) ([]*ec2.TransitGatewayRouteTableAssociation, error) {
	var output []*ec2.TransitGatewayRouteTableAssociation

	err := conn.GetTransitGatewayRouteTableAssociationsPagesWithContext(ctx, input, func(page *ec2.GetTransitGatewayRouteTableAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Associations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayRouteTablePropagationByTwoPartKey(ctx context.Context, conn *ec2.EC2, transitGatewayRouteTableID string, transitGatewayAttachmentID string) (*ec2.TransitGatewayRouteTablePropagation, error) {
	input := &ec2.GetTransitGatewayRouteTablePropagationsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"transit-gateway-attachment-id": transitGatewayAttachmentID,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := FindTransitGatewayRouteTablePropagation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.TransitGatewayPropagationStateDisabled {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayAttachmentId) != transitGatewayAttachmentID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, err
}

func FindTransitGatewayRouteTablePropagation(ctx context.Context, conn *ec2.EC2, input *ec2.GetTransitGatewayRouteTablePropagationsInput) (*ec2.TransitGatewayRouteTablePropagation, error) {
	output, err := FindTransitGatewayRouteTablePropagations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayRouteTablePropagations(ctx context.Context, conn *ec2.EC2, input *ec2.GetTransitGatewayRouteTablePropagationsInput) ([]*ec2.TransitGatewayRouteTablePropagation, error) {
	var output []*ec2.TransitGatewayRouteTablePropagation

	err := conn.GetTransitGatewayRouteTablePropagationsPagesWithContext(ctx, input, func(page *ec2.GetTransitGatewayRouteTablePropagationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayRouteTablePropagations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayVPCAttachment(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayVpcAttachmentsInput) (*ec2.TransitGatewayVpcAttachment, error) {
	output, err := FindTransitGatewayVPCAttachments(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Options == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindTransitGatewayVPCAttachments(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeTransitGatewayVpcAttachmentsInput) ([]*ec2.TransitGatewayVpcAttachment, error) {
	var output []*ec2.TransitGatewayVpcAttachment

	err := conn.DescribeTransitGatewayVpcAttachmentsPagesWithContext(ctx, input, func(page *ec2.DescribeTransitGatewayVpcAttachmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TransitGatewayVpcAttachments {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindTransitGatewayVPCAttachmentByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.TransitGatewayVpcAttachment, error) {
	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{
		TransitGatewayAttachmentIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayVPCAttachment(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/vpc/latest/tgw/tgw-vpc-attachments.html#vpc-attachment-lifecycle.
	switch state := aws.StringValue(output.State); state {
	case ec2.TransitGatewayAttachmentStateDeleted,
		ec2.TransitGatewayAttachmentStateFailed,
		ec2.TransitGatewayAttachmentStateRejected:
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.TransitGatewayAttachmentId) != id {
		return nil, &resource.NotFoundError{
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

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
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

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindFleet(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeFleetsInput) (*ec2.FleetData, error) {
	output, err := FindFleets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindFleets(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeFleetsInput) ([]*ec2.FleetData, error) {
	var output []*ec2.FleetData

	err := conn.DescribeFleetsPagesWithContext(ctx, input, func(page *ec2.DescribeFleetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Fleets {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidFleetIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindFleetByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.FleetData, error) {
	input := &ec2.DescribeFleetsInput{
		FleetIds: aws.StringSlice([]string{id}),
	}

	output, err := FindFleet(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.FleetState); state == ec2.FleetStateCodeDeleted || state == ec2.FleetStateCodeDeletedRunning || state == ec2.FleetStateCodeDeletedTerminating {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.FleetId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindFlowLogByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.FlowLog, error) {
	input := &ec2.DescribeFlowLogsInput{
		FlowLogIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeFlowLogsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.FlowLogs) == 0 || output.FlowLogs[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.FlowLogs); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.FlowLogs[0], nil
}

func FindInternetGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeInternetGatewaysInput) (*ec2.InternetGateway, error) {
	output, err := FindInternetGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
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

func FindIPAM(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeIpamsInput) (*ec2.Ipam, error) {
	output, err := FindIPAMs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindIPAMs(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeIpamsInput) ([]*ec2.Ipam, error) {
	var output []*ec2.Ipam

	err := conn.DescribeIpamsPagesWithContext(ctx, input, func(page *ec2.DescribeIpamsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Ipams {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindIPAMByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Ipam, error) {
	input := &ec2.DescribeIpamsInput{
		IpamIds: aws.StringSlice([]string{id}),
	}

	output, err := FindIPAM(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.IpamStateDeleteComplete {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.IpamId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindIPAMPool(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeIpamPoolsInput) (*ec2.IpamPool, error) {
	output, err := FindIPAMPools(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindIPAMPools(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeIpamPoolsInput) ([]*ec2.IpamPool, error) {
	var output []*ec2.IpamPool

	err := conn.DescribeIpamPoolsPagesWithContext(ctx, input, func(page *ec2.DescribeIpamPoolsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.IpamPools {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindIPAMPoolByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.IpamPool, error) {
	input := &ec2.DescribeIpamPoolsInput{
		IpamPoolIds: aws.StringSlice([]string{id}),
	}

	output, err := FindIPAMPool(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.IpamPoolStateDeleteComplete {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.IpamPoolId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindIPAMPoolAllocation(ctx context.Context, conn *ec2.EC2, input *ec2.GetIpamPoolAllocationsInput) (*ec2.IpamPoolAllocation, error) {
	output, err := FindIPAMPoolAllocations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindIPAMPoolAllocations(ctx context.Context, conn *ec2.EC2, input *ec2.GetIpamPoolAllocationsInput) ([]*ec2.IpamPoolAllocation, error) {
	var output []*ec2.IpamPoolAllocation

	err := conn.GetIpamPoolAllocationsPagesWithContext(ctx, input, func(page *ec2.GetIpamPoolAllocationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.IpamPoolAllocations {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolAllocationIdNotFound, errCodeInvalidIPAMPoolIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindIPAMPoolAllocationByTwoPartKey(ctx context.Context, conn *ec2.EC2, allocationID, poolID string) (*ec2.IpamPoolAllocation, error) {
	input := &ec2.GetIpamPoolAllocationsInput{
		IpamPoolAllocationId: aws.String(allocationID),
		IpamPoolId:           aws.String(poolID),
	}

	output, err := FindIPAMPoolAllocation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.IpamPoolAllocationId) != allocationID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindIPAMPoolCIDR(ctx context.Context, conn *ec2.EC2, input *ec2.GetIpamPoolCidrsInput) (*ec2.IpamPoolCidr, error) {
	output, err := FindIPAMPoolCIDRs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindIPAMPoolCIDRs(ctx context.Context, conn *ec2.EC2, input *ec2.GetIpamPoolCidrsInput) ([]*ec2.IpamPoolCidr, error) {
	var output []*ec2.IpamPoolCidr

	err := conn.GetIpamPoolCidrsPagesWithContext(ctx, input, func(page *ec2.GetIpamPoolCidrsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.IpamPoolCidrs {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindIPAMPoolCIDRByTwoPartKey(ctx context.Context, conn *ec2.EC2, cidrBlock, poolID string) (*ec2.IpamPoolCidr, error) {
	input := &ec2.GetIpamPoolCidrsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"cidr": cidrBlock,
		}),
		IpamPoolId: aws.String(poolID),
	}

	output, err := FindIPAMPoolCIDR(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.IpamPoolCidrStateDeprovisioned {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.Cidr) != cidrBlock {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindIPAMPoolCIDRByPoolCIDRId(ctx context.Context, conn *ec2.EC2, poolCidrId, poolID string) (*ec2.IpamPoolCidr, error) {
	input := &ec2.GetIpamPoolCidrsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"ipam-pool-cidr-id": poolCidrId,
		}),
		IpamPoolId: aws.String(poolID),
	}

	output, err := FindIPAMPoolCIDR(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check
	cidrBlock := aws.StringValue(output.Cidr)
	if cidrBlock == "" {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	if state := aws.StringValue(output.State); state == ec2.IpamPoolCidrStateDeprovisioned {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return output, nil
}

func FindIPAMScope(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeIpamScopesInput) (*ec2.IpamScope, error) {
	output, err := FindIPAMScopes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindIPAMScopes(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeIpamScopesInput) ([]*ec2.IpamScope, error) {
	var output []*ec2.IpamScope

	err := conn.DescribeIpamScopesPagesWithContext(ctx, input, func(page *ec2.DescribeIpamScopesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.IpamScopes {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMScopeIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindIPAMScopeByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.IpamScope, error) {
	input := &ec2.DescribeIpamScopesInput{
		IpamScopeIds: aws.StringSlice([]string{id}),
	}

	output, err := FindIPAMScope(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == ec2.IpamScopeStateDeleteComplete {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.IpamScopeId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindKeyPair(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeKeyPairsInput) (*ec2.KeyPairInfo, error) {
	output, err := FindKeyPairs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindKeyPairs(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeKeyPairsInput) ([]*ec2.KeyPairInfo, error) {
	output, err := conn.DescribeKeyPairsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidKeyPairNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output.KeyPairs, nil
}

func FindKeyPairByName(ctx context.Context, conn *ec2.EC2, name string) (*ec2.KeyPairInfo, error) {
	input := &ec2.DescribeKeyPairsInput{
		KeyNames: aws.StringSlice([]string{name}),
	}

	output, err := FindKeyPair(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.KeyName) != name {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindLaunchTemplate(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLaunchTemplatesInput) (*ec2.LaunchTemplate, error) {
	output, err := FindLaunchTemplates(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindLaunchTemplates(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLaunchTemplatesInput) ([]*ec2.LaunchTemplate, error) {
	var output []*ec2.LaunchTemplate

	err := conn.DescribeLaunchTemplatesPagesWithContext(ctx, input, func(page *ec2.DescribeLaunchTemplatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LaunchTemplates {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidLaunchTemplateIdMalformed, errCodeInvalidLaunchTemplateIdNotFound, errCodeInvalidLaunchTemplateNameNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindLaunchTemplateByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.LaunchTemplate, error) {
	input := &ec2.DescribeLaunchTemplatesInput{
		LaunchTemplateIds: aws.StringSlice([]string{id}),
	}

	output, err := FindLaunchTemplate(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.LaunchTemplateId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindLaunchTemplateVersion(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.LaunchTemplateVersion, error) {
	output, err := FindLaunchTemplateVersions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].LaunchTemplateData == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindLaunchTemplateVersions(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeLaunchTemplateVersionsInput) ([]*ec2.LaunchTemplateVersion, error) {
	var output []*ec2.LaunchTemplateVersion

	err := conn.DescribeLaunchTemplateVersionsPagesWithContext(ctx, input, func(page *ec2.DescribeLaunchTemplateVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LaunchTemplateVersions {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidLaunchTemplateIdNotFound, errCodeInvalidLaunchTemplateNameNotFoundException, errCodeInvalidLaunchTemplateIdVersionNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindLaunchTemplateVersionByTwoPartKey(ctx context.Context, conn *ec2.EC2, launchTemplateID, version string) (*ec2.LaunchTemplateVersion, error) {
	input := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateId: aws.String(launchTemplateID),
		Versions:         aws.StringSlice([]string{version}),
	}

	output, err := FindLaunchTemplateVersion(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.LaunchTemplateId) != launchTemplateID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindManagedPrefixList(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeManagedPrefixListsInput) (*ec2.ManagedPrefixList, error) {
	output, err := FindManagedPrefixLists(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.PrefixListId) != id {
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
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

	return nil, &resource.NotFoundError{}
}

func FindNATGateway(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeNatGatewaysInput) (*ec2.NatGateway, error) {
	output, err := FindNATGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.NatGatewayId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindPlacementGroupByName(ctx context.Context, conn *ec2.EC2, name string) (*ec2.PlacementGroup, error) {
	input := &ec2.DescribePlacementGroupsInput{
		GroupNames: aws.StringSlice([]string{name}),
	}

	output, err := conn.DescribePlacementGroupsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPlacementGroupUnknown) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PlacementGroups) == 0 || output.PlacementGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.PlacementGroups); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	placementGroup := output.PlacementGroups[0]

	if state := aws.StringValue(placementGroup.State); state == ec2.PlacementGroupStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return placementGroup, nil
}

func FindPrefixList(ctx context.Context, conn *ec2.EC2, input *ec2.DescribePrefixListsInput) (*ec2.PrefixList, error) {
	output, err := FindPrefixLists(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
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
		return nil, &resource.NotFoundError{
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
		Filters: BuildAttributeFilterList(map[string]string{
			"prefix-list-name": name,
		}),
	}

	return FindPrefixList(ctx, conn, input)
}

func FindVPCEndpointConnectionByServiceIDAndVPCEndpointID(ctx context.Context, conn *ec2.EC2, serviceID, vpcEndpointID string) (*ec2.VpcEndpointConnection, error) {
	input := &ec2.DescribeVpcEndpointConnectionsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"service-id": serviceID,
			// "InvalidFilter: The filter vpc-endpoint-id  is invalid"
			// "vpc-endpoint-id ": vpcEndpointID,
		}),
	}

	var output *ec2.VpcEndpointConnection

	err := conn.DescribeVpcEndpointConnectionsPagesWithContext(ctx, input, func(page *ec2.DescribeVpcEndpointConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VpcEndpointConnections {
			if aws.StringValue(v.VpcEndpointId) == vpcEndpointID {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if vpcEndpointState := aws.StringValue(output.VpcEndpointState); vpcEndpointState == vpcEndpointStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     vpcEndpointState,
			LastRequest: input,
		}
	}

	return output, nil
}

func FindImportSnapshotTasks(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeImportSnapshotTasksInput) ([]*ec2.ImportSnapshotTask, error) {
	var output []*ec2.ImportSnapshotTask

	err := conn.DescribeImportSnapshotTasksPagesWithContext(ctx, input, func(page *ec2.DescribeImportSnapshotTasksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ImportSnapshotTasks {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrMessageContains(err, errCodeInvalidConversionTaskIdMalformed, "not found") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindImportSnapshotTask(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeImportSnapshotTasksInput) (*ec2.ImportSnapshotTask, error) {
	output, err := FindImportSnapshotTasks(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].SnapshotTaskDetail == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindImportSnapshotTaskByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ImportSnapshotTask, error) {
	input := &ec2.DescribeImportSnapshotTasksInput{
		ImportTaskIds: aws.StringSlice([]string{id}),
	}

	output, err := FindImportSnapshotTask(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.ImportTaskId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSnapshots(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSnapshotsInput) ([]*ec2.Snapshot, error) {
	var output []*ec2.Snapshot

	err := conn.DescribeSnapshotsPagesWithContext(ctx, input, func(page *ec2.DescribeSnapshotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Snapshots {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSnapshotNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSnapshot(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSnapshotsInput) (*ec2.Snapshot, error) {
	output, err := FindSnapshots(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindSnapshotByID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.Snapshot, error) {
	input := &ec2.DescribeSnapshotsInput{
		SnapshotIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSnapshot(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.SnapshotId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSnapshotAttribute(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSnapshotAttributeInput) (*ec2.DescribeSnapshotAttributeOutput, error) {
	output, err := conn.DescribeSnapshotAttributeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSnapshotNotFound) {
		return nil, &resource.NotFoundError{
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

	return output, nil
}

func FindCreateSnapshotCreateVolumePermissionByTwoPartKey(ctx context.Context, conn *ec2.EC2, snapshotID, accountID string) (*ec2.CreateVolumePermission, error) {
	input := &ec2.DescribeSnapshotAttributeInput{
		Attribute:  aws.String(ec2.SnapshotAttributeNameCreateVolumePermission),
		SnapshotId: aws.String(snapshotID),
	}

	output, err := FindSnapshotAttribute(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.CreateVolumePermissions {
		if aws.StringValue(v.UserId) == accountID {
			return v, nil
		}
	}

	return nil, &resource.NotFoundError{LastRequest: input}
}

func FindFindSnapshotTierStatuses(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSnapshotTierStatusInput) ([]*ec2.SnapshotTierStatus, error) {
	var output []*ec2.SnapshotTierStatus

	err := conn.DescribeSnapshotTierStatusPagesWithContext(ctx, input, func(page *ec2.DescribeSnapshotTierStatusOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SnapshotTierStatuses {
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

func FindFindSnapshotTierStatus(ctx context.Context, conn *ec2.EC2, input *ec2.DescribeSnapshotTierStatusInput) (*ec2.SnapshotTierStatus, error) {
	output, err := FindFindSnapshotTierStatuses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindSnapshotTierStatusBySnapshotID(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SnapshotTierStatus, error) {
	input := &ec2.DescribeSnapshotTierStatusInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"snapshot-id": id,
		}),
	}

	output, err := FindFindSnapshotTierStatus(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.SnapshotId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNetworkPerformanceMetricSubscriptions(ctx context.Context, conn *ec2_sdkv2.Client, input *ec2_sdkv2.DescribeAwsNetworkPerformanceMetricSubscriptionsInput) ([]types.Subscription, error) {
	var output []types.Subscription
	paginator := ec2_sdkv2.NewDescribeAwsNetworkPerformanceMetricSubscriptionsPaginator(conn, input, func(o *ec2_sdkv2.DescribeAwsNetworkPerformanceMetricSubscriptionsPaginatorOptions) {
		o.Limit = 100
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Subscriptions...)
	}

	return output, nil
}

func FindNetworkPerformanceMetricSubscriptionByFourPartKey(ctx context.Context, conn *ec2_sdkv2.Client, source, destination, metric, statistic string) (*types.Subscription, error) {
	input := &ec2_sdkv2.DescribeAwsNetworkPerformanceMetricSubscriptionsInput{}

	output, err := FindNetworkPerformanceMetricSubscriptions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if aws_sdkv2.ToString(v.Source) == source && aws_sdkv2.ToString(v.Destination) == destination && string(v.Metric) == metric && string(v.Statistic) == statistic {
			return &v, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindInstanceStateById(ctx context.Context, conn *ec2.EC2, id string) (*ec2.InstanceState, error) {
	in := &ec2.DescribeInstanceStatusInput{
		InstanceIds:         aws.StringSlice([]string{id}),
		IncludeAllInstances: aws.Bool(true),
	}

	out, err := conn.DescribeInstanceStatusWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.InstanceStatuses) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	instanceState := out.InstanceStatuses[0].InstanceState

	if instanceState == nil || aws.StringValue(instanceState.Name) == ec2.InstanceStateNameTerminated {
		return nil, tfresource.NewEmptyResultError(in)
	}

	// Eventual consistency check.
	if aws.StringValue(out.InstanceStatuses[0].InstanceId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: in,
		}
	}

	return instanceState, nil
}
