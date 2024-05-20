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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func findAvailabilityZonesV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeAvailabilityZonesInput) ([]awstypes.AvailabilityZone, error) {
	output, err := conn.DescribeAvailabilityZones(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AvailabilityZones, nil
}

func findAvailabilityZone(ctx context.Context, conn *ec2.Client, input *ec2.DescribeAvailabilityZonesInput) (*awstypes.AvailabilityZone, error) {
	output, err := findAvailabilityZonesV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAvailabilityZoneGroupByName(ctx context.Context, conn *ec2.Client, name string) (*awstypes.AvailabilityZone, error) {
	input := &ec2.DescribeAvailabilityZonesInput{
		AllAvailabilityZones: aws.Bool(true),
		Filters: newAttributeFilterListV2(map[string]string{
			"group-name": name,
		}),
	}

	output, err := findAvailabilityZonesV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	// An AZ group may contain more than one AZ.
	availabilityZone := output[0]

	// Eventual consistency check.
	if aws.ToString(availabilityZone.GroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return &availabilityZone, nil
}

func findCapacityReservation(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCapacityReservationsInput) (*awstypes.CapacityReservation, error) {
	output, err := findCapacityReservations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCapacityReservations(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCapacityReservationsInput) ([]awstypes.CapacityReservation, error) {
	var output []awstypes.CapacityReservation

	pages := ec2.NewDescribeCapacityReservationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidCapacityReservationIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.CapacityReservations...)
	}

	return output, nil
}

func findCapacityReservationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.CapacityReservation, error) {
	input := &ec2.DescribeCapacityReservationsInput{
		CapacityReservationIds: []string{id},
	}

	output, err := findCapacityReservation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/capacity-reservations-using.html#capacity-reservations-view.
	if state := output.State; state == awstypes.CapacityReservationStateCancelled || state == awstypes.CapacityReservationStateExpired {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.CapacityReservationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findFleet(ctx context.Context, conn *ec2.Client, input *ec2.DescribeFleetsInput) (*awstypes.FleetData, error) {
	output, err := findFleets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFleets(ctx context.Context, conn *ec2.Client, input *ec2.DescribeFleetsInput) ([]awstypes.FleetData, error) {
	var output []awstypes.FleetData

	pages := ec2.NewDescribeFleetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidFleetIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Fleets...)
	}

	return output, nil
}

func findFleetByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.FleetData, error) {
	input := &ec2.DescribeFleetsInput{
		FleetIds: []string{id},
	}

	output, err := findFleet(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.FleetState; state == awstypes.FleetStateCodeDeleted || state == awstypes.FleetStateCodeDeletedRunning || state == awstypes.FleetStateCodeDeletedTerminatingInstances {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.FleetId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findHostByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Host, error) {
	input := &ec2.DescribeHostsInput{
		HostIds: []string{id},
	}

	output, err := findHost(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.AllocationStateReleased || state == awstypes.AllocationStateReleasedPermanentFailure {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.HostId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findHosts(ctx context.Context, conn *ec2.Client, input *ec2.DescribeHostsInput) ([]awstypes.Host, error) {
	var output []awstypes.Host

	pages := ec2.NewDescribeHostsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidHostIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Hosts...)
	}

	return output, nil
}

func findHost(ctx context.Context, conn *ec2.Client, input *ec2.DescribeHostsInput) (*awstypes.Host, error) {
	output, err := findHosts(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.Host) bool { return v.HostProperties != nil })
}

func findInstanceCreditSpecifications(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstanceCreditSpecificationsInput) ([]awstypes.InstanceCreditSpecification, error) {
	var output []awstypes.InstanceCreditSpecification

	pages := ec2.NewDescribeInstanceCreditSpecificationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.InstanceCreditSpecifications...)
	}

	return output, nil
}

func findInstanceCreditSpecification(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstanceCreditSpecificationsInput) (*awstypes.InstanceCreditSpecification, error) {
	output, err := findInstanceCreditSpecifications(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInstanceCreditSpecificationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.InstanceCreditSpecification, error) {
	input := &ec2.DescribeInstanceCreditSpecificationsInput{
		InstanceIds: []string{id},
	}

	output, err := findInstanceCreditSpecification(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.InstanceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findInstances(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstancesInput) ([]awstypes.Instance, error) {
	var output []awstypes.Instance

	pages := ec2.NewDescribeInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceIDNotFound) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		for _, v := range page.Reservations {
			output = append(output, v.Instances...)
		}
	}

	return output, nil
}

func findInstance(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstancesInput) (*awstypes.Instance, error) {
	output, err := findInstances(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.Instance) bool { return v.State != nil })
}

func FindInstanceByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{id},
	}

	output, err := findInstance(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State.Name; state == awstypes.InstanceStateNameTerminated {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.InstanceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findInstanceStatus(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstanceStatusInput) (*awstypes.InstanceStatus, error) {
	output, err := findInstanceStatuses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInstanceStatuses(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstanceStatusInput) ([]awstypes.InstanceStatus, error) {
	var output []awstypes.InstanceStatus

	pages := ec2.NewDescribeInstanceStatusPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.InstanceStatuses...)
	}

	return output, nil
}

func findInstanceState(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstanceStatusInput) (*awstypes.InstanceState, error) {
	output, err := findInstanceStatus(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if output.InstanceState == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.InstanceState, nil
}

func findInstanceStateByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.InstanceState, error) {
	input := &ec2.DescribeInstanceStatusInput{
		InstanceIds:         []string{id},
		IncludeAllInstances: aws.Bool(true),
	}

	output, err := findInstanceState(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if name := output.Name; name == awstypes.InstanceStateNameTerminated {
		return nil, &retry.NotFoundError{
			Message:     string(name),
			LastRequest: input,
		}
	}

	return output, nil
}

func findInstanceTypes(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstanceTypesInput) ([]awstypes.InstanceTypeInfo, error) {
	var output []awstypes.InstanceTypeInfo

	pages := ec2.NewDescribeInstanceTypesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.InstanceTypes...)
	}

	return output, nil
}

