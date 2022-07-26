package ec2

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func FindAvailabilityZones(conn *ec2.EC2, input *ec2.DescribeAvailabilityZonesInput) ([]*ec2.AvailabilityZone, error) {
	output, err := conn.DescribeAvailabilityZones(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AvailabilityZones, nil
}

func FindAvailabilityZone(conn *ec2.EC2, input *ec2.DescribeAvailabilityZonesInput) (*ec2.AvailabilityZone, error) {
	output, err := FindAvailabilityZones(conn, input)

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

func FindAvailabilityZoneGroupByName(conn *ec2.EC2, name string) (*ec2.AvailabilityZone, error) {
	input := &ec2.DescribeAvailabilityZonesInput{
		AllAvailabilityZones: aws.Bool(true),
		Filters: BuildAttributeFilterList(map[string]string{
			"group-name": name,
		}),
	}

	output, err := FindAvailabilityZones(conn, input)

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

func FindCapacityReservation(conn *ec2.EC2, input *ec2.DescribeCapacityReservationsInput) (*ec2.CapacityReservation, error) {
	output, err := FindCapacityReservations(conn, input)

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

func FindCapacityReservations(conn *ec2.EC2, input *ec2.DescribeCapacityReservationsInput) ([]*ec2.CapacityReservation, error) {
	var output []*ec2.CapacityReservation

	err := conn.DescribeCapacityReservationsPages(input, func(page *ec2.DescribeCapacityReservationsOutput, lastPage bool) bool {
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

func FindCapacityReservationByID(conn *ec2.EC2, id string) (*ec2.CapacityReservation, error) {
	input := &ec2.DescribeCapacityReservationsInput{
		CapacityReservationIds: aws.StringSlice([]string{id}),
	}

	output, err := FindCapacityReservation(conn, input)

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

// FindCarrierGatewayByID returns the carrier gateway corresponding to the specified identifier.
// Returns nil and potentially an error if no carrier gateway is found.
func FindCarrierGatewayByID(conn *ec2.EC2, id string) (*ec2.CarrierGateway, error) {
	input := &ec2.DescribeCarrierGatewaysInput{
		CarrierGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeCarrierGateways(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CarrierGateways) == 0 {
		return nil, nil
	}

	return output.CarrierGateways[0], nil
}

func FindClientVPNEndpoint(conn *ec2.EC2, input *ec2.DescribeClientVpnEndpointsInput) (*ec2.ClientVpnEndpoint, error) {
	output, err := FindClientVPNEndpoints(conn, input)

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

func FindClientVPNEndpoints(conn *ec2.EC2, input *ec2.DescribeClientVpnEndpointsInput) ([]*ec2.ClientVpnEndpoint, error) {
	var output []*ec2.ClientVpnEndpoint

	err := conn.DescribeClientVpnEndpointsPages(input, func(page *ec2.DescribeClientVpnEndpointsOutput, lastPage bool) bool {
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

func FindClientVPNEndpointByID(conn *ec2.EC2, id string) (*ec2.ClientVpnEndpoint, error) {
	input := &ec2.DescribeClientVpnEndpointsInput{
		ClientVpnEndpointIds: aws.StringSlice([]string{id}),
	}

	output, err := FindClientVPNEndpoint(conn, input)

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

func FindClientVPNEndpointClientConnectResponseOptionsByID(conn *ec2.EC2, id string) (*ec2.ClientConnectResponseOptions, error) {
	output, err := FindClientVPNEndpointByID(conn, id)

	if err != nil {
		return nil, err
	}

	if output.ClientConnectOptions == nil || output.ClientConnectOptions.Status == nil {
		return nil, tfresource.NewEmptyResultError(id)
	}

	return output.ClientConnectOptions, nil
}

func FindClientVPNAuthorizationRule(conn *ec2.EC2, input *ec2.DescribeClientVpnAuthorizationRulesInput) (*ec2.AuthorizationRule, error) {
	output, err := FindClientVPNAuthorizationRules(conn, input)

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

func FindClientVPNAuthorizationRules(conn *ec2.EC2, input *ec2.DescribeClientVpnAuthorizationRulesInput) ([]*ec2.AuthorizationRule, error) {
	var output []*ec2.AuthorizationRule

	err := conn.DescribeClientVpnAuthorizationRulesPages(input, func(page *ec2.DescribeClientVpnAuthorizationRulesOutput, lastPage bool) bool {
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

func FindClientVPNAuthorizationRuleByThreePartKey(conn *ec2.EC2, endpointID, targetNetworkCIDR, accessGroupID string) (*ec2.AuthorizationRule, error) {
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

	return FindClientVPNAuthorizationRule(conn, input)
}

func FindClientVPNNetworkAssociation(conn *ec2.EC2, input *ec2.DescribeClientVpnTargetNetworksInput) (*ec2.TargetNetwork, error) {
	output, err := FindClientVPNNetworkAssociations(conn, input)

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

func FindClientVPNNetworkAssociations(conn *ec2.EC2, input *ec2.DescribeClientVpnTargetNetworksInput) ([]*ec2.TargetNetwork, error) {
	var output []*ec2.TargetNetwork

	err := conn.DescribeClientVpnTargetNetworksPages(input, func(page *ec2.DescribeClientVpnTargetNetworksOutput, lastPage bool) bool {
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

func FindClientVPNNetworkAssociationByIDs(conn *ec2.EC2, associationID, endpointID string) (*ec2.TargetNetwork, error) {
	input := &ec2.DescribeClientVpnTargetNetworksInput{
		AssociationIds:      aws.StringSlice([]string{associationID}),
		ClientVpnEndpointId: aws.String(endpointID),
	}

	output, err := FindClientVPNNetworkAssociation(conn, input)

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

func FindClientVPNRoute(conn *ec2.EC2, input *ec2.DescribeClientVpnRoutesInput) (*ec2.ClientVpnRoute, error) {
	output, err := FindClientVPNRoutes(conn, input)

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

func FindClientVPNRoutes(conn *ec2.EC2, input *ec2.DescribeClientVpnRoutesInput) ([]*ec2.ClientVpnRoute, error) {
	var output []*ec2.ClientVpnRoute

	err := conn.DescribeClientVpnRoutesPages(input, func(page *ec2.DescribeClientVpnRoutesOutput, lastPage bool) bool {
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

func FindClientVPNRouteByThreePartKey(conn *ec2.EC2, endpointID, targetSubnetID, destinationCIDR string) (*ec2.ClientVpnRoute, error) {
	input := &ec2.DescribeClientVpnRoutesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters: BuildAttributeFilterList(map[string]string{
			"destination-cidr": destinationCIDR,
			"target-subnet":    targetSubnetID,
		}),
	}

	return FindClientVPNRoute(conn, input)
}

func FindCOIPPools(conn *ec2.EC2, input *ec2.DescribeCoipPoolsInput) ([]*ec2.CoipPool, error) {
	var output []*ec2.CoipPool

	err := conn.DescribeCoipPoolsPages(input, func(page *ec2.DescribeCoipPoolsOutput, lastPage bool) bool {
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

func FindCOIPPool(conn *ec2.EC2, input *ec2.DescribeCoipPoolsInput) (*ec2.CoipPool, error) {
	output, err := FindCOIPPools(conn, input)

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

func FindEBSVolumes(conn *ec2.EC2, input *ec2.DescribeVolumesInput) ([]*ec2.Volume, error) {
	var output []*ec2.Volume

	err := conn.DescribeVolumesPages(input, func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
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

func FindEBSVolume(conn *ec2.EC2, input *ec2.DescribeVolumesInput) (*ec2.Volume, error) {
	output, err := FindEBSVolumes(conn, input)

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

func FindEBSVolumeByID(conn *ec2.EC2, id string) (*ec2.Volume, error) {
	input := &ec2.DescribeVolumesInput{
		VolumeIds: aws.StringSlice([]string{id}),
	}

	output, err := FindEBSVolume(conn, input)

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

func FindEBSVolumeAttachment(conn *ec2.EC2, volumeID, instanceID, deviceName string) (*ec2.VolumeAttachment, error) {
	input := &ec2.DescribeVolumesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"attachment.device":      deviceName,
			"attachment.instance-id": instanceID,
		}),
		VolumeIds: aws.StringSlice([]string{volumeID}),
	}

	output, err := FindEBSVolume(conn, input)

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

func FindEIPs(conn *ec2.EC2, input *ec2.DescribeAddressesInput) ([]*ec2.Address, error) {
	var addresses []*ec2.Address

	output, err := conn.DescribeAddresses(input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAddressNotFound, errCodeInvalidAllocationIDNotFound) {
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

func FindHostByID(conn *ec2.EC2, id string) (*ec2.Host, error) {
	input := &ec2.DescribeHostsInput{
		HostIds: aws.StringSlice([]string{id}),
	}

	return FindHost(conn, input)
}

func FindHostByIDAndFilters(conn *ec2.EC2, id string, filters []*ec2.Filter) (*ec2.Host, error) {
	input := &ec2.DescribeHostsInput{}

	if id != "" {
		input.HostIds = aws.StringSlice([]string{id})
	}

	if len(filters) > 0 {
		input.Filter = filters
	}

	return FindHost(conn, input)
}

func FindHost(conn *ec2.EC2, input *ec2.DescribeHostsInput) (*ec2.Host, error) {
	output, err := conn.DescribeHosts(input)

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

func FindImage(conn *ec2.EC2, input *ec2.DescribeImagesInput) (*ec2.Image, error) {
	output, err := conn.DescribeImages(input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAMIIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Images) == 0 || output.Images[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Images); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Images[0], nil
}

func FindImageByID(conn *ec2.EC2, id string) (*ec2.Image, error) {
	input := &ec2.DescribeImagesInput{
		ImageIds: aws.StringSlice([]string{id}),
	}

	output, err := FindImage(conn, input)

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

func FindInstances(conn *ec2.EC2, input *ec2.DescribeInstancesInput) ([]*ec2.Instance, error) {
	var output []*ec2.Instance

	err := conn.DescribeInstancesPages(input, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
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

func FindInstance(conn *ec2.EC2, input *ec2.DescribeInstancesInput) (*ec2.Instance, error) {
	output, err := FindInstances(conn, input)

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

func FindInstanceByID(conn *ec2.EC2, id string) (*ec2.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindInstance(conn, input)

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

func FindInstanceCreditSpecifications(conn *ec2.EC2, input *ec2.DescribeInstanceCreditSpecificationsInput) ([]*ec2.InstanceCreditSpecification, error) {
	var output []*ec2.InstanceCreditSpecification

	err := conn.DescribeInstanceCreditSpecificationsPages(input, func(page *ec2.DescribeInstanceCreditSpecificationsOutput, lastPage bool) bool {
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

func FindInstanceCreditSpecification(conn *ec2.EC2, input *ec2.DescribeInstanceCreditSpecificationsInput) (*ec2.InstanceCreditSpecification, error) {
	output, err := FindInstanceCreditSpecifications(conn, input)

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

func FindInstanceCreditSpecificationByID(conn *ec2.EC2, id string) (*ec2.InstanceCreditSpecification, error) {
	input := &ec2.DescribeInstanceCreditSpecificationsInput{
		InstanceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindInstanceCreditSpecification(conn, input)

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

func FindInstanceTypes(conn *ec2.EC2, input *ec2.DescribeInstanceTypesInput) ([]*ec2.InstanceTypeInfo, error) {
	var output []*ec2.InstanceTypeInfo

	err := conn.DescribeInstanceTypesPages(input, func(page *ec2.DescribeInstanceTypesOutput, lastPage bool) bool {
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

func FindInstanceType(conn *ec2.EC2, input *ec2.DescribeInstanceTypesInput) (*ec2.InstanceTypeInfo, error) {
	output, err := FindInstanceTypes(conn, input)

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

func FindInstanceTypeByName(conn *ec2.EC2, name string) (*ec2.InstanceTypeInfo, error) {
	input := &ec2.DescribeInstanceTypesInput{
		InstanceTypes: aws.StringSlice([]string{name}),
	}

	output, err := FindInstanceType(conn, input)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindInstanceTypeOfferings(conn *ec2.EC2, input *ec2.DescribeInstanceTypeOfferingsInput) ([]*ec2.InstanceTypeOffering, error) {
	var output []*ec2.InstanceTypeOffering

	err := conn.DescribeInstanceTypeOfferingsPages(input, func(page *ec2.DescribeInstanceTypeOfferingsOutput, lastPage bool) bool {
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

func FindLocalGatewayRouteTables(conn *ec2.EC2, input *ec2.DescribeLocalGatewayRouteTablesInput) ([]*ec2.LocalGatewayRouteTable, error) {
	var output []*ec2.LocalGatewayRouteTable

	err := conn.DescribeLocalGatewayRouteTablesPages(input, func(page *ec2.DescribeLocalGatewayRouteTablesOutput, lastPage bool) bool {
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

func FindLocalGatewayRouteTable(conn *ec2.EC2, input *ec2.DescribeLocalGatewayRouteTablesInput) (*ec2.LocalGatewayRouteTable, error) {
	output, err := FindLocalGatewayRouteTables(conn, input)

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

func FindLocalGatewayVirtualInterfaceGroups(conn *ec2.EC2, input *ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput) ([]*ec2.LocalGatewayVirtualInterfaceGroup, error) {
	var output []*ec2.LocalGatewayVirtualInterfaceGroup

	err := conn.DescribeLocalGatewayVirtualInterfaceGroupsPages(input, func(page *ec2.DescribeLocalGatewayVirtualInterfaceGroupsOutput, lastPage bool) bool {
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

func FindLocalGatewayVirtualInterfaceGroup(conn *ec2.EC2, input *ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput) (*ec2.LocalGatewayVirtualInterfaceGroup, error) {
	output, err := FindLocalGatewayVirtualInterfaceGroups(conn, input)

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

func FindLocalGateways(conn *ec2.EC2, input *ec2.DescribeLocalGatewaysInput) ([]*ec2.LocalGateway, error) {
	var output []*ec2.LocalGateway

	err := conn.DescribeLocalGatewaysPages(input, func(page *ec2.DescribeLocalGatewaysOutput, lastPage bool) bool {
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

func FindLocalGateway(conn *ec2.EC2, input *ec2.DescribeLocalGatewaysInput) (*ec2.LocalGateway, error) {
	output, err := FindLocalGateways(conn, input)

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

func FindNetworkACL(conn *ec2.EC2, input *ec2.DescribeNetworkAclsInput) (*ec2.NetworkAcl, error) {
	output, err := FindNetworkACLs(conn, input)

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

func FindNetworkACLs(conn *ec2.EC2, input *ec2.DescribeNetworkAclsInput) ([]*ec2.NetworkAcl, error) {
	var output []*ec2.NetworkAcl

	err := conn.DescribeNetworkAclsPages(input, func(page *ec2.DescribeNetworkAclsOutput, lastPage bool) bool {
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

func FindNetworkACLByID(conn *ec2.EC2, id string) (*ec2.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		NetworkAclIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkACL(conn, input)

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

func FindNetworkACLAssociationByID(conn *ec2.EC2, associationID string) (*ec2.NetworkAclAssociation, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.association-id": associationID,
		}),
	}

	output, err := FindNetworkACL(conn, input)

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

func FindNetworkACLAssociationBySubnetID(conn *ec2.EC2, subnetID string) (*ec2.NetworkAclAssociation, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.subnet-id": subnetID,
		}),
	}

	output, err := FindNetworkACL(conn, input)

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

func FindNetworkACLEntryByThreePartKey(conn *ec2.EC2, naclID string, egress bool, ruleNumber int) (*ec2.NetworkAclEntry, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"entry.egress":      strconv.FormatBool(egress),
			"entry.rule-number": strconv.Itoa(ruleNumber),
		}),
		NetworkAclIds: aws.StringSlice([]string{naclID}),
	}

	output, err := FindNetworkACL(conn, input)

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

func FindNetworkInterface(conn *ec2.EC2, input *ec2.DescribeNetworkInterfacesInput) (*ec2.NetworkInterface, error) {
	output, err := FindNetworkInterfaces(conn, input)

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

func FindNetworkInterfaces(conn *ec2.EC2, input *ec2.DescribeNetworkInterfacesInput) ([]*ec2.NetworkInterface, error) {
	var output []*ec2.NetworkInterface

	err := conn.DescribeNetworkInterfacesPages(input, func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
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

func FindNetworkInterfaceByID(conn *ec2.EC2, id string) (*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkInterface(conn, input)

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

func FindNetworkInterfacesByAttachmentInstanceOwnerIDAndDescription(conn *ec2.EC2, attachmentInstanceOwnerID, description string) ([]*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"attachment.instance-owner-id": attachmentInstanceOwnerID,
			"description":                  description,
		}),
	}

	return FindNetworkInterfaces(conn, input)
}

func FindNetworkInterfaceAttachmentByID(conn *ec2.EC2, id string) (*ec2.NetworkInterfaceAttachment, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"attachment.attachment-id": id,
		}),
	}

	networkInterface, err := FindNetworkInterface(conn, input)

	if err != nil {
		return nil, err
	}

	if networkInterface.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return networkInterface.Attachment, nil
}

func FindNetworkInterfaceSecurityGroup(conn *ec2.EC2, networkInterfaceID string, securityGroupID string) (*ec2.GroupIdentifier, error) {
	networkInterface, err := FindNetworkInterfaceByID(conn, networkInterfaceID)

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

func FindNetworkInsightsPath(conn *ec2.EC2, input *ec2.DescribeNetworkInsightsPathsInput) (*ec2.NetworkInsightsPath, error) {
	output, err := FindNetworkInsightsPaths(conn, input)

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

func FindNetworkInsightsPaths(conn *ec2.EC2, input *ec2.DescribeNetworkInsightsPathsInput) ([]*ec2.NetworkInsightsPath, error) {
	var output []*ec2.NetworkInsightsPath

	err := conn.DescribeNetworkInsightsPathsPages(input, func(page *ec2.DescribeNetworkInsightsPathsOutput, lastPage bool) bool {
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

func FindNetworkInsightsPathByID(conn *ec2.EC2, id string) (*ec2.NetworkInsightsPath, error) {
	input := &ec2.DescribeNetworkInsightsPathsInput{
		NetworkInsightsPathIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkInsightsPath(conn, input)

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
func FindMainRouteTableAssociationByID(conn *ec2.EC2, associationID string) (*ec2.RouteTableAssociation, error) {
	association, err := FindRouteTableAssociationByID(conn, associationID)

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
func FindMainRouteTableAssociationByVPCID(conn *ec2.EC2, vpcID string) (*ec2.RouteTableAssociation, error) {
	routeTable, err := FindMainRouteTableByVPCID(conn, vpcID)

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
func FindRouteTableAssociationByID(conn *ec2.EC2, associationID string) (*ec2.RouteTableAssociation, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.route-table-association-id": associationID,
		}),
	}

	routeTable, err := FindRouteTable(conn, input)

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
func FindMainRouteTableByVPCID(conn *ec2.EC2, vpcID string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.main": "true",
			"vpc-id":           vpcID,
		}),
	}

	return FindRouteTable(conn, input)
}

// FindRouteTableByID returns the route table corresponding to the specified identifier.
// Returns NotFoundError if no route table is found.
func FindRouteTableByID(conn *ec2.EC2, routeTableID string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		RouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	return FindRouteTable(conn, input)
}

func FindRouteTable(conn *ec2.EC2, input *ec2.DescribeRouteTablesInput) (*ec2.RouteTable, error) {
	output, err := FindRouteTables(conn, input)

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

func FindRouteTables(conn *ec2.EC2, input *ec2.DescribeRouteTablesInput) ([]*ec2.RouteTable, error) {
	var output []*ec2.RouteTable

	err := conn.DescribeRouteTablesPages(input, func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
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
type RouteFinder func(*ec2.EC2, string, string) (*ec2.Route, error)

// FindRouteByIPv4Destination returns the route corresponding to the specified IPv4 destination.
// Returns NotFoundError if no route is found.
func FindRouteByIPv4Destination(conn *ec2.EC2, routeTableID, destinationCidr string) (*ec2.Route, error) {
	routeTable, err := FindRouteTableByID(conn, routeTableID)

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
func FindRouteByIPv6Destination(conn *ec2.EC2, routeTableID, destinationIpv6Cidr string) (*ec2.Route, error) {
	routeTable, err := FindRouteTableByID(conn, routeTableID)

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
func FindRouteByPrefixListIDDestination(conn *ec2.EC2, routeTableID, prefixListID string) (*ec2.Route, error) {
	routeTable, err := FindRouteTableByID(conn, routeTableID)
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

func FindSecurityGroupByID(conn *ec2.EC2, id string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSecurityGroup(conn, input)

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

// FindSecurityGroupByNameAndVPCID looks up a security group by name and VPC ID. Returns a resource.NotFoundError if not found.
func FindSecurityGroupByNameAndVPCID(conn *ec2.EC2, name, vpcID string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: BuildAttributeFilterList(
			map[string]string{
				"group-name": name,
				"vpc-id":     vpcID,
			},
		),
	}
	return FindSecurityGroup(conn, input)
}

// FindSecurityGroup looks up a security group using an ec2.DescribeSecurityGroupsInput. Returns a resource.NotFoundError if not found.
func FindSecurityGroup(conn *ec2.EC2, input *ec2.DescribeSecurityGroupsInput) (*ec2.SecurityGroup, error) {
	output, err := FindSecurityGroups(conn, input)

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

func FindSecurityGroups(conn *ec2.EC2, input *ec2.DescribeSecurityGroupsInput) ([]*ec2.SecurityGroup, error) {
	var output []*ec2.SecurityGroup

	err := conn.DescribeSecurityGroupsPages(input, func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
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

func FindSpotFleetInstances(conn *ec2.EC2, input *ec2.DescribeSpotFleetInstancesInput) ([]*ec2.ActiveInstance, error) {
	var output []*ec2.ActiveInstance

	err := describeSpotFleetInstancesPages(conn, input, func(page *ec2.DescribeSpotFleetInstancesOutput, lastPage bool) bool {
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

func FindSpotFleetRequests(conn *ec2.EC2, input *ec2.DescribeSpotFleetRequestsInput) ([]*ec2.SpotFleetRequestConfig, error) {
	var output []*ec2.SpotFleetRequestConfig

	err := conn.DescribeSpotFleetRequestsPages(input, func(page *ec2.DescribeSpotFleetRequestsOutput, lastPage bool) bool {
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

func FindSpotFleetRequest(conn *ec2.EC2, input *ec2.DescribeSpotFleetRequestsInput) (*ec2.SpotFleetRequestConfig, error) {
	output, err := FindSpotFleetRequests(conn, input)

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

func FindSpotFleetRequestByID(conn *ec2.EC2, id string) (*ec2.SpotFleetRequestConfig, error) {
	input := &ec2.DescribeSpotFleetRequestsInput{
		SpotFleetRequestIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSpotFleetRequest(conn, input)

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

func FindSpotFleetRequestHistoryRecords(conn *ec2.EC2, input *ec2.DescribeSpotFleetRequestHistoryInput) ([]*ec2.HistoryRecord, error) {
	var output []*ec2.HistoryRecord

	err := describeSpotFleetRequestHistoryPages(conn, input, func(page *ec2.DescribeSpotFleetRequestHistoryOutput, lastPage bool) bool {
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

func FindSpotInstanceRequests(conn *ec2.EC2, input *ec2.DescribeSpotInstanceRequestsInput) ([]*ec2.SpotInstanceRequest, error) {
	var output []*ec2.SpotInstanceRequest

	err := conn.DescribeSpotInstanceRequestsPages(input, func(page *ec2.DescribeSpotInstanceRequestsOutput, lastPage bool) bool {
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

func FindSpotInstanceRequest(conn *ec2.EC2, input *ec2.DescribeSpotInstanceRequestsInput) (*ec2.SpotInstanceRequest, error) {
	output, err := FindSpotInstanceRequests(conn, input)

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

func FindSpotInstanceRequestByID(conn *ec2.EC2, id string) (*ec2.SpotInstanceRequest, error) {
	input := &ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSpotInstanceRequest(conn, input)

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

func FindSubnetByID(conn *ec2.EC2, id string) (*ec2.Subnet, error) {
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSubnet(conn, input)

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

func FindSubnet(conn *ec2.EC2, input *ec2.DescribeSubnetsInput) (*ec2.Subnet, error) {
	output, err := FindSubnets(conn, input)

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

func FindSubnets(conn *ec2.EC2, input *ec2.DescribeSubnetsInput) ([]*ec2.Subnet, error) {
	var output []*ec2.Subnet

	err := conn.DescribeSubnetsPages(input, func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
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

func FindSubnetCIDRReservationBySubnetIDAndReservationID(conn *ec2.EC2, subnetID, reservationID string) (*ec2.SubnetCidrReservation, error) {
	input := &ec2.GetSubnetCidrReservationsInput{
		SubnetId: aws.String(subnetID),
	}

	output, err := conn.GetSubnetCidrReservations(input)

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

func FindSubnetIPv6CIDRBlockAssociationByID(conn *ec2.EC2, associationID string) (*ec2.SubnetIpv6CidrBlockAssociation, error) {
	input := &ec2.DescribeSubnetsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"ipv6-cidr-block-association.association-id": associationID,
		}),
	}

	output, err := FindSubnet(conn, input)

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

func FindVolumeModifications(conn *ec2.EC2, input *ec2.DescribeVolumesModificationsInput) ([]*ec2.VolumeModification, error) {
	var output []*ec2.VolumeModification

	err := conn.DescribeVolumesModificationsPages(input, func(page *ec2.DescribeVolumesModificationsOutput, lastPage bool) bool {
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

func FindVolumeModification(conn *ec2.EC2, input *ec2.DescribeVolumesModificationsInput) (*ec2.VolumeModification, error) {
	output, err := FindVolumeModifications(conn, input)

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

func FindVolumeModificationByID(conn *ec2.EC2, id string) (*ec2.VolumeModification, error) {
	input := &ec2.DescribeVolumesModificationsInput{
		VolumeIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVolumeModification(conn, input)

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

func FindVPCAttribute(conn *ec2.EC2, vpcID string, attribute string) (bool, error) {
	input := &ec2.DescribeVpcAttributeInput{
		Attribute: aws.String(attribute),
		VpcId:     aws.String(vpcID),
	}

	output, err := conn.DescribeVpcAttribute(input)

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
	default:
		return false, fmt.Errorf("unsupported VPC attribute: %s", attribute)
	}

	if v == nil {
		return false, tfresource.NewEmptyResultError(input)
	}

	return aws.BoolValue(v.Value), nil
}

func FindVPCClassicLinkEnabled(conn *ec2.EC2, vpcID string) (bool, error) {
	input := &ec2.DescribeVpcClassicLinkInput{
		VpcIds: aws.StringSlice([]string{vpcID}),
	}

	output, err := conn.DescribeVpcClassicLink(input)

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

func FindVPCClassicLinkDNSSupported(conn *ec2.EC2, vpcID string) (bool, error) {
	input := &ec2.DescribeVpcClassicLinkDnsSupportInput{
		VpcIds: aws.StringSlice([]string{vpcID}),
	}

	output, err := conn.DescribeVpcClassicLinkDnsSupport(input)

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

func FindVPC(conn *ec2.EC2, input *ec2.DescribeVpcsInput) (*ec2.Vpc, error) {
	output, err := FindVPCs(conn, input)

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

func FindVPCs(conn *ec2.EC2, input *ec2.DescribeVpcsInput) ([]*ec2.Vpc, error) {
	var output []*ec2.Vpc

	err := conn.DescribeVpcsPages(input, func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
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

func FindVPCByID(conn *ec2.EC2, id string) (*ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		VpcIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPC(conn, input)

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

func FindVPCDHCPOptionsAssociation(conn *ec2.EC2, vpcID string, dhcpOptionsID string) error {
	vpc, err := FindVPCByID(conn, vpcID)

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

func FindVPCCIDRBlockAssociationByID(conn *ec2.EC2, id string) (*ec2.VpcCidrBlockAssociation, *ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"cidr-block-association.association-id": id,
		}),
	}

	vpc, err := FindVPC(conn, input)

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

func FindVPCIPv6CIDRBlockAssociationByID(conn *ec2.EC2, id string) (*ec2.VpcIpv6CidrBlockAssociation, *ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"ipv6-cidr-block-association.association-id": id,
		}),
	}

	vpc, err := FindVPC(conn, input)

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

func FindVPCDefaultNetworkACL(conn *ec2.EC2, id string) (*ec2.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"default": "true",
			"vpc-id":  id,
		}),
	}

	return FindNetworkACL(conn, input)
}

func FindVPCDefaultSecurityGroup(conn *ec2.EC2, id string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"group-name": DefaultSecurityGroupName,
			"vpc-id":     id,
		}),
	}

	return FindSecurityGroup(conn, input)
}

func FindVPCMainRouteTable(conn *ec2.EC2, id string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.main": "true",
			"vpc-id":           id,
		}),
	}

	return FindRouteTable(conn, input)
}

func FindVPCEndpoint(conn *ec2.EC2, input *ec2.DescribeVpcEndpointsInput) (*ec2.VpcEndpoint, error) {
	output, err := FindVPCEndpoints(conn, input)

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

func FindVPCEndpoints(conn *ec2.EC2, input *ec2.DescribeVpcEndpointsInput) ([]*ec2.VpcEndpoint, error) {
	var output []*ec2.VpcEndpoint

	err := conn.DescribeVpcEndpointsPages(input, func(page *ec2.DescribeVpcEndpointsOutput, lastPage bool) bool {
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

func FindVPCEndpointByID(conn *ec2.EC2, id string) (*ec2.VpcEndpoint, error) {
	input := &ec2.DescribeVpcEndpointsInput{
		VpcEndpointIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPCEndpoint(conn, input)

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

func FindVPCEndpointServiceConfiguration(conn *ec2.EC2, input *ec2.DescribeVpcEndpointServiceConfigurationsInput) (*ec2.ServiceConfiguration, error) {
	output, err := FindVPCEndpointServiceConfigurations(conn, input)

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

func FindVPCEndpointServiceConfigurations(conn *ec2.EC2, input *ec2.DescribeVpcEndpointServiceConfigurationsInput) ([]*ec2.ServiceConfiguration, error) {
	var output []*ec2.ServiceConfiguration

	err := conn.DescribeVpcEndpointServiceConfigurationsPages(input, func(page *ec2.DescribeVpcEndpointServiceConfigurationsOutput, lastPage bool) bool {
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

func FindVPCEndpointServiceConfigurationByServiceName(conn *ec2.EC2, name string) (*ec2.ServiceConfiguration, error) {
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"service-name": name,
		}),
	}

	return FindVPCEndpointServiceConfiguration(conn, input)
}

func FindVPCEndpointServices(conn *ec2.EC2, input *ec2.DescribeVpcEndpointServicesInput) ([]*ec2.ServiceDetail, []string, error) {
	var serviceDetails []*ec2.ServiceDetail
	var serviceNames []string

	err := describeVPCEndpointServicesPages(conn, input, func(page *ec2.DescribeVpcEndpointServicesOutput, lastPage bool) bool {
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

func FindVPCEndpointServiceConfigurationByID(conn *ec2.EC2, id string) (*ec2.ServiceConfiguration, error) {
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		ServiceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPCEndpointServiceConfiguration(conn, input)

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

func FindVPCEndpointServicePermissions(conn *ec2.EC2, input *ec2.DescribeVpcEndpointServicePermissionsInput) ([]*ec2.AllowedPrincipal, error) {
	var output []*ec2.AllowedPrincipal

	err := conn.DescribeVpcEndpointServicePermissionsPages(input, func(page *ec2.DescribeVpcEndpointServicePermissionsOutput, lastPage bool) bool {
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

func FindVPCEndpointServicePermissionsByID(conn *ec2.EC2, id string) ([]*ec2.AllowedPrincipal, error) {
	input := &ec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(id),
	}

	return FindVPCEndpointServicePermissions(conn, input)
}

func FindVPCEndpointServicePermissionExists(conn *ec2.EC2, serviceID, principalARN string) error {
	allowedPrincipals, err := FindVPCEndpointServicePermissionsByID(conn, serviceID)

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
func FindVPCEndpointRouteTableAssociationExists(conn *ec2.EC2, vpcEndpointID string, routeTableID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(conn, vpcEndpointID)

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
func FindVPCEndpointSecurityGroupAssociationExists(conn *ec2.EC2, vpcEndpointID, securityGroupID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(conn, vpcEndpointID)

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
func FindVPCEndpointSubnetAssociationExists(conn *ec2.EC2, vpcEndpointID string, subnetID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(conn, vpcEndpointID)

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

func FindVPCPeeringConnection(conn *ec2.EC2, input *ec2.DescribeVpcPeeringConnectionsInput) (*ec2.VpcPeeringConnection, error) {
	output, err := FindVPCPeeringConnections(conn, input)

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

func FindVPCPeeringConnections(conn *ec2.EC2, input *ec2.DescribeVpcPeeringConnectionsInput) ([]*ec2.VpcPeeringConnection, error) {
	var output []*ec2.VpcPeeringConnection

	err := conn.DescribeVpcPeeringConnectionsPages(input, func(page *ec2.DescribeVpcPeeringConnectionsOutput, lastPage bool) bool {
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

func FindVPCPeeringConnectionByID(conn *ec2.EC2, id string) (*ec2.VpcPeeringConnection, error) {
	input := &ec2.DescribeVpcPeeringConnectionsInput{
		VpcPeeringConnectionIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPCPeeringConnection(conn, input)

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
func FindVPNGatewayRoutePropagationExists(conn *ec2.EC2, routeTableID, gatewayID string) error {
	routeTable, err := FindRouteTableByID(conn, routeTableID)

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

func FindVPNGatewayVPCAttachment(conn *ec2.EC2, vpnGatewayID, vpcID string) (*ec2.VpcAttachment, error) {
	vpnGateway, err := FindVPNGatewayByID(conn, vpnGatewayID)

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

func FindVPNGatewayByID(conn *ec2.EC2, id string) (*ec2.VpnGateway, error) {
	input := &ec2.DescribeVpnGatewaysInput{
		VpnGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPNGateway(conn, input)

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

func FindVPNGateway(conn *ec2.EC2, input *ec2.DescribeVpnGatewaysInput) (*ec2.VpnGateway, error) {
	output, err := conn.DescribeVpnGateways(input)

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

func FindCustomerGatewayByID(conn *ec2.EC2, id string) (*ec2.CustomerGateway, error) {
	input := &ec2.DescribeCustomerGatewaysInput{
		CustomerGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindCustomerGateway(conn, input)

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

func FindCustomerGateway(conn *ec2.EC2, input *ec2.DescribeCustomerGatewaysInput) (*ec2.CustomerGateway, error) {
	output, err := conn.DescribeCustomerGateways(input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCustomerGatewayIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CustomerGateways) == 0 || output.CustomerGateways[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.CustomerGateways); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.CustomerGateways[0], nil
}

func FindVPNConnectionByID(conn *ec2.EC2, id string) (*ec2.VpnConnection, error) {
	input := &ec2.DescribeVpnConnectionsInput{
		VpnConnectionIds: aws.StringSlice([]string{id}),
	}

	output, err := FindVPNConnection(conn, input)

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

func FindVPNConnection(conn *ec2.EC2, input *ec2.DescribeVpnConnectionsInput) (*ec2.VpnConnection, error) {
	output, err := conn.DescribeVpnConnections(input)

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

func FindVPNConnectionRouteByVPNConnectionIDAndCIDR(conn *ec2.EC2, vpnConnectionID, cidrBlock string) (*ec2.VpnStaticRoute, error) {
	input := &ec2.DescribeVpnConnectionsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"route.destination-cidr-block": cidrBlock,
			"vpn-connection-id":            vpnConnectionID,
		}),
	}

	output, err := FindVPNConnection(conn, input)

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

func FindTransitGateway(conn *ec2.EC2, input *ec2.DescribeTransitGatewaysInput) (*ec2.TransitGateway, error) {
	output, err := FindTransitGateways(conn, input)

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

func FindTransitGateways(conn *ec2.EC2, input *ec2.DescribeTransitGatewaysInput) ([]*ec2.TransitGateway, error) {
	var output []*ec2.TransitGateway

	err := conn.DescribeTransitGatewaysPages(input, func(page *ec2.DescribeTransitGatewaysOutput, lastPage bool) bool {
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

func FindTransitGatewayByID(conn *ec2.EC2, id string) (*ec2.TransitGateway, error) {
	input := &ec2.DescribeTransitGatewaysInput{
		TransitGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGateway(conn, input)

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

func FindTransitGatewayAttachment(conn *ec2.EC2, input *ec2.DescribeTransitGatewayAttachmentsInput) (*ec2.TransitGatewayAttachment, error) {
	output, err := FindTransitGatewayAttachments(conn, input)

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

func FindTransitGatewayAttachments(conn *ec2.EC2, input *ec2.DescribeTransitGatewayAttachmentsInput) ([]*ec2.TransitGatewayAttachment, error) {
	var output []*ec2.TransitGatewayAttachment

	err := conn.DescribeTransitGatewayAttachmentsPages(input, func(page *ec2.DescribeTransitGatewayAttachmentsOutput, lastPage bool) bool {
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

func FindTransitGatewayAttachmentByID(conn *ec2.EC2, id string) (*ec2.TransitGatewayAttachment, error) {
	input := &ec2.DescribeTransitGatewayAttachmentsInput{
		TransitGatewayAttachmentIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayAttachment(conn, input)

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

func FindTransitGatewayConnect(conn *ec2.EC2, input *ec2.DescribeTransitGatewayConnectsInput) (*ec2.TransitGatewayConnect, error) {
	output, err := FindTransitGatewayConnects(conn, input)

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

func FindTransitGatewayConnects(conn *ec2.EC2, input *ec2.DescribeTransitGatewayConnectsInput) ([]*ec2.TransitGatewayConnect, error) {
	var output []*ec2.TransitGatewayConnect

	err := conn.DescribeTransitGatewayConnectsPages(input, func(page *ec2.DescribeTransitGatewayConnectsOutput, lastPage bool) bool {
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

func FindTransitGatewayConnectByID(conn *ec2.EC2, id string) (*ec2.TransitGatewayConnect, error) {
	input := &ec2.DescribeTransitGatewayConnectsInput{
		TransitGatewayAttachmentIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayConnect(conn, input)

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

func FindTransitGatewayConnectPeer(conn *ec2.EC2, input *ec2.DescribeTransitGatewayConnectPeersInput) (*ec2.TransitGatewayConnectPeer, error) {
	output, err := FindTransitGatewayConnectPeers(conn, input)

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

func FindTransitGatewayConnectPeers(conn *ec2.EC2, input *ec2.DescribeTransitGatewayConnectPeersInput) ([]*ec2.TransitGatewayConnectPeer, error) {
	var output []*ec2.TransitGatewayConnectPeer

	err := conn.DescribeTransitGatewayConnectPeersPages(input, func(page *ec2.DescribeTransitGatewayConnectPeersOutput, lastPage bool) bool {
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

func FindTransitGatewayConnectPeerByID(conn *ec2.EC2, id string) (*ec2.TransitGatewayConnectPeer, error) {
	input := &ec2.DescribeTransitGatewayConnectPeersInput{
		TransitGatewayConnectPeerIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayConnectPeer(conn, input)

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

func FindTransitGatewayMulticastDomain(conn *ec2.EC2, input *ec2.DescribeTransitGatewayMulticastDomainsInput) (*ec2.TransitGatewayMulticastDomain, error) {
	output, err := FindTransitGatewayMulticastDomains(conn, input)

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

func FindTransitGatewayMulticastDomains(conn *ec2.EC2, input *ec2.DescribeTransitGatewayMulticastDomainsInput) ([]*ec2.TransitGatewayMulticastDomain, error) {
	var output []*ec2.TransitGatewayMulticastDomain

	err := conn.DescribeTransitGatewayMulticastDomainsPages(input, func(page *ec2.DescribeTransitGatewayMulticastDomainsOutput, lastPage bool) bool {
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

func FindTransitGatewayMulticastDomainByID(conn *ec2.EC2, id string) (*ec2.TransitGatewayMulticastDomain, error) {
	input := &ec2.DescribeTransitGatewayMulticastDomainsInput{
		TransitGatewayMulticastDomainIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayMulticastDomain(conn, input)

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

func FindTransitGatewayMulticastDomainAssociation(conn *ec2.EC2, input *ec2.GetTransitGatewayMulticastDomainAssociationsInput) (*ec2.TransitGatewayMulticastDomainAssociation, error) {
	output, err := FindTransitGatewayMulticastDomainAssociations(conn, input)

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

func FindTransitGatewayMulticastDomainAssociations(conn *ec2.EC2, input *ec2.GetTransitGatewayMulticastDomainAssociationsInput) ([]*ec2.TransitGatewayMulticastDomainAssociation, error) {
	var output []*ec2.TransitGatewayMulticastDomainAssociation

	err := conn.GetTransitGatewayMulticastDomainAssociationsPages(input, func(page *ec2.GetTransitGatewayMulticastDomainAssociationsOutput, lastPage bool) bool {
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

func FindTransitGatewayMulticastDomainAssociationByThreePartKey(conn *ec2.EC2, multicastDomainID, attachmentID, subnetID string) (*ec2.TransitGatewayMulticastDomainAssociation, error) {
	input := &ec2.GetTransitGatewayMulticastDomainAssociationsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"subnet-id":                     subnetID,
			"transit-gateway-attachment-id": attachmentID,
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := FindTransitGatewayMulticastDomainAssociation(conn, input)

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

func FindTransitGatewayMulticastGroups(conn *ec2.EC2, input *ec2.SearchTransitGatewayMulticastGroupsInput) ([]*ec2.TransitGatewayMulticastGroup, error) {
	var output []*ec2.TransitGatewayMulticastGroup

	err := conn.SearchTransitGatewayMulticastGroupsPages(input, func(page *ec2.SearchTransitGatewayMulticastGroupsOutput, lastPage bool) bool {
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

func FindTransitGatewayMulticastGroupMemberByThreePartKey(conn *ec2.EC2, multicastDomainID, groupIPAddress, eniID string) (*ec2.TransitGatewayMulticastGroup, error) {
	input := &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"group-ip-address": groupIPAddress,
			"is-group-member":  "true",
			"is-group-source":  "false",
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := FindTransitGatewayMulticastGroups(conn, input)

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

func FindTransitGatewayMulticastGroupSourceByThreePartKey(conn *ec2.EC2, multicastDomainID, groupIPAddress, eniID string) (*ec2.TransitGatewayMulticastGroup, error) {
	input := &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"group-ip-address": groupIPAddress,
			"is-group-member":  "false",
			"is-group-source":  "true",
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := FindTransitGatewayMulticastGroups(conn, input)

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

func FindTransitGatewayPrefixListReference(conn *ec2.EC2, input *ec2.GetTransitGatewayPrefixListReferencesInput) (*ec2.TransitGatewayPrefixListReference, error) {
	output, err := FindTransitGatewayPrefixListReferences(conn, input)

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

func FindTransitGatewayPrefixListReferences(conn *ec2.EC2, input *ec2.GetTransitGatewayPrefixListReferencesInput) ([]*ec2.TransitGatewayPrefixListReference, error) {
	var output []*ec2.TransitGatewayPrefixListReference

	err := conn.GetTransitGatewayPrefixListReferencesPages(input, func(page *ec2.GetTransitGatewayPrefixListReferencesOutput, lastPage bool) bool {
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

func FindTransitGatewayPrefixListReferenceByTwoPartKey(conn *ec2.EC2, transitGatewayRouteTableID, prefixListID string) (*ec2.TransitGatewayPrefixListReference, error) {
	input := &ec2.GetTransitGatewayPrefixListReferencesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"prefix-list-id": prefixListID,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := FindTransitGatewayPrefixListReference(conn, input)

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

func FindTransitGatewayRoute(conn *ec2.EC2, transitGatewayRouteTableID, destination string) (*ec2.TransitGatewayRoute, error) {
	input := &ec2.SearchTransitGatewayRoutesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"type": ec2.TransitGatewayRouteTypeStatic,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := conn.SearchTransitGatewayRoutes(input)

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

func FindTransitGatewayRouteTables(conn *ec2.EC2, input *ec2.DescribeTransitGatewayRouteTablesInput) ([]*ec2.TransitGatewayRouteTable, error) {
	var output []*ec2.TransitGatewayRouteTable

	err := conn.DescribeTransitGatewayRouteTablesPages(input, func(page *ec2.DescribeTransitGatewayRouteTablesOutput, lastPage bool) bool {
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

func FindTransitGatewayRouteTablePropagation(conn *ec2.EC2, transitGatewayRouteTableID string, transitGatewayAttachmentID string) (*ec2.TransitGatewayRouteTablePropagation, error) {
	if transitGatewayRouteTableID == "" {
		return nil, nil
	}

	input := &ec2.GetTransitGatewayRouteTablePropagationsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("transit-gateway-attachment-id"),
				Values: aws.StringSlice([]string{transitGatewayAttachmentID}),
			},
		},
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	var result *ec2.TransitGatewayRouteTablePropagation

	err := conn.GetTransitGatewayRouteTablePropagationsPages(input, func(page *ec2.GetTransitGatewayRouteTablePropagationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, transitGatewayRouteTablePropagation := range page.TransitGatewayRouteTablePropagations {
			if transitGatewayRouteTablePropagation == nil {
				continue
			}

			if aws.StringValue(transitGatewayRouteTablePropagation.TransitGatewayAttachmentId) == transitGatewayAttachmentID {
				result = transitGatewayRouteTablePropagation
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func FindTransitGatewayVPCAttachment(conn *ec2.EC2, input *ec2.DescribeTransitGatewayVpcAttachmentsInput) (*ec2.TransitGatewayVpcAttachment, error) {
	output, err := FindTransitGatewayVPCAttachments(conn, input)

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

func FindTransitGatewayVPCAttachments(conn *ec2.EC2, input *ec2.DescribeTransitGatewayVpcAttachmentsInput) ([]*ec2.TransitGatewayVpcAttachment, error) {
	var output []*ec2.TransitGatewayVpcAttachment

	err := conn.DescribeTransitGatewayVpcAttachmentsPages(input, func(page *ec2.DescribeTransitGatewayVpcAttachmentsOutput, lastPage bool) bool {
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

func FindTransitGatewayVPCAttachmentByID(conn *ec2.EC2, id string) (*ec2.TransitGatewayVpcAttachment, error) {
	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{
		TransitGatewayAttachmentIds: aws.StringSlice([]string{id}),
	}

	output, err := FindTransitGatewayVPCAttachment(conn, input)

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

func FindDHCPOptions(conn *ec2.EC2, input *ec2.DescribeDhcpOptionsInput) (*ec2.DhcpOptions, error) {
	output, err := FindDHCPOptionses(conn, input)

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

func FindDHCPOptionses(conn *ec2.EC2, input *ec2.DescribeDhcpOptionsInput) ([]*ec2.DhcpOptions, error) {
	var output []*ec2.DhcpOptions

	err := conn.DescribeDhcpOptionsPages(input, func(page *ec2.DescribeDhcpOptionsOutput, lastPage bool) bool {
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

func FindDHCPOptionsByID(conn *ec2.EC2, id string) (*ec2.DhcpOptions, error) {
	input := &ec2.DescribeDhcpOptionsInput{
		DhcpOptionsIds: aws.StringSlice([]string{id}),
	}

	output, err := FindDHCPOptions(conn, input)

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

func FindEgressOnlyInternetGateway(conn *ec2.EC2, input *ec2.DescribeEgressOnlyInternetGatewaysInput) (*ec2.EgressOnlyInternetGateway, error) {
	output, err := FindEgressOnlyInternetGateways(conn, input)

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

func FindEgressOnlyInternetGateways(conn *ec2.EC2, input *ec2.DescribeEgressOnlyInternetGatewaysInput) ([]*ec2.EgressOnlyInternetGateway, error) {
	var output []*ec2.EgressOnlyInternetGateway

	err := conn.DescribeEgressOnlyInternetGatewaysPages(input, func(page *ec2.DescribeEgressOnlyInternetGatewaysOutput, lastPage bool) bool {
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

func FindEgressOnlyInternetGatewayByID(conn *ec2.EC2, id string) (*ec2.EgressOnlyInternetGateway, error) {
	input := &ec2.DescribeEgressOnlyInternetGatewaysInput{
		EgressOnlyInternetGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindEgressOnlyInternetGateway(conn, input)

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

func FindFleet(conn *ec2.EC2, input *ec2.DescribeFleetsInput) (*ec2.FleetData, error) {
	output, err := FindFleets(conn, input)

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

func FindFleets(conn *ec2.EC2, input *ec2.DescribeFleetsInput) ([]*ec2.FleetData, error) {
	var output []*ec2.FleetData

	err := conn.DescribeFleetsPages(input, func(page *ec2.DescribeFleetsOutput, lastPage bool) bool {
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

func FindFleetByID(conn *ec2.EC2, id string) (*ec2.FleetData, error) {
	input := &ec2.DescribeFleetsInput{
		FleetIds: aws.StringSlice([]string{id}),
	}

	output, err := FindFleet(conn, input)

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

func FindFlowLogByID(conn *ec2.EC2, id string) (*ec2.FlowLog, error) {
	input := &ec2.DescribeFlowLogsInput{
		FlowLogIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeFlowLogs(input)

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

func FindInternetGateway(conn *ec2.EC2, input *ec2.DescribeInternetGatewaysInput) (*ec2.InternetGateway, error) {
	output, err := FindInternetGateways(conn, input)

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

func FindInternetGateways(conn *ec2.EC2, input *ec2.DescribeInternetGatewaysInput) ([]*ec2.InternetGateway, error) {
	var output []*ec2.InternetGateway

	err := conn.DescribeInternetGatewaysPages(input, func(page *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
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

func FindInternetGatewayByID(conn *ec2.EC2, id string) (*ec2.InternetGateway, error) {
	input := &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindInternetGateway(conn, input)

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

func FindInternetGatewayAttachment(conn *ec2.EC2, internetGatewayID, vpcID string) (*ec2.InternetGatewayAttachment, error) {
	internetGateway, err := FindInternetGatewayByID(conn, internetGatewayID)

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

func FindKeyPair(conn *ec2.EC2, input *ec2.DescribeKeyPairsInput) (*ec2.KeyPairInfo, error) {
	output, err := FindKeyPairs(conn, input)

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

func FindKeyPairs(conn *ec2.EC2, input *ec2.DescribeKeyPairsInput) ([]*ec2.KeyPairInfo, error) {
	output, err := conn.DescribeKeyPairs(input)

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

func FindKeyPairByName(conn *ec2.EC2, name string) (*ec2.KeyPairInfo, error) {
	input := &ec2.DescribeKeyPairsInput{
		KeyNames: aws.StringSlice([]string{name}),
	}

	output, err := FindKeyPair(conn, input)

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

func FindLaunchTemplate(conn *ec2.EC2, input *ec2.DescribeLaunchTemplatesInput) (*ec2.LaunchTemplate, error) {
	output, err := FindLaunchTemplates(conn, input)

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

func FindLaunchTemplates(conn *ec2.EC2, input *ec2.DescribeLaunchTemplatesInput) ([]*ec2.LaunchTemplate, error) {
	var output []*ec2.LaunchTemplate

	err := conn.DescribeLaunchTemplatesPages(input, func(page *ec2.DescribeLaunchTemplatesOutput, lastPage bool) bool {
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

func FindLaunchTemplateByID(conn *ec2.EC2, id string) (*ec2.LaunchTemplate, error) {
	input := &ec2.DescribeLaunchTemplatesInput{
		LaunchTemplateIds: aws.StringSlice([]string{id}),
	}

	output, err := FindLaunchTemplate(conn, input)

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

func FindLaunchTemplateVersion(conn *ec2.EC2, input *ec2.DescribeLaunchTemplateVersionsInput) (*ec2.LaunchTemplateVersion, error) {
	output, err := FindLaunchTemplateVersions(conn, input)

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

func FindLaunchTemplateVersions(conn *ec2.EC2, input *ec2.DescribeLaunchTemplateVersionsInput) ([]*ec2.LaunchTemplateVersion, error) {
	var output []*ec2.LaunchTemplateVersion

	err := conn.DescribeLaunchTemplateVersionsPages(input, func(page *ec2.DescribeLaunchTemplateVersionsOutput, lastPage bool) bool {
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

func FindLaunchTemplateVersionByTwoPartKey(conn *ec2.EC2, launchTemplateID, version string) (*ec2.LaunchTemplateVersion, error) {
	input := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateId: aws.String(launchTemplateID),
		Versions:         aws.StringSlice([]string{version}),
	}

	output, err := FindLaunchTemplateVersion(conn, input)

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

func FindManagedPrefixListByID(conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	input := &ec2.DescribeManagedPrefixListsInput{
		PrefixListIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeManagedPrefixLists(input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPrefixListIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PrefixLists) == 0 || output.PrefixLists[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.PrefixLists); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	prefixList := output.PrefixLists[0]

	if state := aws.StringValue(prefixList.State); state == ec2.PrefixListStateDeleteComplete {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return prefixList, nil
}

func FindManagedPrefixListEntriesByID(conn *ec2.EC2, id string) ([]*ec2.PrefixListEntry, error) {
	input := &ec2.GetManagedPrefixListEntriesInput{
		PrefixListId: aws.String(id),
	}

	var prefixListEntries []*ec2.PrefixListEntry

	err := conn.GetManagedPrefixListEntriesPages(input, func(page *ec2.GetManagedPrefixListEntriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, entry := range page.Entries {
			if entry == nil {
				continue
			}

			prefixListEntries = append(prefixListEntries, entry)
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

	return prefixListEntries, nil
}

func FindManagedPrefixListEntryByIDAndCIDR(conn *ec2.EC2, id, cidr string) (*ec2.PrefixListEntry, error) {
	prefixListEntries, err := FindManagedPrefixListEntriesByID(conn, id)

	if err != nil {
		return nil, err
	}

	for _, entry := range prefixListEntries {
		if aws.StringValue(entry.Cidr) == cidr {
			return entry, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindNATGateway(conn *ec2.EC2, input *ec2.DescribeNatGatewaysInput) (*ec2.NatGateway, error) {
	output, err := FindNATGateways(conn, input)

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

func FindNATGateways(conn *ec2.EC2, input *ec2.DescribeNatGatewaysInput) ([]*ec2.NatGateway, error) {
	var output []*ec2.NatGateway

	err := conn.DescribeNatGatewaysPages(input, func(page *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
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

func FindNATGatewayByID(conn *ec2.EC2, id string) (*ec2.NatGateway, error) {
	input := &ec2.DescribeNatGatewaysInput{
		NatGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNATGateway(conn, input)

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

func FindPlacementGroupByName(conn *ec2.EC2, name string) (*ec2.PlacementGroup, error) {
	input := &ec2.DescribePlacementGroupsInput{
		GroupNames: aws.StringSlice([]string{name}),
	}

	output, err := conn.DescribePlacementGroups(input)

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

func FindPrefixList(conn *ec2.EC2, input *ec2.DescribePrefixListsInput) (*ec2.PrefixList, error) {
	output, err := FindPrefixLists(conn, input)

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

func FindPrefixLists(conn *ec2.EC2, input *ec2.DescribePrefixListsInput) ([]*ec2.PrefixList, error) {
	var output []*ec2.PrefixList

	err := conn.DescribePrefixListsPages(input, func(page *ec2.DescribePrefixListsOutput, lastPage bool) bool {
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

func FindPrefixListByName(conn *ec2.EC2, name string) (*ec2.PrefixList, error) {
	input := &ec2.DescribePrefixListsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"prefix-list-name": name,
		}),
	}

	return FindPrefixList(conn, input)
}

func FindVPCEndpointConnectionByServiceIDAndVPCEndpointID(conn *ec2.EC2, serviceID, vpcEndpointID string) (*ec2.VpcEndpointConnection, error) {
	input := &ec2.DescribeVpcEndpointConnectionsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"service-id": serviceID,
			// "InvalidFilter: The filter vpc-endpoint-id  is invalid"
			// "vpc-endpoint-id ": vpcEndpointID,
		}),
	}

	var output *ec2.VpcEndpointConnection

	err := conn.DescribeVpcEndpointConnectionsPages(input, func(page *ec2.DescribeVpcEndpointConnectionsOutput, lastPage bool) bool {
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

func FindImportSnapshotTasks(conn *ec2.EC2, input *ec2.DescribeImportSnapshotTasksInput) ([]*ec2.ImportSnapshotTask, error) {
	var output []*ec2.ImportSnapshotTask

	err := conn.DescribeImportSnapshotTasksPages(input, func(page *ec2.DescribeImportSnapshotTasksOutput, lastPage bool) bool {
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

func FindImportSnapshotTask(conn *ec2.EC2, input *ec2.DescribeImportSnapshotTasksInput) (*ec2.ImportSnapshotTask, error) {
	output, err := FindImportSnapshotTasks(conn, input)

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

func FindImportSnapshotTaskByID(conn *ec2.EC2, id string) (*ec2.ImportSnapshotTask, error) {
	input := &ec2.DescribeImportSnapshotTasksInput{
		ImportTaskIds: aws.StringSlice([]string{id}),
	}

	output, err := FindImportSnapshotTask(conn, input)

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

func FindSnapshots(conn *ec2.EC2, input *ec2.DescribeSnapshotsInput) ([]*ec2.Snapshot, error) {
	var output []*ec2.Snapshot

	err := conn.DescribeSnapshotsPages(input, func(page *ec2.DescribeSnapshotsOutput, lastPage bool) bool {
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

func FindSnapshot(conn *ec2.EC2, input *ec2.DescribeSnapshotsInput) (*ec2.Snapshot, error) {
	output, err := FindSnapshots(conn, input)

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

func FindSnapshotByID(conn *ec2.EC2, id string) (*ec2.Snapshot, error) {
	input := &ec2.DescribeSnapshotsInput{
		SnapshotIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSnapshot(conn, input)

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

func FindSnapshotAttribute(conn *ec2.EC2, input *ec2.DescribeSnapshotAttributeInput) (*ec2.DescribeSnapshotAttributeOutput, error) {
	output, err := conn.DescribeSnapshotAttribute(input)

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

func FindCreateSnapshotCreateVolumePermissionByTwoPartKey(conn *ec2.EC2, snapshotID, accountID string) (*ec2.CreateVolumePermission, error) {
	input := &ec2.DescribeSnapshotAttributeInput{
		Attribute:  aws.String(ec2.SnapshotAttributeNameCreateVolumePermission),
		SnapshotId: aws.String(snapshotID),
	}

	output, err := FindSnapshotAttribute(conn, input)

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

func FindFindSnapshotTierStatuses(conn *ec2.EC2, input *ec2.DescribeSnapshotTierStatusInput) ([]*ec2.SnapshotTierStatus, error) {
	var output []*ec2.SnapshotTierStatus

	err := conn.DescribeSnapshotTierStatusPages(input, func(page *ec2.DescribeSnapshotTierStatusOutput, lastPage bool) bool {
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

func FindFindSnapshotTierStatus(conn *ec2.EC2, input *ec2.DescribeSnapshotTierStatusInput) (*ec2.SnapshotTierStatus, error) {
	output, err := FindFindSnapshotTierStatuses(conn, input)

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

func FindSnapshotTierStatusBySnapshotID(conn *ec2.EC2, id string) (*ec2.SnapshotTierStatus, error) {
	input := &ec2.DescribeSnapshotTierStatusInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"snapshot-id": id,
		}),
	}

	output, err := FindFindSnapshotTierStatus(conn, input)

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