func findInstanceType(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstanceTypesInput) (*awstypes.InstanceTypeInfo, error) {
	output, err := findInstanceTypes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInstanceTypeByName(ctx context.Context, conn *ec2.Client, name string) (*awstypes.InstanceTypeInfo, error) {
	input := &ec2.DescribeInstanceTypesInput{
		InstanceTypes: []awstypes.InstanceType{awstypes.InstanceType(name)},
	}

	output, err := findInstanceType(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findInstanceTypeOfferings(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstanceTypeOfferingsInput) ([]awstypes.InstanceTypeOffering, error) {
	var output []awstypes.InstanceTypeOffering

	pages := ec2.NewDescribeInstanceTypeOfferingsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.InstanceTypeOfferings...)
	}

	return output, nil
}

func findLaunchTemplate(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLaunchTemplatesInput) (*awstypes.LaunchTemplate, error) {
	output, err := findLaunchTemplates(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLaunchTemplates(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLaunchTemplatesInput) ([]awstypes.LaunchTemplate, error) {
	var output []awstypes.LaunchTemplate

	pages := ec2.NewDescribeLaunchTemplatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidLaunchTemplateIdMalformed, errCodeInvalidLaunchTemplateIdNotFound, errCodeInvalidLaunchTemplateNameNotFoundException) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.LaunchTemplates...)
	}

	return output, nil
}

func findLaunchTemplateByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.LaunchTemplate, error) {
	input := &ec2.DescribeLaunchTemplatesInput{
		LaunchTemplateIds: []string{id},
	}

	output, err := findLaunchTemplate(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.LaunchTemplateId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findLaunchTemplateVersion(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLaunchTemplateVersionsInput) (*awstypes.LaunchTemplateVersion, error) {
	output, err := findLaunchTemplateVersions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.LaunchTemplateVersion) bool { return v.LaunchTemplateData != nil })
}

func findLaunchTemplateVersions(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLaunchTemplateVersionsInput) ([]awstypes.LaunchTemplateVersion, error) {
	var output []awstypes.LaunchTemplateVersion

	pages := ec2.NewDescribeLaunchTemplateVersionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidLaunchTemplateIdNotFound, errCodeInvalidLaunchTemplateNameNotFoundException, errCodeInvalidLaunchTemplateIdVersionNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.LaunchTemplateVersions...)
	}

	return output, nil
}

func FindLaunchTemplateVersionByTwoPartKey(ctx context.Context, conn *ec2.Client, launchTemplateID, version string) (*awstypes.LaunchTemplateVersion, error) {
	input := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateId: aws.String(launchTemplateID),
		Versions:         []string{version},
	}

	output, err := findLaunchTemplateVersion(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.LaunchTemplateId) != launchTemplateID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findPlacementGroup(ctx context.Context, conn *ec2.Client, input *ec2.DescribePlacementGroupsInput) (*awstypes.PlacementGroup, error) {
	output, err := findPlacementGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPlacementGroups(ctx context.Context, conn *ec2.Client, input *ec2.DescribePlacementGroupsInput) ([]awstypes.PlacementGroup, error) {
	output, err := conn.DescribePlacementGroups(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPlacementGroupUnknown) {
		return nil, &retry.NotFoundError{
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

	return output.PlacementGroups, nil
}

func findPlacementGroupByName(ctx context.Context, conn *ec2.Client, name string) (*awstypes.PlacementGroup, error) {
	input := &ec2.DescribePlacementGroupsInput{
		GroupNames: []string{name},
	}

	output, err := findPlacementGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.PlacementGroupStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func findPublicIPv4Pool(ctx context.Context, conn *ec2.Client, input *ec2.DescribePublicIpv4PoolsInput) (*awstypes.PublicIpv4Pool, error) {
	output, err := findPublicIPv4Pools(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPublicIPv4Pools(ctx context.Context, conn *ec2.Client, input *ec2.DescribePublicIpv4PoolsInput) ([]awstypes.PublicIpv4Pool, error) {
	var output []awstypes.PublicIpv4Pool

	pages := ec2.NewDescribePublicIpv4PoolsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidPublicIpv4PoolIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.PublicIpv4Pools...)
	}

	return output, nil
}

func findPublicIPv4PoolByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.PublicIpv4Pool, error) {
	input := &ec2.DescribePublicIpv4PoolsInput{
		PoolIds: []string{id},
	}

	output, err := findPublicIPv4Pool(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.PoolId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVolumeAttachmentInstanceByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{id},
	}

	output, err := findInstance(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State.Name; state == awstypes.InstanceStateNameTerminated {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.InstanceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findSpotDatafeedSubscription(ctx context.Context, conn *ec2.Client) (*awstypes.SpotDatafeedSubscription, error) {
	input := &ec2.DescribeSpotDatafeedSubscriptionInput{}

	output, err := conn.DescribeSpotDatafeedSubscription(ctx, input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidSpotDatafeedNotFound) {
		return nil, &retry.NotFoundError{
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

func FindSpotInstanceRequests(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSpotInstanceRequestsInput) ([]awstypes.SpotInstanceRequest, error) {
	var output []awstypes.SpotInstanceRequest

	pages := ec2.NewDescribeSpotInstanceRequestsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidSpotInstanceRequestIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.SpotInstanceRequests...)
	}

	return output, nil
}

func FindSpotInstanceRequest(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSpotInstanceRequestsInput) (*awstypes.SpotInstanceRequest, error) {
	output, err := FindSpotInstanceRequests(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.SpotInstanceRequest) bool { return v.Status != nil })
}

func FindSpotInstanceRequestByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.SpotInstanceRequest, error) {
	input := &ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: []string{id},
	}

	output, err := FindSpotInstanceRequest(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.SpotInstanceStateCancelled || state == awstypes.SpotInstanceStateClosed {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.SpotInstanceRequestId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findSubnetByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Subnet, error) {
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{id},
	}

	output, err := findSubnetV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.SubnetId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findSubnetV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSubnetsInput) (*awstypes.Subnet, error) {
	output, err := findSubnetsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSubnetsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSubnetsInput) ([]awstypes.Subnet, error) {
	var output []awstypes.Subnet

	pages := ec2.NewDescribeSubnetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidSubnetIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Subnets...)

	}

	return output, nil
}

func findVolumeModifications(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVolumesModificationsInput) ([]awstypes.VolumeModification, error) {
	var output []awstypes.VolumeModification

	pages := ec2.NewDescribeVolumesModificationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidVolumeNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.VolumesModifications...)
	}

	return output, nil
}

func findVolumeModification(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVolumesModificationsInput) (*awstypes.VolumeModification, error) {
	output, err := findVolumeModifications(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVolumeModificationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VolumeModification, error) {
	input := &ec2.DescribeVolumesModificationsInput{
		VolumeIds: []string{id},
	}

	output, err := findVolumeModification(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.VolumeId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

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

func findPrefixListV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribePrefixListsInput) (*awstypes.PrefixList, error) {
	output, err := findPrefixListsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPrefixListsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribePrefixListsInput) ([]awstypes.PrefixList, error) {
	var output []awstypes.PrefixList

	paginator := ec2.NewDescribePrefixListsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidPrefixListIdNotFound) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.PrefixLists...)
	}

	return output, nil
}

func findVPCEndpointByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpcEndpoint, error) {
	input := &ec2.DescribeVpcEndpointsInput{
		VpcEndpointIds: []string{id},
	}

	output, err := findVPCEndpointV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if output.State == awstypes.StateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(output.State),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.VpcEndpointId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPCEndpointV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointsInput) (*awstypes.VpcEndpoint, error) {
	output, err := findVPCEndpointsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCEndpointsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointsInput) ([]awstypes.VpcEndpoint, error) {
	var output []awstypes.VpcEndpoint

	paginator := ec2.NewDescribeVpcEndpointsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointIdNotFound) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.VpcEndpoints...)
	}

	return output, nil
}

func findPrefixListByNameV2(ctx context.Context, conn *ec2.Client, name string) (*awstypes.PrefixList, error) {
	input := &ec2.DescribePrefixListsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"prefix-list-name": name,
		}),
	}

	return findPrefixListV2(ctx, conn, input)
}

func findSpotFleetInstances(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSpotFleetInstancesInput) ([]awstypes.ActiveInstance, error) {
	var output []awstypes.ActiveInstance

	err := describeSpotFleetInstancesPages(ctx, conn, input, func(page *ec2.DescribeSpotFleetInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.ActiveInstances...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSpotFleetRequestIdNotFound) {
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

func findSpotFleetRequests(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSpotFleetRequestsInput) ([]awstypes.SpotFleetRequestConfig, error) {
	var output []awstypes.SpotFleetRequestConfig

	paginator := ec2.NewDescribeSpotFleetRequestsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidSpotFleetRequestIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.SpotFleetRequestConfigs...)
	}

	return output, nil
}

func findSpotFleetRequest(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSpotFleetRequestsInput) (*awstypes.SpotFleetRequestConfig, error) {
	output, err := findSpotFleetRequests(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.SpotFleetRequestConfig) bool { return v.SpotFleetRequestConfig != nil })
}

func findSpotFleetRequestByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.SpotFleetRequestConfig, error) {
	input := &ec2.DescribeSpotFleetRequestsInput{
		SpotFleetRequestIds: []string{id},
	}

	output, err := findSpotFleetRequest(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.SpotFleetRequestState; state == awstypes.BatchStateCancelled || state == awstypes.BatchStateCancelledRunning || state == awstypes.BatchStateCancelledTerminatingInstances {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.SpotFleetRequestId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findSpotFleetRequestHistoryRecords(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSpotFleetRequestHistoryInput) ([]awstypes.HistoryRecord, error) {
	var output []awstypes.HistoryRecord

	err := describeSpotFleetRequestHistoryPages(ctx, conn, input, func(page *ec2.DescribeSpotFleetRequestHistoryOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.HistoryRecords...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSpotFleetRequestIdNotFound) {
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

func findVPCEndpointServiceConfigurationByServiceNameV2(ctx context.Context, conn *ec2.Client, name string) (*awstypes.ServiceConfiguration, error) {
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"service-name": name,
		}),
	}

	return findVPCEndpointServiceConfigurationV2(ctx, conn, input)
}

func findVPCEndpointServiceConfigurationV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointServiceConfigurationsInput) (*awstypes.ServiceConfiguration, error) {
	output, err := findVPCEndpointServiceConfigurationsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCEndpointServiceConfigurationsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointServiceConfigurationsInput) ([]awstypes.ServiceConfiguration, error) {
	var output []awstypes.ServiceConfiguration

	paginator := ec2.NewDescribeVpcEndpointServiceConfigurationsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointServiceIdNotFound) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.ServiceConfigurations...)
	}

	return output, nil
}

// findRouteTableByIDV2 returns the route table corresponding to the specified identifier.
// Returns NotFoundError if no route table is found.
func findRouteTableByIDV2(ctx context.Context, conn *ec2.Client, routeTableID string) (*awstypes.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		RouteTableIds: []string{routeTableID},
	}

	return findRouteTableV2(ctx, conn, input)
}

// routeFinderV2 returns the route corresponding to the specified destination.
// Returns NotFoundError if no route is found.
type routeFinderV2 func(context.Context, *ec2.Client, string, string) (*awstypes.Route, error)

// findRouteByIPv4DestinationV2 returns the route corresponding to the specified IPv4 destination.
// Returns NotFoundError if no route is found.
func findRouteByIPv4DestinationV2(ctx context.Context, conn *ec2.Client, routeTableID, destinationCidr string) (*awstypes.Route, error) {
	routeTable, err := findRouteTableByIDV2(ctx, conn, routeTableID)

	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if types.CIDRBlocksEqual(aws.ToString(route.DestinationCidrBlock), destinationCidr) {
			return &route, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with IPv4 destination (%s) not found", routeTableID, destinationCidr),
	}
}

// findRouteByIPv6DestinationV2 returns the route corresponding to the specified IPv6 destination.
// Returns NotFoundError if no route is found.
func findRouteByIPv6DestinationV2(ctx context.Context, conn *ec2.Client, routeTableID, destinationIpv6Cidr string) (*awstypes.Route, error) {
	routeTable, err := findRouteTableByIDV2(ctx, conn, routeTableID)

	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if types.CIDRBlocksEqual(aws.ToString(route.DestinationIpv6CidrBlock), destinationIpv6Cidr) {
			return &route, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with IPv6 destination (%s) not found", routeTableID, destinationIpv6Cidr),
	}
}

// findRouteByPrefixListIDDestinationV2 returns the route corresponding to the specified prefix list destination.
// Returns NotFoundError if no route is found.
func findRouteByPrefixListIDDestinationV2(ctx context.Context, conn *ec2.Client, routeTableID, prefixListID string) (*awstypes.Route, error) {
	routeTable, err := findRouteTableByIDV2(ctx, conn, routeTableID)
	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if aws.ToString(route.DestinationPrefixListId) == prefixListID {
			return &route, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with Prefix List ID destination (%s) not found", routeTableID, prefixListID),
	}
}

// findRouteTableAssociationByIDV2 returns the route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func findRouteTableAssociationByIDV2(ctx context.Context, conn *ec2.Client, associationID string) (*awstypes.RouteTableAssociation, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"association.route-table-association-id": associationID,
		}),
	}

	routeTable, err := findRouteTableV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, association := range routeTable.Associations {
		if aws.ToString(association.RouteTableAssociationId) == associationID {
			if association.AssociationState != nil {
				if state := association.AssociationState.State; state == awstypes.RouteTableAssociationStateCodeDisassociated {
					return nil, &retry.NotFoundError{Message: string(state)}
				}
			}

			return &association, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

// findMainRouteTableByVPCIDV2 returns the main route table for the specified VPC.
// Returns NotFoundError if no route table is found.
func findMainRouteTableByVPCIDV2(ctx context.Context, conn *ec2.Client, vpcID string) (*awstypes.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"association.main": "true",
			"vpc-id":           vpcID,
		}),
	}

	return findRouteTableV2(ctx, conn, input)
}

// findVPNGatewayRoutePropagationExistsV2 returns NotFoundError if no route propagation for the specified VPN gateway is found.
func findVPNGatewayRoutePropagationExistsV2(ctx context.Context, conn *ec2.Client, routeTableID, gatewayID string) error {
	routeTable, err := findRouteTableByIDV2(ctx, conn, routeTableID)

	if err != nil {
		return err
	}

	for _, v := range routeTable.PropagatingVgws {
		if aws.ToString(v.GatewayId) == gatewayID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("Route Table (%s) VPN Gateway (%s) route propagation not found", routeTableID, gatewayID),
	}
}

func findVPCEndpointServiceConfigurationByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ServiceConfiguration, error) {
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		ServiceIds: []string{id},
	}

	output, err := findVPCEndpointServiceConfigurationV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.ServiceState; state == awstypes.ServiceStateDeleted || state == awstypes.ServiceStateFailed {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.ServiceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPCEndpointServicePermissionsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointServicePermissionsInput) ([]awstypes.AllowedPrincipal, error) {
	var output []awstypes.AllowedPrincipal

	paginator := ec2.NewDescribeVpcEndpointServicePermissionsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointServiceIdNotFound) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.AllowedPrincipals...)
	}

	return output, nil
}

func findVPCEndpointServicePermissionsByServiceIDV2(ctx context.Context, conn *ec2.Client, id string) ([]awstypes.AllowedPrincipal, error) {
	input := &ec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(id),
	}

	return findVPCEndpointServicePermissionsV2(ctx, conn, input)
}

func findVPCEndpointServicesV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointServicesInput) ([]awstypes.ServiceDetail, []string, error) {
	var serviceDetails []awstypes.ServiceDetail
	var serviceNames []string

	err := describeVPCEndpointServicesPages(ctx, conn, input, func(page *ec2.DescribeVpcEndpointServicesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		serviceDetails = append(serviceDetails, page.ServiceDetails...)
		serviceNames = append(serviceNames, page.ServiceNames...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidServiceName) {
		return nil, nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, nil, err
	}

	return serviceDetails, serviceNames, nil
}

// findVPCEndpointRouteTableAssociationExistsV2 returns NotFoundError if no association for the specified VPC endpoint and route table IDs is found.
func findVPCEndpointRouteTableAssociationExistsV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, routeTableID string) error {
	vpcEndpoint, err := findVPCEndpointByIDV2(ctx, conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointRouteTableID := range vpcEndpoint.RouteTableIds {
		if vpcEndpointRouteTableID == routeTableID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Route Table (%s) Association not found", vpcEndpointID, routeTableID),
	}
}

// findVPCEndpointSecurityGroupAssociationExistsV2 returns NotFoundError if no association for the specified VPC endpoint and security group IDs is found.
func findVPCEndpointSecurityGroupAssociationExistsV2(ctx context.Context, conn *ec2.Client, vpcEndpointID, securityGroupID string) error {
	vpcEndpoint, err := findVPCEndpointByIDV2(ctx, conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, group := range vpcEndpoint.Groups {
		if aws.ToString(group.GroupId) == securityGroupID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Security Group (%s) Association not found", vpcEndpointID, securityGroupID),
	}
}

// findVPCEndpointSubnetAssociationExistsV2 returns NotFoundError if no association for the specified VPC endpoint and subnet IDs is found.
func findVPCEndpointSubnetAssociationExistsV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, subnetID string) error {
	vpcEndpoint, err := findVPCEndpointByIDV2(ctx, conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointSubnetID := range vpcEndpoint.SubnetIds {
		if vpcEndpointSubnetID == subnetID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Subnet (%s) Association not found", vpcEndpointID, subnetID),
	}
}

func findVPCEndpointConnectionByServiceIDAndVPCEndpointIDV2(ctx context.Context, conn *ec2.Client, serviceID, vpcEndpointID string) (*awstypes.VpcEndpointConnection, error) {
	input := &ec2.DescribeVpcEndpointConnectionsInput{
		Filters: newAttributeFilterListV2(map[string]string{
			"service-id": serviceID,
			// "InvalidFilter: The filter vpc-endpoint-id  is invalid"
			// "vpc-endpoint-id ": vpcEndpointID,
		}),
	}

	var output *awstypes.VpcEndpointConnection

	paginator := ec2.NewDescribeVpcEndpointConnectionsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.VpcEndpointConnections {
			v := v
			if aws.ToString(v.VpcEndpointId) == vpcEndpointID {
				output = &v
				break
			}
		}
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if vpcEndpointState := string(output.VpcEndpointState); vpcEndpointState == vpcEndpointStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     vpcEndpointState,
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPCEndpointConnectionNotificationV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointConnectionNotificationsInput) (*awstypes.ConnectionNotification, error) {
	output, err := findVPCEndpointConnectionNotificationsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCEndpointConnectionNotificationsV2(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointConnectionNotificationsInput) ([]awstypes.ConnectionNotification, error) {
	var output []awstypes.ConnectionNotification

	paginator := ec2.NewDescribeVpcEndpointConnectionNotificationsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidConnectionNotification) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.ConnectionNotificationSet...)
	}

	return output, nil
}

func findVPCEndpointConnectionNotificationByIDV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ConnectionNotification, error) {
	input := &ec2.DescribeVpcEndpointConnectionNotificationsInput{
		ConnectionNotificationId: aws.String(id),
	}

	output, err := findVPCEndpointConnectionNotificationV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.ConnectionNotificationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPCEndpointServicePermissionV2(ctx context.Context, conn *ec2.Client, serviceID, principalARN string) (*awstypes.AllowedPrincipal, error) {
	// Applying a server-side filter on "principal" can lead to errors like
	// "An error occurred (InvalidFilter) when calling the DescribeVpcEndpointServicePermissions operation: The filter value arn:aws:iam::123456789012:role/developer contains unsupported characters".
	// Apply the filter client-side.
	input := &ec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(serviceID),
	}

	allowedPrincipals, err := findVPCEndpointServicePermissionsV2(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	allowedPrincipals = tfslices.Filter(allowedPrincipals, func(v awstypes.AllowedPrincipal) bool {
		return aws.ToString(v.Principal) == principalARN
	})

	return tfresource.AssertSingleValueResult(allowedPrincipals)
}

func FindImages(ctx context.Context, conn *ec2.Client, input *ec2.DescribeImagesInput) ([]awstypes.Image, error) {
	var output []awstypes.Image

	pages := ec2.NewDescribeImagesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidAMIIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Images...)
	}

	return output, nil
}

func FindImage(ctx context.Context, conn *ec2.Client, input *ec2.DescribeImagesInput) (*awstypes.Image, error) {
	output, err := FindImages(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func FindImageByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Image, error) {
	input := &ec2.DescribeImagesInput{
		ImageIds: []string{id},
	}

	output, err := FindImage(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.ImageStateDeregistered {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.ImageId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindImageAttribute(ctx context.Context, conn *ec2.Client, input *ec2.DescribeImageAttributeInput) (*ec2.DescribeImageAttributeOutput, error) {
	output, err := conn.DescribeImageAttribute(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAMIIDNotFound, errCodeInvalidAMIIDUnavailable) {
		return nil, &retry.NotFoundError{
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

func FindImageLaunchPermissionsByID(ctx context.Context, conn *ec2.Client, id string) ([]awstypes.LaunchPermission, error) {
	input := &ec2.DescribeImageAttributeInput{
		Attribute: awstypes.ImageAttributeNameLaunchPermission,
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

func FindImageLaunchPermission(ctx context.Context, conn *ec2.Client, imageID, accountID, group, organizationARN, organizationalUnitARN string) (*awstypes.LaunchPermission, error) {
	output, err := FindImageLaunchPermissionsByID(ctx, conn, imageID)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if (accountID != "" && aws.ToString(v.UserId) == accountID) ||
			(group != "" && string(v.Group) == group) ||
			(organizationARN != "" && aws.ToString(v.OrganizationArn) == organizationARN) ||
			(organizationalUnitARN != "" && aws.ToString(v.OrganizationalUnitArn) == organizationalUnitARN) {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}
