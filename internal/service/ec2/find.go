// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func findAvailabilityZones(ctx context.Context, conn *ec2.Client, input *ec2.DescribeAvailabilityZonesInput) ([]awstypes.AvailabilityZone, error) {
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
	output, err := findAvailabilityZones(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAvailabilityZoneGroupByName(ctx context.Context, conn *ec2.Client, name string) (*awstypes.AvailabilityZone, error) {
	input := &ec2.DescribeAvailabilityZonesInput{
		AllAvailabilityZones: aws.Bool(true),
		Filters: newAttributeFilterList(map[string]string{
			"group-name": name,
		}),
	}

	output, err := findAvailabilityZones(ctx, conn, input)

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

		if tfawserr.ErrCodeEquals(err, errCodeInvalidCapacityReservationIdNotFound, errCodeInvalidReservationNotFound) {
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

func findCOIPPool(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCoipPoolsInput) (*awstypes.CoipPool, error) {
	output, err := findCOIPPools(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCOIPPools(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCoipPoolsInput) ([]awstypes.CoipPool, error) {
	var output []awstypes.CoipPool

	pages := ec2.NewDescribeCoipPoolsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidPoolIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.CoipPools...)
	}

	return output, nil
}

func findDHCPOptions(ctx context.Context, conn *ec2.Client, input *ec2.DescribeDhcpOptionsInput) (*awstypes.DhcpOptions, error) {
	output, err := findDHCPOptionses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDHCPOptionses(ctx context.Context, conn *ec2.Client, input *ec2.DescribeDhcpOptionsInput) ([]awstypes.DhcpOptions, error) {
	var output []awstypes.DhcpOptions

	pages := ec2.NewDescribeDhcpOptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidDHCPOptionIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.DhcpOptions...)
	}

	return output, nil
}

func findDHCPOptionsByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.DhcpOptions, error) {
	input := &ec2.DescribeDhcpOptionsInput{
		DhcpOptionsIds: []string{id},
	}

	output, err := findDHCPOptions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DhcpOptionsId) != id {
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

func findInstanceByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Instance, error) {
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

func findInternetGateway(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInternetGatewaysInput) (*awstypes.InternetGateway, error) {
	output, err := findInternetGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInternetGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInternetGatewaysInput) ([]awstypes.InternetGateway, error) {
	var output []awstypes.InternetGateway

	pages := ec2.NewDescribeInternetGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidInternetGatewayIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.InternetGateways...)
	}

	return output, nil
}

func findInternetGatewayByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.InternetGateway, error) {
	input := &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: []string{id},
	}

	output, err := findInternetGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.InternetGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findInternetGatewayAttachment(ctx context.Context, conn *ec2.Client, internetGatewayID, vpcID string) (*awstypes.InternetGatewayAttachment, error) {
	internetGateway, err := findInternetGatewayByID(ctx, conn, internetGatewayID)

	if err != nil {
		return nil, err
	}

	if len(internetGateway.Attachments) == 0 {
		return nil, tfresource.NewEmptyResultError(internetGatewayID)
	}

	if count := len(internetGateway.Attachments); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, internetGatewayID)
	}

	attachment := internetGateway.Attachments[0]

	if aws.ToString(attachment.VpcId) != vpcID {
		return nil, tfresource.NewEmptyResultError(vpcID)
	}

	return &attachment, nil
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

func findLaunchTemplateVersionByTwoPartKey(ctx context.Context, conn *ec2.Client, launchTemplateID, version string) (*awstypes.LaunchTemplateVersion, error) {
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

func findLocalGatewayRouteTable(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLocalGatewayRouteTablesInput) (*awstypes.LocalGatewayRouteTable, error) {
	output, err := findLocalGatewayRouteTables(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLocalGatewayRouteTables(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLocalGatewayRouteTablesInput) ([]awstypes.LocalGatewayRouteTable, error) {
	var output []awstypes.LocalGatewayRouteTable

	pages := ec2.NewDescribeLocalGatewayRouteTablesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.LocalGatewayRouteTables...)
	}

	return output, nil
}

func findLocalGatewayRoutes(ctx context.Context, conn *ec2.Client, input *ec2.SearchLocalGatewayRoutesInput) ([]awstypes.LocalGatewayRoute, error) {
	var output []awstypes.LocalGatewayRoute

	pages := ec2.NewSearchLocalGatewayRoutesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidLocalGatewayRouteTableIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Routes...)
	}

	return output, nil
}

func findLocalGatewayRouteByTwoPartKey(ctx context.Context, conn *ec2.Client, localGatewayRouteTableID, destinationCIDRBlock string) (*awstypes.LocalGatewayRoute, error) {
	input := &ec2.SearchLocalGatewayRoutesInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String(names.AttrType),
				Values: enum.Slice(awstypes.LocalGatewayRouteTypeStatic),
			},
		},
		LocalGatewayRouteTableId: aws.String(localGatewayRouteTableID),
	}

	localGatewayRoutes, err := findLocalGatewayRoutes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	localGatewayRoutes = tfslices.Filter(localGatewayRoutes, func(v awstypes.LocalGatewayRoute) bool {
		return aws.ToString(v.DestinationCidrBlock) == destinationCIDRBlock
	})

	output, err := tfresource.AssertSingleValueResult(localGatewayRoutes)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.LocalGatewayRouteStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func findLocalGatewayRouteTableVPCAssociation(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput) (*awstypes.LocalGatewayRouteTableVpcAssociation, error) {
	output, err := findLocalGatewayRouteTableVPCAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLocalGatewayRouteTableVPCAssociations(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput) ([]awstypes.LocalGatewayRouteTableVpcAssociation, error) {
	var output []awstypes.LocalGatewayRouteTableVpcAssociation

	pages := ec2.NewDescribeLocalGatewayRouteTableVpcAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.LocalGatewayRouteTableVpcAssociations...)
	}

	return output, nil
}

func findLocalGatewayRouteTableVPCAssociationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.LocalGatewayRouteTableVpcAssociation, error) {
	input := &ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{
		LocalGatewayRouteTableVpcAssociationIds: []string{id},
	}

	output, err := findLocalGatewayRouteTableVPCAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.ToString(output.State); state == string(awstypes.RouteTableAssociationStateCodeDisassociated) {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.LocalGatewayRouteTableVpcAssociationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findLocalGatewayVirtualInterface(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLocalGatewayVirtualInterfacesInput) (*awstypes.LocalGatewayVirtualInterface, error) {
	output, err := findLocalGatewayVirtualInterfaces(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLocalGatewayVirtualInterfaces(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLocalGatewayVirtualInterfacesInput) ([]awstypes.LocalGatewayVirtualInterface, error) {
	var output []awstypes.LocalGatewayVirtualInterface

	pages := ec2.NewDescribeLocalGatewayVirtualInterfacesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.LocalGatewayVirtualInterfaces...)
	}

	return output, nil
}

func findLocalGatewayVirtualInterfaceGroup(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput) (*awstypes.LocalGatewayVirtualInterfaceGroup, error) {
	output, err := findLocalGatewayVirtualInterfaceGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLocalGatewayVirtualInterfaceGroups(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLocalGatewayVirtualInterfaceGroupsInput) ([]awstypes.LocalGatewayVirtualInterfaceGroup, error) {
	var output []awstypes.LocalGatewayVirtualInterfaceGroup

	pages := ec2.NewDescribeLocalGatewayVirtualInterfaceGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.LocalGatewayVirtualInterfaceGroups...)
	}

	return output, nil
}

func findLocalGateway(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLocalGatewaysInput) (*awstypes.LocalGateway, error) {
	output, err := findLocalGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLocalGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeLocalGatewaysInput) ([]awstypes.LocalGateway, error) {
	var output []awstypes.LocalGateway

	pages := ec2.NewDescribeLocalGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.LocalGateways...)
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

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSpotDatafeedNotFound) {
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

func findSpotInstanceRequests(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSpotInstanceRequestsInput) ([]awstypes.SpotInstanceRequest, error) {
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

func findSpotInstanceRequest(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSpotInstanceRequestsInput) (*awstypes.SpotInstanceRequest, error) {
	output, err := findSpotInstanceRequests(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.SpotInstanceRequest) bool { return v.Status != nil })
}

func findSpotInstanceRequestByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.SpotInstanceRequest, error) {
	input := &ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: []string{id},
	}

	output, err := findSpotInstanceRequest(ctx, conn, input)

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

func findSpotPrices(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSpotPriceHistoryInput) ([]awstypes.SpotPrice, error) {
	var output []awstypes.SpotPrice
	pages := ec2.NewDescribeSpotPriceHistoryPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.SpotPriceHistory...)
	}

	return output, nil
}

func findSpotPrice(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSpotPriceHistoryInput) (*awstypes.SpotPrice, error) {
	output, err := findSpotPrices(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSubnetByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Subnet, error) {
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{id},
	}

	output, err := findSubnet(ctx, conn, input)

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

func findSubnet(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSubnetsInput) (*awstypes.Subnet, error) {
	output, err := findSubnets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSubnets(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSubnetsInput) ([]awstypes.Subnet, error) {
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

func findSubnetCIDRReservationBySubnetIDAndReservationID(ctx context.Context, conn *ec2.Client, subnetID, reservationID string) (*awstypes.SubnetCidrReservation, error) {
	input := &ec2.GetSubnetCidrReservationsInput{
		SubnetId: aws.String(subnetID),
	}

	output, err := conn.GetSubnetCidrReservations(ctx, input)

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
		if aws.ToString(r.SubnetCidrReservationId) == reservationID {
			return &r, nil
		}
	}
	for _, r := range output.SubnetIpv6CidrReservations {
		if aws.ToString(r.SubnetCidrReservationId) == reservationID {
			return &r, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError:   err,
		LastRequest: input,
	}
}

func findSubnetIPv6CIDRBlockAssociationByID(ctx context.Context, conn *ec2.Client, associationID string) (*awstypes.SubnetIpv6CidrBlockAssociation, error) {
	input := &ec2.DescribeSubnetsInput{
		Filters: newAttributeFilterList(map[string]string{
			"ipv6-cidr-block-association.association-id": associationID,
		}),
	}

	output, err := findSubnet(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, association := range output.Ipv6CidrBlockAssociationSet {
		if aws.ToString(association.AssociationId) == associationID {
			if state := association.Ipv6CidrBlockState.State; state == awstypes.SubnetCidrBlockStateCodeDisassociated {
				return nil, &retry.NotFoundError{Message: string(state)}
			}

			return &association, nil
		}
	}

	return nil, &retry.NotFoundError{}
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

func findVPCAttribute(ctx context.Context, conn *ec2.Client, vpcID string, attribute awstypes.VpcAttributeName) (bool, error) {
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

func findVPC(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcsInput) (*awstypes.Vpc, error) {
	output, err := findVPCs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcsInput) ([]awstypes.Vpc, error) {
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

func findVPCByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		VpcIds: []string{id},
	}

	return findVPC(ctx, conn, input)
}

func findVPCCIDRBlockAssociationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpcCidrBlockAssociation, *awstypes.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: newAttributeFilterList(map[string]string{
			"cidr-block-association.association-id": id,
		}),
	}

	vpc, err := findVPC(ctx, conn, input)

	if err != nil {
		return nil, nil, err
	}

	association, err := tfresource.AssertSingleValueResult(tfslices.Filter(vpc.CidrBlockAssociationSet, func(v awstypes.VpcCidrBlockAssociation) bool {
		return aws.ToString(v.AssociationId) == id
	}))

	if err != nil {
		return nil, nil, err
	}

	if state := association.CidrBlockState.State; state == awstypes.VpcCidrBlockStateCodeDisassociated {
		return nil, nil, &retry.NotFoundError{
			Message: string(state),
		}
	}

	return association, vpc, nil
}

func findVPCIPv6CIDRBlockAssociationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpcIpv6CidrBlockAssociation, *awstypes.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: newAttributeFilterList(map[string]string{
			"ipv6-cidr-block-association.association-id": id,
		}),
	}

	vpc, err := findVPC(ctx, conn, input)

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

func findVPCDefaultNetworkACL(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"default": "true",
			"vpc-id":  id,
		}),
	}

	return findNetworkACL(ctx, conn, input)
}

func findNATGateway(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNatGatewaysInput) (*awstypes.NatGateway, error) {
	output, err := findNATGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findNATGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNatGatewaysInput) ([]awstypes.NatGateway, error) {
	var output []awstypes.NatGateway

	pages := ec2.NewDescribeNatGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeNatGatewayNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.NatGateways...)
	}

	return output, nil
}

func findNATGatewayByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NatGateway, error) {
	input := &ec2.DescribeNatGatewaysInput{
		NatGatewayIds: []string{id},
	}

	output, err := findNATGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.NatGatewayStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.NatGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findNATGatewayAddressByNATGatewayIDAndAllocationID(ctx context.Context, conn *ec2.Client, natGatewayID, allocationID string) (*awstypes.NatGatewayAddress, error) {
	output, err := findNATGatewayByID(ctx, conn, natGatewayID)

	if err != nil {
		return nil, err
	}

	for _, v := range output.NatGatewayAddresses {
		if aws.ToString(v.AllocationId) == allocationID {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func findNATGatewayAddressByNATGatewayIDAndPrivateIP(ctx context.Context, conn *ec2.Client, natGatewayID, privateIP string) (*awstypes.NatGatewayAddress, error) {
	output, err := findNATGatewayByID(ctx, conn, natGatewayID)

	if err != nil {
		return nil, err
	}

	for _, v := range output.NatGatewayAddresses {
		if aws.ToString(v.PrivateIp) == privateIP {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func findNetworkACLByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		NetworkAclIds: []string{id},
	}

	output, err := findNetworkACL(ctx, conn, input)

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

func findNetworkACL(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkAclsInput) (*awstypes.NetworkAcl, error) {
	output, err := findNetworkACLs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findNetworkACLs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkAclsInput) ([]awstypes.NetworkAcl, error) {
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

func findNetworkACLAssociationByID(ctx context.Context, conn *ec2.Client, associationID string) (*awstypes.NetworkAclAssociation, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"association.association-id": associationID,
		}),
	}

	output, err := findNetworkACL(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Associations {
		if aws.ToString(v.NetworkAclAssociationId) == associationID {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func findNetworkACLAssociationBySubnetID(ctx context.Context, conn *ec2.Client, subnetID string) (*awstypes.NetworkAclAssociation, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"association.subnet-id": subnetID,
		}),
	}

	output, err := findNetworkACL(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Associations {
		if aws.ToString(v.SubnetId) == subnetID {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func findNetworkACLEntryByThreePartKey(ctx context.Context, conn *ec2.Client, naclID string, egress bool, ruleNumber int) (*awstypes.NetworkAclEntry, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"entry.egress":      strconv.FormatBool(egress),
			"entry.rule-number": strconv.Itoa(ruleNumber),
		}),
		NetworkAclIds: []string{naclID},
	}

	output, err := findNetworkACL(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Entries {
		if aws.ToBool(v.Egress) == egress && aws.ToInt32(v.RuleNumber) == int32(ruleNumber) {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func findVPCDefaultSecurityGroup(ctx context.Context, conn *ec2.Client, id string) (*awstypes.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterList(map[string]string{
			"group-name": defaultSecurityGroupName,
			"vpc-id":     id,
		}),
	}

	return findSecurityGroup(ctx, conn, input)
}

func findVPCDHCPOptionsAssociation(ctx context.Context, conn *ec2.Client, vpcID string, dhcpOptionsID string) error {
	vpc, err := findVPCByID(ctx, conn, vpcID)

	if err != nil {
		return err
	}

	if aws.ToString(vpc.DhcpOptionsId) != dhcpOptionsID {
		return &retry.NotFoundError{
			LastError: fmt.Errorf("EC2 VPC (%s) DHCP Options Set (%s) Association not found", vpcID, dhcpOptionsID),
		}
	}

	return nil
}

func findVPCMainRouteTable(ctx context.Context, conn *ec2.Client, id string) (*awstypes.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: newAttributeFilterList(map[string]string{
			"association.main": "true",
			"vpc-id":           id,
		}),
	}

	return findRouteTable(ctx, conn, input)
}

func findRouteTable(ctx context.Context, conn *ec2.Client, input *ec2.DescribeRouteTablesInput) (*awstypes.RouteTable, error) {
	output, err := findRouteTables(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findRouteTables(ctx context.Context, conn *ec2.Client, input *ec2.DescribeRouteTablesInput) ([]awstypes.RouteTable, error) {
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

func findSecurityGroup(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupsInput) (*awstypes.SecurityGroup, error) {
	output, err := findSecurityGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSecurityGroups(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupsInput) ([]awstypes.SecurityGroup, error) {
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

// findSecurityGroupByNameAndVPCID looks up a security group by name, VPC ID. Returns a retry.NotFoundError if not found.
func findSecurityGroupByNameAndVPCID(ctx context.Context, conn *ec2.Client, name, vpcID string) (*awstypes.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"group-name": name,
				"vpc-id":     vpcID,
			},
		),
	}

	return findSecurityGroup(ctx, conn, input)
}

func findSecurityGroupByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{id},
	}

	output, err := findSecurityGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.GroupId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findSecurityGroupByDescriptionAndVPCID(ctx context.Context, conn *ec2.Client, description, vpcID string) (*awstypes.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"description": description, // nosemgrep:ci.literal-description-string-constant
				"vpc-id":      vpcID,
			},
		),
	}
	return findSecurityGroup(ctx, conn, input)
}

func findSecurityGroupByNameAndVPCIDAndOwnerID(ctx context.Context, conn *ec2.Client, name, vpcID, ownerID string) (*awstypes.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"group-name": name,
				"vpc-id":     vpcID,
				"owner-id":   ownerID,
			},
		),
	}
	return findSecurityGroup(ctx, conn, input)
}

func findSecurityGroupRule(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupRulesInput) (*awstypes.SecurityGroupRule, error) {
	output, err := findSecurityGroupRules(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSecurityGroupRules(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupRulesInput) ([]awstypes.SecurityGroupRule, error) {
	var output []awstypes.SecurityGroupRule

	pages := ec2.NewDescribeSecurityGroupRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidSecurityGroupRuleIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.SecurityGroupRules...)
	}

	return output, nil
}

func findSecurityGroupRuleByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.SecurityGroupRule, error) {
	input := &ec2.DescribeSecurityGroupRulesInput{
		SecurityGroupRuleIds: []string{id},
	}

	output, err := findSecurityGroupRule(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.SecurityGroupRuleId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findSecurityGroupEgressRuleByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.SecurityGroupRule, error) {
	output, err := findSecurityGroupRuleByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	if !aws.ToBool(output.IsEgress) {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func findSecurityGroupIngressRuleByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.SecurityGroupRule, error) {
	output, err := findSecurityGroupRuleByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	if aws.ToBool(output.IsEgress) {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func findSecurityGroupRulesBySecurityGroupID(ctx context.Context, conn *ec2.Client, id string) ([]awstypes.SecurityGroupRule, error) {
	input := &ec2.DescribeSecurityGroupRulesInput{
		Filters: newAttributeFilterList(map[string]string{
			"group-id": id,
		}),
	}

	return findSecurityGroupRules(ctx, conn, input)
}

func findNetworkInterfaces(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkInterfacesInput) ([]awstypes.NetworkInterface, error) {
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

func findNetworkInterface(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkInterfacesInput) (*awstypes.NetworkInterface, error) {
	output, err := findNetworkInterfaces(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findNetworkInterfaceByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{id},
	}

	output, err := findNetworkInterface(ctx, conn, input)

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

func findNetworkInterfaceAttachmentByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkInterfaceAttachment, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: newAttributeFilterList(map[string]string{
			"attachment.attachment-id": id,
		}),
	}

	networkInterface, err := findNetworkInterface(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if networkInterface.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return networkInterface.Attachment, nil
}

func findNetworkInterfacesByAttachmentInstanceOwnerIDAndDescription(ctx context.Context, conn *ec2.Client, attachmentInstanceOwnerID, description string) ([]awstypes.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: newAttributeFilterList(map[string]string{
			"attachment.instance-owner-id": attachmentInstanceOwnerID,
			names.AttrDescription:          description,
		}),
	}

	return findNetworkInterfaces(ctx, conn, input)
}

func findNetworkInterfaceByAttachmentID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: newAttributeFilterList(map[string]string{
			"attachment.attachment-id": id,
		}),
	}

	networkInterface, err := findNetworkInterface(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if networkInterface == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return networkInterface, nil
}

func findNetworkInterfaceSecurityGroup(ctx context.Context, conn *ec2.Client, networkInterfaceID string, securityGroupID string) (*awstypes.GroupIdentifier, error) {
	networkInterface, err := findNetworkInterfaceByID(ctx, conn, networkInterfaceID)

	if err != nil {
		return nil, err
	}

	for _, groupIdentifier := range networkInterface.Groups {
		if aws.ToString(groupIdentifier.GroupId) == securityGroupID {
			return &groupIdentifier, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError: fmt.Errorf("Network Interface (%s) Security Group (%s) not found", networkInterfaceID, securityGroupID),
	}
}

func findEBSVolumes(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVolumesInput) ([]awstypes.Volume, error) {
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

func findEBSVolume(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVolumesInput) (*awstypes.Volume, error) {
	output, err := findEBSVolumes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEBSVolumeByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Volume, error) {
	input := &ec2.DescribeVolumesInput{
		VolumeIds: []string{id},
	}

	output, err := findEBSVolume(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.VolumeStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.VolumeId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findEgressOnlyInternetGateway(ctx context.Context, conn *ec2.Client, input *ec2.DescribeEgressOnlyInternetGatewaysInput) (*awstypes.EgressOnlyInternetGateway, error) {
	output, err := findEgressOnlyInternetGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEgressOnlyInternetGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeEgressOnlyInternetGatewaysInput) ([]awstypes.EgressOnlyInternetGateway, error) {
	var output []awstypes.EgressOnlyInternetGateway

	pages := ec2.NewDescribeEgressOnlyInternetGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.EgressOnlyInternetGateways...)
	}

	return output, nil
}

func findEgressOnlyInternetGatewayByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.EgressOnlyInternetGateway, error) {
	input := &ec2.DescribeEgressOnlyInternetGatewaysInput{
		EgressOnlyInternetGatewayIds: []string{id},
	}

	output, err := findEgressOnlyInternetGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.EgressOnlyInternetGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findPrefixList(ctx context.Context, conn *ec2.Client, input *ec2.DescribePrefixListsInput) (*awstypes.PrefixList, error) {
	output, err := findPrefixLists(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findPrefixLists(ctx context.Context, conn *ec2.Client, input *ec2.DescribePrefixListsInput) ([]awstypes.PrefixList, error) {
	var output []awstypes.PrefixList

	pages := ec2.NewDescribePrefixListsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

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

func findVPCEndpointByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpcEndpoint, error) {
	input := &ec2.DescribeVpcEndpointsInput{
		VpcEndpointIds: []string{id},
	}

	output, err := findVPCEndpoint(ctx, conn, input)

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

func findVPCEndpoint(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointsInput) (*awstypes.VpcEndpoint, error) {
	output, err := findVPCEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCEndpoints(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointsInput) ([]awstypes.VpcEndpoint, error) {
	var output []awstypes.VpcEndpoint

	pages := ec2.NewDescribeVpcEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

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

func findPrefixListByName(ctx context.Context, conn *ec2.Client, name string) (*awstypes.PrefixList, error) {
	input := &ec2.DescribePrefixListsInput{
		Filters: newAttributeFilterList(map[string]string{
			"prefix-list-name": name,
		}),
	}

	return findPrefixList(ctx, conn, input)
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

func findVPCEndpointServiceConfigurationByServiceName(ctx context.Context, conn *ec2.Client, name string) (*awstypes.ServiceConfiguration, error) {
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		Filters: newAttributeFilterList(map[string]string{
			"service-name": name,
		}),
	}

	return findVPCEndpointServiceConfiguration(ctx, conn, input)
}

func findVPCEndpointServiceConfiguration(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointServiceConfigurationsInput) (*awstypes.ServiceConfiguration, error) {
	output, err := findVPCEndpointServiceConfigurations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCEndpointServiceConfigurations(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointServiceConfigurationsInput) ([]awstypes.ServiceConfiguration, error) {
	var output []awstypes.ServiceConfiguration

	pages := ec2.NewDescribeVpcEndpointServiceConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

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

// findRouteTableByID returns the route table corresponding to the specified identifier.
// Returns NotFoundError if no route table is found.
func findRouteTableByID(ctx context.Context, conn *ec2.Client, routeTableID string) (*awstypes.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		RouteTableIds: []string{routeTableID},
	}

	return findRouteTable(ctx, conn, input)
}

// routeFinder returns the route corresponding to the specified destination.
// Returns NotFoundError if no route is found.
type routeFinder func(context.Context, *ec2.Client, string, string) (*awstypes.Route, error)

// findRouteByIPv4Destination returns the route corresponding to the specified IPv4 destination.
// Returns NotFoundError if no route is found.
func findRouteByIPv4Destination(ctx context.Context, conn *ec2.Client, routeTableID, destinationCidr string) (*awstypes.Route, error) {
	routeTable, err := findRouteTableByID(ctx, conn, routeTableID)

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

// findRouteByIPv6Destination returns the route corresponding to the specified IPv6 destination.
// Returns NotFoundError if no route is found.
func findRouteByIPv6Destination(ctx context.Context, conn *ec2.Client, routeTableID, destinationIpv6Cidr string) (*awstypes.Route, error) {
	routeTable, err := findRouteTableByID(ctx, conn, routeTableID)

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

// findRouteByPrefixListIDDestination returns the route corresponding to the specified prefix list destination.
// Returns NotFoundError if no route is found.
func findRouteByPrefixListIDDestination(ctx context.Context, conn *ec2.Client, routeTableID, prefixListID string) (*awstypes.Route, error) {
	routeTable, err := findRouteTableByID(ctx, conn, routeTableID)
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

func findManagedPrefixList(ctx context.Context, conn *ec2.Client, input *ec2.DescribeManagedPrefixListsInput) (*awstypes.ManagedPrefixList, error) {
	output, err := findManagedPrefixLists(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findManagedPrefixLists(ctx context.Context, conn *ec2.Client, input *ec2.DescribeManagedPrefixListsInput) ([]awstypes.ManagedPrefixList, error) {
	var output []awstypes.ManagedPrefixList

	pages := ec2.NewDescribeManagedPrefixListsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidPrefixListIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.PrefixLists...)
	}

	return output, nil
}

func findManagedPrefixListByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ManagedPrefixList, error) {
	input := &ec2.DescribeManagedPrefixListsInput{
		PrefixListIds: []string{id},
	}

	output, err := findManagedPrefixList(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.PrefixListStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.PrefixListId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findManagedPrefixListEntries(ctx context.Context, conn *ec2.Client, input *ec2.GetManagedPrefixListEntriesInput) ([]awstypes.PrefixListEntry, error) {
	var output []awstypes.PrefixListEntry

	pages := ec2.NewGetManagedPrefixListEntriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidPrefixListIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Entries...)
	}

	return output, nil
}

func findManagedPrefixListEntriesByID(ctx context.Context, conn *ec2.Client, id string) ([]awstypes.PrefixListEntry, error) {
	input := &ec2.GetManagedPrefixListEntriesInput{
		PrefixListId: aws.String(id),
	}

	return findManagedPrefixListEntries(ctx, conn, input)
}

func findManagedPrefixListEntryByIDAndCIDR(ctx context.Context, conn *ec2.Client, id, cidr string) (*awstypes.PrefixListEntry, error) {
	prefixListEntries, err := findManagedPrefixListEntriesByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	for _, v := range prefixListEntries {
		if aws.ToString(v.Cidr) == cidr {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

// findMainRouteTableAssociationByID returns the main route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func findMainRouteTableAssociationByID(ctx context.Context, conn *ec2.Client, associationID string) (*awstypes.RouteTableAssociation, error) {
	association, err := findRouteTableAssociationByID(ctx, conn, associationID)

	if err != nil {
		return nil, err
	}

	if !aws.ToBool(association.Main) {
		return nil, &retry.NotFoundError{
			Message: fmt.Sprintf("%s is not the association with the main route table", associationID),
		}
	}

	return association, err
}

// findMainRouteTableAssociationByVPCID returns the main route table association for the specified VPC.
// Returns NotFoundError if no route table association is found.
func findMainRouteTableAssociationByVPCID(ctx context.Context, conn *ec2.Client, vpcID string) (*awstypes.RouteTableAssociation, error) {
	routeTable, err := findMainRouteTableByVPCID(ctx, conn, vpcID)

	if err != nil {
		return nil, err
	}

	for _, association := range routeTable.Associations {
		if aws.ToBool(association.Main) {
			if association.AssociationState != nil {
				if state := association.AssociationState.State; state == awstypes.RouteTableAssociationStateCodeDisassociated {
					continue
				}
			}

			return &association, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

// findRouteTableAssociationByID returns the route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func findRouteTableAssociationByID(ctx context.Context, conn *ec2.Client, associationID string) (*awstypes.RouteTableAssociation, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: newAttributeFilterList(map[string]string{
			"association.route-table-association-id": associationID,
		}),
	}

	routeTable, err := findRouteTable(ctx, conn, input)

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

// findMainRouteTableByVPCID returns the main route table for the specified VPC.
// Returns NotFoundError if no route table is found.
func findMainRouteTableByVPCID(ctx context.Context, conn *ec2.Client, vpcID string) (*awstypes.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: newAttributeFilterList(map[string]string{
			"association.main": "true",
			"vpc-id":           vpcID,
		}),
	}

	return findRouteTable(ctx, conn, input)
}

// findVPNGatewayRoutePropagationExists returns NotFoundError if no route propagation for the specified VPN gateway is found.
func findVPNGatewayRoutePropagationExists(ctx context.Context, conn *ec2.Client, routeTableID, gatewayID string) error {
	routeTable, err := findRouteTableByID(ctx, conn, routeTableID)

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

func findVPCEndpointServiceConfigurationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ServiceConfiguration, error) {
	input := &ec2.DescribeVpcEndpointServiceConfigurationsInput{
		ServiceIds: []string{id},
	}

	output, err := findVPCEndpointServiceConfiguration(ctx, conn, input)

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

func findVPCEndpointServicePrivateDNSNameConfigurationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.PrivateDnsNameConfiguration, error) {
	out, err := findVPCEndpointServiceConfigurationByID(ctx, conn, id)
	if err != nil {
		return nil, err
	}

	return out.PrivateDnsNameConfiguration, nil
}

func findVPCEndpointServicePermissions(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointServicePermissionsInput) ([]awstypes.AllowedPrincipal, error) {
	var output []awstypes.AllowedPrincipal

	pages := ec2.NewDescribeVpcEndpointServicePermissionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

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

func findVPCEndpointServicePermissionsByServiceID(ctx context.Context, conn *ec2.Client, id string) ([]awstypes.AllowedPrincipal, error) {
	input := &ec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(id),
	}

	return findVPCEndpointServicePermissions(ctx, conn, input)
}

func findVPCEndpointServices(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointServicesInput) ([]awstypes.ServiceDetail, []string, error) {
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

// findVPCEndpointRouteTableAssociationExists returns NotFoundError if no association for the specified VPC endpoint and route table IDs is found.
func findVPCEndpointRouteTableAssociationExists(ctx context.Context, conn *ec2.Client, vpcEndpointID string, routeTableID string) error {
	vpcEndpoint, err := findVPCEndpointByID(ctx, conn, vpcEndpointID)

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

// findVPCEndpointSecurityGroupAssociationExists returns NotFoundError if no association for the specified VPC endpoint and security group IDs is found.
func findVPCEndpointSecurityGroupAssociationExists(ctx context.Context, conn *ec2.Client, vpcEndpointID, securityGroupID string) error {
	vpcEndpoint, err := findVPCEndpointByID(ctx, conn, vpcEndpointID)

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

// findVPCEndpointSubnetAssociationExists returns NotFoundError if no association for the specified VPC endpoint and subnet IDs is found.
func findVPCEndpointSubnetAssociationExists(ctx context.Context, conn *ec2.Client, vpcEndpointID string, subnetID string) error {
	vpcEndpoint, err := findVPCEndpointByID(ctx, conn, vpcEndpointID)

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

func findVPCEndpointConnectionByServiceIDAndVPCEndpointID(ctx context.Context, conn *ec2.Client, serviceID, vpcEndpointID string) (*awstypes.VpcEndpointConnection, error) {
	input := &ec2.DescribeVpcEndpointConnectionsInput{
		Filters: newAttributeFilterList(map[string]string{
			"service-id": serviceID,
			// "InvalidFilter: The filter vpc-endpoint-id  is invalid"
			// "vpc-endpoint-id ": vpcEndpointID,
		}),
	}

	var output *awstypes.VpcEndpointConnection

	pages := ec2.NewDescribeVpcEndpointConnectionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
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

func findVPCEndpointConnectionNotification(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointConnectionNotificationsInput) (*awstypes.ConnectionNotification, error) {
	output, err := findVPCEndpointConnectionNotifications(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPCEndpointConnectionNotifications(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcEndpointConnectionNotificationsInput) ([]awstypes.ConnectionNotification, error) {
	var output []awstypes.ConnectionNotification

	pages := ec2.NewDescribeVpcEndpointConnectionNotificationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

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

func findVPCEndpointConnectionNotificationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ConnectionNotification, error) {
	input := &ec2.DescribeVpcEndpointConnectionNotificationsInput{
		ConnectionNotificationId: aws.String(id),
	}

	output, err := findVPCEndpointConnectionNotification(ctx, conn, input)

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

func findVPCEndpointServicePermission(ctx context.Context, conn *ec2.Client, serviceID, principalARN string) (*awstypes.AllowedPrincipal, error) {
	// Applying a server-side filter on "principal" can lead to errors like
	// "An error occurred (InvalidFilter) when calling the DescribeVpcEndpointServicePermissions operation: The filter value arn:aws:iam::123456789012:role/developer contains unsupported characters".
	// Apply the filter client-side.
	input := &ec2.DescribeVpcEndpointServicePermissionsInput{
		ServiceId: aws.String(serviceID),
	}

	allowedPrincipals, err := findVPCEndpointServicePermissions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	allowedPrincipals = tfslices.Filter(allowedPrincipals, func(v awstypes.AllowedPrincipal) bool {
		return aws.ToString(v.Principal) == principalARN
	})

	return tfresource.AssertSingleValueResult(allowedPrincipals)
}

func findVPCPeeringConnection(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcPeeringConnectionsInput) (*awstypes.VpcPeeringConnection, error) {
	output, err := findVPCPeeringConnections(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.VpcPeeringConnection) bool { return v.Status != nil })
}

func findVPCPeeringConnections(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpcPeeringConnectionsInput) ([]awstypes.VpcPeeringConnection, error) {
	var output []awstypes.VpcPeeringConnection

	pages := ec2.NewDescribeVpcPeeringConnectionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCPeeringConnectionIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.VpcPeeringConnections...)
	}

	return output, nil
}

func findVPCPeeringConnectionByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpcPeeringConnection, error) {
	input := &ec2.DescribeVpcPeeringConnectionsInput{
		VpcPeeringConnectionIds: []string{id},
	}

	output, err := findVPCPeeringConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/vpc/latest/peering/vpc-peering-basics.html#vpc-peering-lifecycle.
	switch statusCode := output.Status.Code; statusCode {
	case awstypes.VpcPeeringConnectionStateReasonCodeDeleted,
		awstypes.VpcPeeringConnectionStateReasonCodeExpired,
		awstypes.VpcPeeringConnectionStateReasonCodeFailed,
		awstypes.VpcPeeringConnectionStateReasonCodeRejected:
		return nil, &retry.NotFoundError{
			Message:     string(statusCode),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.VpcPeeringConnectionId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findClientVPNEndpoint(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnEndpointsInput) (*awstypes.ClientVpnEndpoint, error) {
	output, err := findClientVPNEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClientVPNEndpoints(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnEndpointsInput) ([]awstypes.ClientVpnEndpoint, error) {
	var output []awstypes.ClientVpnEndpoint

	pages := ec2.NewDescribeClientVpnEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ClientVpnEndpoints...)
	}

	return output, nil
}

func findClientVPNEndpointByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ClientVpnEndpoint, error) {
	input := &ec2.DescribeClientVpnEndpointsInput{
		ClientVpnEndpointIds: []string{id},
	}

	output, err := findClientVPNEndpoint(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.Status.Code; state == awstypes.ClientVpnEndpointStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.ClientVpnEndpointId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findClientVPNEndpointClientConnectResponseOptionsByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ClientConnectResponseOptions, error) {
	output, err := findClientVPNEndpointByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	if output.ClientConnectOptions == nil || output.ClientConnectOptions.Status == nil {
		return nil, tfresource.NewEmptyResultError(id)
	}

	return output.ClientConnectOptions, nil
}

func findClientVPNAuthorizationRule(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnAuthorizationRulesInput) (*awstypes.AuthorizationRule, error) {
	output, err := findClientVPNAuthorizationRules(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClientVPNAuthorizationRules(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnAuthorizationRulesInput) ([]awstypes.AuthorizationRule, error) {
	var output []awstypes.AuthorizationRule

	pages := ec2.NewDescribeClientVpnAuthorizationRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.AuthorizationRules...)
	}

	return output, nil
}

func findClientVPNAuthorizationRuleByThreePartKey(ctx context.Context, conn *ec2.Client, endpointID, targetNetworkCIDR, accessGroupID string) (*awstypes.AuthorizationRule, error) {
	filters := map[string]string{
		"destination-cidr": targetNetworkCIDR,
	}
	if accessGroupID != "" {
		filters["group-id"] = accessGroupID
	}
	input := &ec2.DescribeClientVpnAuthorizationRulesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters:             newAttributeFilterList(filters),
	}

	return findClientVPNAuthorizationRule(ctx, conn, input)
}

func findClientVPNNetworkAssociation(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnTargetNetworksInput) (*awstypes.TargetNetwork, error) {
	output, err := findClientVPNNetworkAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClientVPNNetworkAssociations(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnTargetNetworksInput) ([]awstypes.TargetNetwork, error) {
	var output []awstypes.TargetNetwork

	pages := ec2.NewDescribeClientVpnTargetNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound, errCodeInvalidClientVPNAssociationIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ClientVpnTargetNetworks...)
	}

	return output, nil
}

func findClientVPNNetworkAssociationByTwoPartKey(ctx context.Context, conn *ec2.Client, associationID, endpointID string) (*awstypes.TargetNetwork, error) {
	input := &ec2.DescribeClientVpnTargetNetworksInput{
		AssociationIds:      []string{associationID},
		ClientVpnEndpointId: aws.String(endpointID),
	}

	output, err := findClientVPNNetworkAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.Status.Code; state == awstypes.AssociationStatusCodeDisassociated {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.ClientVpnEndpointId) != endpointID || aws.ToString(output.AssociationId) != associationID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findClientVPNRoute(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnRoutesInput) (*awstypes.ClientVpnRoute, error) {
	output, err := findClientVPNRoutes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findClientVPNRoutes(ctx context.Context, conn *ec2.Client, input *ec2.DescribeClientVpnRoutesInput) ([]awstypes.ClientVpnRoute, error) {
	var output []awstypes.ClientVpnRoute

	pages := ec2.NewDescribeClientVpnRoutesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidClientVPNEndpointIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Routes...)
	}

	return output, nil
}

func findClientVPNRouteByThreePartKey(ctx context.Context, conn *ec2.Client, endpointID, targetSubnetID, destinationCIDR string) (*awstypes.ClientVpnRoute, error) {
	input := &ec2.DescribeClientVpnRoutesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters: newAttributeFilterList(map[string]string{
			"destination-cidr": destinationCIDR,
			"target-subnet":    targetSubnetID,
		}),
	}

	return findClientVPNRoute(ctx, conn, input)
}

func findCarrierGateway(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCarrierGatewaysInput) (*awstypes.CarrierGateway, error) {
	output, err := findCarrierGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCarrierGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCarrierGatewaysInput) ([]awstypes.CarrierGateway, error) {
	var output []awstypes.CarrierGateway

	pages := ec2.NewDescribeCarrierGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidCarrierGatewayIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.CarrierGateways...)
	}

	return output, nil
}

func findCarrierGatewayByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.CarrierGateway, error) {
	input := &ec2.DescribeCarrierGatewaysInput{
		CarrierGatewayIds: []string{id},
	}

	output, err := findCarrierGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.CarrierGatewayStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.CarrierGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPNConnection(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpnConnectionsInput) (*awstypes.VpnConnection, error) {
	output, err := findVPNConnections(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPNConnections(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpnConnectionsInput) ([]awstypes.VpnConnection, error) {
	output, err := conn.DescribeVpnConnections(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNConnectionIDNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output.VpnConnections, nil
}

func findVPNConnectionByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpnConnection, error) {
	input := &ec2.DescribeVpnConnectionsInput{
		VpnConnectionIds: []string{id},
	}

	output, err := findVPNConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.VpnStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.VpnConnectionId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVPNConnectionRouteByTwoPartKey(ctx context.Context, conn *ec2.Client, vpnConnectionID, cidrBlock string) (*awstypes.VpnStaticRoute, error) {
	input := &ec2.DescribeVpnConnectionsInput{
		Filters: newAttributeFilterList(map[string]string{
			"route.destination-cidr-block": cidrBlock,
			"vpn-connection-id":            vpnConnectionID,
		}),
	}

	output, err := findVPNConnection(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.Routes {
		if aws.ToString(v.DestinationCidrBlock) == cidrBlock && v.State != awstypes.VpnStateDeleted {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{
		LastError: fmt.Errorf("EC2 VPN Connection (%s) Route (%s) not found", vpnConnectionID, cidrBlock),
	}
}

func findVPNGatewayVPCAttachmentByTwoPartKey(ctx context.Context, conn *ec2.Client, vpnGatewayID, vpcID string) (*awstypes.VpcAttachment, error) {
	vpnGateway, err := findVPNGatewayByID(ctx, conn, vpnGatewayID)

	if err != nil {
		return nil, err
	}

	for _, vpcAttachment := range vpnGateway.VpcAttachments {
		if aws.ToString(vpcAttachment.VpcId) == vpcID {
			if state := vpcAttachment.State; state == awstypes.AttachmentStatusDetached {
				return nil, &retry.NotFoundError{
					Message:     string(state),
					LastRequest: vpcID,
				}
			}

			return &vpcAttachment, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(vpcID)
}

func findVPNGateway(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpnGatewaysInput) (*awstypes.VpnGateway, error) {
	output, err := findVPNGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVPNGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVpnGatewaysInput) ([]awstypes.VpnGateway, error) {
	output, err := conn.DescribeVpnGateways(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNGatewayIDNotFound) {
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

	return output.VpnGateways, nil
}

func findVPNGatewayByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpnGateway, error) {
	input := &ec2.DescribeVpnGatewaysInput{
		VpnGatewayIds: []string{id},
	}

	output, err := findVPNGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.VpnStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.VpnGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findCustomerGateway(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCustomerGatewaysInput) (*awstypes.CustomerGateway, error) {
	output, err := findCustomerGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCustomerGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCustomerGatewaysInput) ([]awstypes.CustomerGateway, error) {
	output, err := conn.DescribeCustomerGateways(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidCustomerGatewayIDNotFound) {
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

	return output.CustomerGateways, nil
}

func findCustomerGatewayByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.CustomerGateway, error) {
	input := &ec2.DescribeCustomerGatewaysInput{
		CustomerGatewayIds: []string{id},
	}

	output, err := findCustomerGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.ToString(output.State); state == customerGatewayStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.CustomerGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAM(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamsInput) (*awstypes.Ipam, error) {
	output, err := findIPAMs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamsInput) ([]awstypes.Ipam, error) {
	var output []awstypes.Ipam

	pages := ec2.NewDescribeIpamsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Ipams...)
	}

	return output, nil
}

func findIPAMByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Ipam, error) {
	input := &ec2.DescribeIpamsInput{
		IpamIds: []string{id},
	}

	output, err := findIPAM(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.IpamStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.IpamId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMPool(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamPoolsInput) (*awstypes.IpamPool, error) {
	output, err := findIPAMPools(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMPools(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamPoolsInput) ([]awstypes.IpamPool, error) {
	var output []awstypes.IpamPool

	pages := ec2.NewDescribeIpamPoolsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpamPools...)
	}

	return output, nil
}

func findIPAMPoolByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.IpamPool, error) {
	input := &ec2.DescribeIpamPoolsInput{
		IpamPoolIds: []string{id},
	}

	output, err := findIPAMPool(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.IpamPoolStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.IpamPoolId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMPoolAllocation(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolAllocationsInput) (*awstypes.IpamPoolAllocation, error) {
	output, err := findIPAMPoolAllocations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMPoolAllocations(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolAllocationsInput) ([]awstypes.IpamPoolAllocation, error) {
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

func findIPAMPoolAllocationByTwoPartKey(ctx context.Context, conn *ec2.Client, allocationID, poolID string) (*awstypes.IpamPoolAllocation, error) {
	input := &ec2.GetIpamPoolAllocationsInput{
		IpamPoolAllocationId: aws.String(allocationID),
		IpamPoolId:           aws.String(poolID),
	}

	output, err := findIPAMPoolAllocation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.IpamPoolAllocationId) != allocationID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMPoolCIDR(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolCidrsInput) (*awstypes.IpamPoolCidr, error) {
	output, err := findIPAMPoolCIDRs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMPoolCIDRs(ctx context.Context, conn *ec2.Client, input *ec2.GetIpamPoolCidrsInput) ([]awstypes.IpamPoolCidr, error) {
	var output []awstypes.IpamPoolCidr

	pages := ec2.NewGetIpamPoolCidrsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMPoolIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpamPoolCidrs...)
	}

	return output, nil
}

func findIPAMPoolCIDRByTwoPartKey(ctx context.Context, conn *ec2.Client, cidrBlock, poolID string) (*awstypes.IpamPoolCidr, error) {
	input := &ec2.GetIpamPoolCidrsInput{
		Filters: newAttributeFilterList(map[string]string{
			"cidr": cidrBlock,
		}),
		IpamPoolId: aws.String(poolID),
	}

	output, err := findIPAMPoolCIDR(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.IpamPoolCidrStateDeprovisioned {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.Cidr) != cidrBlock {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMPoolCIDRByPoolCIDRIDAndPoolID(ctx context.Context, conn *ec2.Client, poolCIDRID, poolID string) (*awstypes.IpamPoolCidr, error) {
	input := &ec2.GetIpamPoolCidrsInput{
		Filters: newAttributeFilterList(map[string]string{
			"ipam-pool-cidr-id": poolCIDRID,
		}),
		IpamPoolId: aws.String(poolID),
	}

	output, err := findIPAMPoolCIDR(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check
	if aws.ToString(output.Cidr) == "" {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	if state := output.State; state == awstypes.IpamPoolCidrStateDeprovisioned {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMResourceDiscovery(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamResourceDiscoveriesInput) (*awstypes.IpamResourceDiscovery, error) {
	output, err := findIPAMResourceDiscoveries(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMResourceDiscoveries(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamResourceDiscoveriesInput) ([]awstypes.IpamResourceDiscovery, error) {
	var output []awstypes.IpamResourceDiscovery

	pages := ec2.NewDescribeIpamResourceDiscoveriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMResourceDiscoveryIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpamResourceDiscoveries...)
	}

	return output, nil
}

func findIPAMResourceDiscoveryByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.IpamResourceDiscovery, error) {
	input := &ec2.DescribeIpamResourceDiscoveriesInput{
		IpamResourceDiscoveryIds: []string{id},
	}

	output, err := findIPAMResourceDiscovery(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.IpamResourceDiscoveryStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.IpamResourceDiscoveryId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMResourceDiscoveryAssociation(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamResourceDiscoveryAssociationsInput) (*awstypes.IpamResourceDiscoveryAssociation, error) {
	output, err := findIPAMResourceDiscoveryAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMResourceDiscoveryAssociations(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamResourceDiscoveryAssociationsInput) ([]awstypes.IpamResourceDiscoveryAssociation, error) {
	var output []awstypes.IpamResourceDiscoveryAssociation

	pages := ec2.NewDescribeIpamResourceDiscoveryAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMResourceDiscoveryAssociationIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpamResourceDiscoveryAssociations...)
	}

	return output, nil
}

func findIPAMResourceDiscoveryAssociationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.IpamResourceDiscoveryAssociation, error) {
	input := &ec2.DescribeIpamResourceDiscoveryAssociationsInput{
		IpamResourceDiscoveryAssociationIds: []string{id},
	}

	output, err := findIPAMResourceDiscoveryAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.IpamResourceDiscoveryAssociationStateDisassociateComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.IpamResourceDiscoveryAssociationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findIPAMScope(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamScopesInput) (*awstypes.IpamScope, error) {
	output, err := findIPAMScopes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findIPAMScopes(ctx context.Context, conn *ec2.Client, input *ec2.DescribeIpamScopesInput) ([]awstypes.IpamScope, error) {
	var output []awstypes.IpamScope

	pages := ec2.NewDescribeIpamScopesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMScopeIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.IpamScopes...)
	}

	return output, nil
}

func findIPAMScopeByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.IpamScope, error) {
	input := &ec2.DescribeIpamScopesInput{
		IpamScopeIds: []string{id},
	}

	output, err := findIPAMScope(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.IpamScopeStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.IpamScopeId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findImages(ctx context.Context, conn *ec2.Client, input *ec2.DescribeImagesInput) ([]awstypes.Image, error) {
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

func findImage(ctx context.Context, conn *ec2.Client, input *ec2.DescribeImagesInput) (*awstypes.Image, error) {
	output, err := findImages(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findImageByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Image, error) {
	input := &ec2.DescribeImagesInput{
		ImageIds: []string{id},
	}

	output, err := findImage(ctx, conn, input)

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

func findImageAttribute(ctx context.Context, conn *ec2.Client, input *ec2.DescribeImageAttributeInput) (*ec2.DescribeImageAttributeOutput, error) {
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

func findImageBlockPublicAccessState(ctx context.Context, conn *ec2.Client) (*string, error) {
	input := &ec2.GetImageBlockPublicAccessStateInput{}
	output, err := conn.GetImageBlockPublicAccessState(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.ImageBlockPublicAccessState == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ImageBlockPublicAccessState, nil
}

func findImageLaunchPermissionsByID(ctx context.Context, conn *ec2.Client, id string) ([]awstypes.LaunchPermission, error) {
	input := &ec2.DescribeImageAttributeInput{
		Attribute: awstypes.ImageAttributeNameLaunchPermission,
		ImageId:   aws.String(id),
	}

	output, err := findImageAttribute(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output.LaunchPermissions) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.LaunchPermissions, nil
}

func findImageLaunchPermission(ctx context.Context, conn *ec2.Client, imageID, accountID, group, organizationARN, organizationalUnitARN string) (*awstypes.LaunchPermission, error) {
	output, err := findImageLaunchPermissionsByID(ctx, conn, imageID)

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

func findTransitGateway(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewaysInput) (*awstypes.TransitGateway, error) {
	output, err := findTransitGateways(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.TransitGateway) bool { return v.Options != nil })
}

func findTransitGateways(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewaysInput) ([]awstypes.TransitGateway, error) {
	var output []awstypes.TransitGateway

	pages := ec2.NewDescribeTransitGatewaysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGateways...)
	}

	return output, nil
}

func findTransitGatewayByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGateway, error) {
	input := &ec2.DescribeTransitGatewaysInput{
		TransitGatewayIds: []string{id},
	}

	output, err := findTransitGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayAttachment(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayAttachmentsInput) (*awstypes.TransitGatewayAttachment, error) {
	output, err := findTransitGatewayAttachments(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayAttachments(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayAttachmentsInput) ([]awstypes.TransitGatewayAttachment, error) {
	var output []awstypes.TransitGatewayAttachment

	pages := ec2.NewDescribeTransitGatewayAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayAttachments...)
	}

	return output, nil
}

func findTransitGatewayAttachmentByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayAttachment, error) {
	input := &ec2.DescribeTransitGatewayAttachmentsInput{
		TransitGatewayAttachmentIds: []string{id},
	}

	output, err := findTransitGatewayAttachment(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayConnect(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayConnectsInput) (*awstypes.TransitGatewayConnect, error) {
	output, err := findTransitGatewayConnects(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.TransitGatewayConnect) bool { return v.Options != nil })
}

func findTransitGatewayConnects(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayConnectsInput) ([]awstypes.TransitGatewayConnect, error) {
	var output []awstypes.TransitGatewayConnect

	pages := ec2.NewDescribeTransitGatewayConnectsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayConnects...)
	}

	return output, nil
}

func findTransitGatewayConnectByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayConnect, error) {
	input := &ec2.DescribeTransitGatewayConnectsInput{
		TransitGatewayAttachmentIds: []string{id},
	}

	output, err := findTransitGatewayConnect(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayAttachmentStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayConnectPeer(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayConnectPeersInput) (*awstypes.TransitGatewayConnectPeer, error) {
	output, err := findTransitGatewayConnectPeers(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output,
		func(v *awstypes.TransitGatewayConnectPeer) bool { return v.ConnectPeerConfiguration != nil },
		func(v *awstypes.TransitGatewayConnectPeer) bool {
			return len(v.ConnectPeerConfiguration.BgpConfigurations) > 0
		},
	)
}

func findTransitGatewayConnectPeers(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayConnectPeersInput) ([]awstypes.TransitGatewayConnectPeer, error) {
	var output []awstypes.TransitGatewayConnectPeer

	pages := ec2.NewDescribeTransitGatewayConnectPeersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayConnectPeerIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayConnectPeers...)
	}

	return output, nil
}

func findTransitGatewayConnectPeerByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayConnectPeer, error) {
	input := &ec2.DescribeTransitGatewayConnectPeersInput{
		TransitGatewayConnectPeerIds: []string{id},
	}

	output, err := findTransitGatewayConnectPeer(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayConnectPeerStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayConnectPeerId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayMulticastDomain(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayMulticastDomainsInput) (*awstypes.TransitGatewayMulticastDomain, error) {
	output, err := findTransitGatewayMulticastDomains(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.TransitGatewayMulticastDomain) bool { return v.Options != nil })
}

func findTransitGatewayMulticastDomains(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayMulticastDomainsInput) ([]awstypes.TransitGatewayMulticastDomain, error) {
	var output []awstypes.TransitGatewayMulticastDomain

	pages := ec2.NewDescribeTransitGatewayMulticastDomainsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayMulticastDomains...)
	}

	return output, nil
}

func findTransitGatewayMulticastDomainByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayMulticastDomain, error) {
	input := &ec2.DescribeTransitGatewayMulticastDomainsInput{
		TransitGatewayMulticastDomainIds: []string{id},
	}

	output, err := findTransitGatewayMulticastDomain(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayMulticastDomainStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayMulticastDomainId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayMulticastDomainAssociation(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayMulticastDomainAssociationsInput) (*awstypes.TransitGatewayMulticastDomainAssociation, error) {
	output, err := findTransitGatewayMulticastDomainAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.TransitGatewayMulticastDomainAssociation) bool { return v.Subnet != nil })
}

func findTransitGatewayMulticastDomainAssociations(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayMulticastDomainAssociationsInput) ([]awstypes.TransitGatewayMulticastDomainAssociation, error) {
	var output []awstypes.TransitGatewayMulticastDomainAssociation

	pages := ec2.NewGetTransitGatewayMulticastDomainAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.MulticastDomainAssociations...)
	}

	return output, nil
}

func findTransitGatewayMulticastDomainAssociationByThreePartKey(ctx context.Context, conn *ec2.Client, multicastDomainID, attachmentID, subnetID string) (*awstypes.TransitGatewayMulticastDomainAssociation, error) {
	input := &ec2.GetTransitGatewayMulticastDomainAssociationsInput{
		Filters: newAttributeFilterList(map[string]string{
			"subnet-id":                     subnetID,
			"transit-gateway-attachment-id": attachmentID,
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := findTransitGatewayMulticastDomainAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.Subnet.State; state == awstypes.TransitGatewayMulitcastDomainAssociationStateDisassociated {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != attachmentID || aws.ToString(output.Subnet.SubnetId) != subnetID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayMulticastGroups(ctx context.Context, conn *ec2.Client, input *ec2.SearchTransitGatewayMulticastGroupsInput) ([]awstypes.TransitGatewayMulticastGroup, error) {
	var output []awstypes.TransitGatewayMulticastGroup

	pages := ec2.NewSearchTransitGatewayMulticastGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayMulticastDomainIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.MulticastGroups...)
	}

	return output, nil
}

func findTransitGatewayMulticastGroupMemberByThreePartKey(ctx context.Context, conn *ec2.Client, multicastDomainID, groupIPAddress, eniID string) (*awstypes.TransitGatewayMulticastGroup, error) {
	input := &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: newAttributeFilterList(map[string]string{
			"group-ip-address": groupIPAddress,
			"is-group-member":  "true",
			"is-group-source":  "false",
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := findTransitGatewayMulticastGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, v := range output {
		if aws.ToString(v.NetworkInterfaceId) == eniID {
			// Eventual consistency check.
			if aws.ToString(v.GroupIpAddress) != groupIPAddress || !aws.ToBool(v.GroupMember) {
				return nil, &retry.NotFoundError{
					LastRequest: input,
				}
			}

			return &v, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func findTransitGatewayMulticastGroupSourceByThreePartKey(ctx context.Context, conn *ec2.Client, multicastDomainID, groupIPAddress, eniID string) (*awstypes.TransitGatewayMulticastGroup, error) {
	input := &ec2.SearchTransitGatewayMulticastGroupsInput{
		Filters: newAttributeFilterList(map[string]string{
			"group-ip-address": groupIPAddress,
			"is-group-member":  "false",
			"is-group-source":  "true",
		}),
		TransitGatewayMulticastDomainId: aws.String(multicastDomainID),
	}

	output, err := findTransitGatewayMulticastGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, v := range output {
		if aws.ToString(v.NetworkInterfaceId) == eniID {
			// Eventual consistency check.
			if aws.ToString(v.GroupIpAddress) != groupIPAddress || !aws.ToBool(v.GroupSource) {
				return nil, &retry.NotFoundError{
					LastRequest: input,
				}
			}

			return &v, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func findTransitGatewayPeeringAttachment(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayPeeringAttachmentsInput) (*awstypes.TransitGatewayPeeringAttachment, error) {
	output, err := findTransitGatewayPeeringAttachments(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output,
		func(v *awstypes.TransitGatewayPeeringAttachment) bool { return v.AccepterTgwInfo != nil },
		func(v *awstypes.TransitGatewayPeeringAttachment) bool { return v.RequesterTgwInfo != nil },
	)
}

func findTransitGatewayPeeringAttachments(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayPeeringAttachmentsInput) ([]awstypes.TransitGatewayPeeringAttachment, error) {
	var output []awstypes.TransitGatewayPeeringAttachment

	pages := ec2.NewDescribeTransitGatewayPeeringAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayPeeringAttachments...)
	}

	return output, nil
}

func findTransitGatewayPeeringAttachmentByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayPeeringAttachment, error) {
	input := &ec2.DescribeTransitGatewayPeeringAttachmentsInput{
		TransitGatewayAttachmentIds: []string{id},
	}

	output, err := findTransitGatewayPeeringAttachment(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/vpc/latest/tgw/tgw-vpc-attachments.html#vpc-attachment-lifecycle.
	switch state := output.State; state {
	case awstypes.TransitGatewayAttachmentStateDeleted,
		awstypes.TransitGatewayAttachmentStateFailed,
		awstypes.TransitGatewayAttachmentStateRejected:
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayPrefixListReference(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayPrefixListReferencesInput) (*awstypes.TransitGatewayPrefixListReference, error) {
	output, err := findTransitGatewayPrefixListReferences(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayPrefixListReferences(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayPrefixListReferencesInput) ([]awstypes.TransitGatewayPrefixListReference, error) {
	var output []awstypes.TransitGatewayPrefixListReference

	pages := ec2.NewGetTransitGatewayPrefixListReferencesPaginator(conn, input)
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

		output = append(output, page.TransitGatewayPrefixListReferences...)
	}

	return output, nil
}

func findTransitGatewayPrefixListReferenceByTwoPartKey(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, prefixListID string) (*awstypes.TransitGatewayPrefixListReference, error) {
	input := &ec2.GetTransitGatewayPrefixListReferencesInput{
		Filters: newAttributeFilterList(map[string]string{
			"prefix-list-id": prefixListID,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := findTransitGatewayPrefixListReference(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.PrefixListId) != prefixListID || aws.ToString(output.TransitGatewayRouteTableId) != transitGatewayRouteTableID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayStaticRoute(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, destination string) (*awstypes.TransitGatewayRoute, error) {
	input := &ec2.SearchTransitGatewayRoutesInput{
		Filters: newAttributeFilterList(map[string]string{
			names.AttrType:             string(awstypes.TransitGatewayRouteTypeStatic),
			"route-search.exact-match": destination,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := findTransitGatewayRoutes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, route := range output {
		if v := aws.ToString(route.DestinationCidrBlock); types.CIDRBlocksEqual(v, destination) {
			if state := route.State; state == awstypes.TransitGatewayRouteStateDeleted {
				return nil, &retry.NotFoundError{
					Message:     string(state),
					LastRequest: input,
				}
			}

			route.DestinationCidrBlock = aws.String(types.CanonicalCIDRBlock(v))

			return &route, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func findTransitGatewayRoutes(ctx context.Context, conn *ec2.Client, input *ec2.SearchTransitGatewayRoutesInput) ([]awstypes.TransitGatewayRoute, error) {
	output, err := conn.SearchTransitGatewayRoutes(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
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

	return output.Routes, err
}

func findTransitGatewayPolicyTable(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayPolicyTablesInput) (*awstypes.TransitGatewayPolicyTable, error) {
	output, err := findTransitGatewayPolicyTables(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayRouteTable(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayRouteTablesInput) (*awstypes.TransitGatewayRouteTable, error) {
	output, err := findTransitGatewayRouteTables(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayPolicyTables(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayPolicyTablesInput) ([]awstypes.TransitGatewayPolicyTable, error) {
	var output []awstypes.TransitGatewayPolicyTable

	pages := ec2.NewDescribeTransitGatewayPolicyTablesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayPolicyTables...)
	}

	return output, nil
}

func findTransitGatewayRouteTables(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayRouteTablesInput) ([]awstypes.TransitGatewayRouteTable, error) {
	var output []awstypes.TransitGatewayRouteTable

	pages := ec2.NewDescribeTransitGatewayRouteTablesPaginator(conn, input)
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

		output = append(output, page.TransitGatewayRouteTables...)
	}

	return output, nil
}

func findTransitGatewayPolicyTableByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayPolicyTable, error) {
	input := &ec2.DescribeTransitGatewayPolicyTablesInput{
		TransitGatewayPolicyTableIds: []string{id},
	}

	output, err := findTransitGatewayPolicyTable(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayPolicyTableId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayRouteTableByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayRouteTable, error) {
	input := &ec2.DescribeTransitGatewayRouteTablesInput{
		TransitGatewayRouteTableIds: []string{id},
	}

	output, err := findTransitGatewayRouteTable(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayRouteTableStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayRouteTableId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTransitGatewayPolicyTableAssociationByTwoPartKey(ctx context.Context, conn *ec2.Client, transitGatewayPolicyTableID, transitGatewayAttachmentID string) (*awstypes.TransitGatewayPolicyTableAssociation, error) {
	input := &ec2.GetTransitGatewayPolicyTableAssociationsInput{
		Filters: newAttributeFilterList(map[string]string{
			"transit-gateway-attachment-id": transitGatewayAttachmentID,
		}),
		TransitGatewayPolicyTableId: aws.String(transitGatewayPolicyTableID),
	}

	output, err := findTransitGatewayPolicyTableAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayAssociationStateDisassociated {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != transitGatewayAttachmentID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, err
}

func findTransitGatewayRouteTableAssociationByTwoPartKey(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, transitGatewayAttachmentID string) (*awstypes.TransitGatewayRouteTableAssociation, error) {
	input := &ec2.GetTransitGatewayRouteTableAssociationsInput{
		Filters: newAttributeFilterList(map[string]string{
			"transit-gateway-attachment-id": transitGatewayAttachmentID,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := findTransitGatewayRouteTableAssociation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayAssociationStateDisassociated {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != transitGatewayAttachmentID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, err
}

func findTransitGatewayRouteTableAssociation(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayRouteTableAssociationsInput) (*awstypes.TransitGatewayRouteTableAssociation, error) {
	output, err := findTransitGatewayRouteTableAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayPolicyTableAssociations(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayPolicyTableAssociationsInput) ([]awstypes.TransitGatewayPolicyTableAssociation, error) {
	var output []awstypes.TransitGatewayPolicyTableAssociation

	pages := ec2.NewGetTransitGatewayPolicyTableAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayPolicyTableIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Associations...)
	}

	return output, nil
}

func findTransitGatewayPolicyTableAssociation(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayPolicyTableAssociationsInput) (*awstypes.TransitGatewayPolicyTableAssociation, error) {
	output, err := findTransitGatewayPolicyTableAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayRouteTableAssociations(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayRouteTableAssociationsInput) ([]awstypes.TransitGatewayRouteTableAssociation, error) {
	var output []awstypes.TransitGatewayRouteTableAssociation

	pages := ec2.NewGetTransitGatewayRouteTableAssociationsPaginator(conn, input)
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

		output = append(output, page.Associations...)
	}

	return output, nil
}

func findTransitGatewayRouteTablePropagationByTwoPartKey(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID string, transitGatewayAttachmentID string) (*awstypes.TransitGatewayRouteTablePropagation, error) {
	input := &ec2.GetTransitGatewayRouteTablePropagationsInput{
		Filters: newAttributeFilterList(map[string]string{
			"transit-gateway-attachment-id": transitGatewayAttachmentID,
		}),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	output, err := findTransitGatewayRouteTablePropagation(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.TransitGatewayPropagationStateDisabled {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != transitGatewayAttachmentID {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, err
}

func findTransitGatewayRouteTablePropagation(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayRouteTablePropagationsInput) (*awstypes.TransitGatewayRouteTablePropagation, error) {
	output, err := findTransitGatewayRouteTablePropagations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTransitGatewayRouteTablePropagations(ctx context.Context, conn *ec2.Client, input *ec2.GetTransitGatewayRouteTablePropagationsInput) ([]awstypes.TransitGatewayRouteTablePropagation, error) {
	var output []awstypes.TransitGatewayRouteTablePropagation

	pages := ec2.NewGetTransitGatewayRouteTablePropagationsPaginator(conn, input)
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

		output = append(output, page.TransitGatewayRouteTablePropagations...)
	}

	return output, nil
}

func findTransitGatewayVPCAttachment(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayVpcAttachmentsInput) (*awstypes.TransitGatewayVpcAttachment, error) {
	output, err := findTransitGatewayVPCAttachments(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.TransitGatewayVpcAttachment) bool { return v.Options != nil })
}

func findTransitGatewayVPCAttachments(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTransitGatewayVpcAttachmentsInput) ([]awstypes.TransitGatewayVpcAttachment, error) {
	var output []awstypes.TransitGatewayVpcAttachment

	pages := ec2.NewDescribeTransitGatewayVpcAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TransitGatewayVpcAttachments...)
	}

	return output, nil
}

func findTransitGatewayVPCAttachmentByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayVpcAttachment, error) {
	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{
		TransitGatewayAttachmentIds: []string{id},
	}

	output, err := findTransitGatewayVPCAttachment(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// See https://docs.aws.amazon.com/vpc/latest/tgw/tgw-vpc-attachments.html#vpc-attachment-lifecycle.
	switch state := output.State; state {
	case awstypes.TransitGatewayAttachmentStateDeleted,
		awstypes.TransitGatewayAttachmentStateFailed,
		awstypes.TransitGatewayAttachmentStateRejected:
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.TransitGatewayAttachmentId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findEIPs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeAddressesInput) ([]awstypes.Address, error) {
	output, err := conn.DescribeAddresses(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAddressNotFound, errCodeInvalidAllocationIDNotFound) ||
		tfawserr.ErrMessageContains(err, errCodeAuthFailure, "does not belong to you") {
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

	return output.Addresses, nil
}

func findEIP(ctx context.Context, conn *ec2.Client, input *ec2.DescribeAddressesInput) (*awstypes.Address, error) {
	output, err := findEIPs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEIPByAllocationID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Address, error) {
	input := &ec2.DescribeAddressesInput{
		AllocationIds: []string{id},
	}

	output, err := findEIP(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.AllocationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findEIPByAssociationID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Address, error) {
	input := &ec2.DescribeAddressesInput{
		Filters: newAttributeFilterList(map[string]string{
			"association-id": id,
		}),
	}

	output, err := findEIP(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.AssociationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findEIPAttributes(ctx context.Context, conn *ec2.Client, input *ec2.DescribeAddressesAttributeInput) ([]awstypes.AddressAttribute, error) {
	var output []awstypes.AddressAttribute

	pages := ec2.NewDescribeAddressesAttributePaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Addresses...)
	}

	return output, nil
}

func findEIPAttribute(ctx context.Context, conn *ec2.Client, input *ec2.DescribeAddressesAttributeInput) (*awstypes.AddressAttribute, error) {
	output, err := findEIPAttributes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findEIPDomainNameAttributeByAllocationID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.AddressAttribute, error) {
	input := &ec2.DescribeAddressesAttributeInput{
		AllocationIds: []string{id},
		Attribute:     awstypes.AddressAttributeNameDomainName,
	}

	output, err := findEIPAttribute(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.AllocationId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findKeyPair(ctx context.Context, conn *ec2.Client, input *ec2.DescribeKeyPairsInput) (*awstypes.KeyPairInfo, error) {
	output, err := findKeyPairs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findKeyPairs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeKeyPairsInput) ([]awstypes.KeyPairInfo, error) {
	output, err := conn.DescribeKeyPairs(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidKeyPairNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output.KeyPairs, nil
}

func findKeyPairByName(ctx context.Context, conn *ec2.Client, name string) (*awstypes.KeyPairInfo, error) {
	input := &ec2.DescribeKeyPairsInput{
		KeyNames: []string{name},
	}

	output, err := findKeyPair(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.KeyName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findImportSnapshotTasks(ctx context.Context, conn *ec2.Client, input *ec2.DescribeImportSnapshotTasksInput) ([]awstypes.ImportSnapshotTask, error) {
	var output []awstypes.ImportSnapshotTask

	pages := ec2.NewDescribeImportSnapshotTasksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidConversionTaskIdMalformed, "not found") {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.ImportSnapshotTasks...)
	}

	return output, nil
}

func findImportSnapshotTask(ctx context.Context, conn *ec2.Client, input *ec2.DescribeImportSnapshotTasksInput) (*awstypes.ImportSnapshotTask, error) {
	output, err := findImportSnapshotTasks(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output, func(v *awstypes.ImportSnapshotTask) bool { return v.SnapshotTaskDetail != nil })
}

func findImportSnapshotTaskByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ImportSnapshotTask, error) {
	input := &ec2.DescribeImportSnapshotTasksInput{
		ImportTaskIds: []string{id},
	}

	output, err := findImportSnapshotTask(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.ImportTaskId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findSnapshots(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSnapshotsInput) ([]awstypes.Snapshot, error) {
	var output []awstypes.Snapshot

	pages := ec2.NewDescribeSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidSnapshotNotFound) {
				return nil, &retry.NotFoundError{
					LastError:   err,
					LastRequest: input,
				}
			}
			return nil, err
		}

		output = append(output, page.Snapshots...)
	}

	return output, nil
}

func findSnapshot(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSnapshotsInput) (*awstypes.Snapshot, error) {
	output, err := findSnapshots(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSnapshotByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Snapshot, error) {
	input := &ec2.DescribeSnapshotsInput{
		SnapshotIds: []string{id},
	}

	output, err := findSnapshot(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.SnapshotId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findSnapshotAttribute(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSnapshotAttributeInput) (*ec2.DescribeSnapshotAttributeOutput, error) {
	output, err := conn.DescribeSnapshotAttribute(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidSnapshotNotFound) {
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

func findCreateSnapshotCreateVolumePermissionByTwoPartKey(ctx context.Context, conn *ec2.Client, snapshotID, accountID string) (awstypes.CreateVolumePermission, error) {
	input := &ec2.DescribeSnapshotAttributeInput{
		Attribute:  awstypes.SnapshotAttributeNameCreateVolumePermission,
		SnapshotId: aws.String(snapshotID),
	}

	output, err := findSnapshotAttribute(ctx, conn, input)

	if err != nil {
		return awstypes.CreateVolumePermission{}, err
	}

	for _, v := range output.CreateVolumePermissions {
		if aws.ToString(v.UserId) == accountID {
			return v, nil
		}
	}

	return awstypes.CreateVolumePermission{}, &retry.NotFoundError{LastRequest: input}
}

func findFindSnapshotTierStatuses(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSnapshotTierStatusInput) ([]awstypes.SnapshotTierStatus, error) {
	var output []awstypes.SnapshotTierStatus

	pages := ec2.NewDescribeSnapshotTierStatusPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.SnapshotTierStatuses...)
	}

	return output, nil
}

func findFindSnapshotTierStatus(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSnapshotTierStatusInput) (*awstypes.SnapshotTierStatus, error) {
	output, err := findFindSnapshotTierStatuses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFlowLogByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.FlowLog, error) {
	input := &ec2.DescribeFlowLogsInput{
		FlowLogIds: []string{id},
	}

	output, err := findFlowLog(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.FlowLogId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findFlowLogs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeFlowLogsInput) ([]awstypes.FlowLog, error) {
	var output []awstypes.FlowLog

	pages := ec2.NewDescribeFlowLogsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.FlowLogs...)
	}

	return output, nil
}

func findFlowLog(ctx context.Context, conn *ec2.Client, input *ec2.DescribeFlowLogsInput) (*awstypes.FlowLog, error) {
	output, err := findFlowLogs(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSnapshotTierStatusBySnapshotID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.SnapshotTierStatus, error) {
	input := &ec2.DescribeSnapshotTierStatusInput{
		Filters: newAttributeFilterList(map[string]string{
			"snapshot-id": id,
		}),
	}

	output, err := findFindSnapshotTierStatus(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.SnapshotId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findNetworkPerformanceMetricSubscriptions(ctx context.Context, conn *ec2.Client, input *ec2.DescribeAwsNetworkPerformanceMetricSubscriptionsInput) ([]awstypes.Subscription, error) {
	var output []awstypes.Subscription

	pages := ec2.NewDescribeAwsNetworkPerformanceMetricSubscriptionsPaginator(conn, input, func(o *ec2.DescribeAwsNetworkPerformanceMetricSubscriptionsPaginatorOptions) {
		o.Limit = 100
	})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Subscriptions...)
	}

	return output, nil
}

func findNetworkPerformanceMetricSubscriptionByFourPartKey(ctx context.Context, conn *ec2.Client, source, destination, metric, statistic string) (*awstypes.Subscription, error) {
	input := &ec2.DescribeAwsNetworkPerformanceMetricSubscriptionsInput{}

	output, err := findNetworkPerformanceMetricSubscriptions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if aws.ToString(v.Source) == source && aws.ToString(v.Destination) == destination && string(v.Metric) == metric && string(v.Statistic) == statistic {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func findInstanceConnectEndpoint(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstanceConnectEndpointsInput) (*awstypes.Ec2InstanceConnectEndpoint, error) {
	output, err := findInstanceConnectEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInstanceConnectEndpoints(ctx context.Context, conn *ec2.Client, input *ec2.DescribeInstanceConnectEndpointsInput) ([]awstypes.Ec2InstanceConnectEndpoint, error) {
	var output []awstypes.Ec2InstanceConnectEndpoint

	pages := ec2.NewDescribeInstanceConnectEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceConnectEndpointIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.InstanceConnectEndpoints...)
	}

	return output, nil
}

func findInstanceConnectEndpointByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Ec2InstanceConnectEndpoint, error) {
	input := &ec2.DescribeInstanceConnectEndpointsInput{
		InstanceConnectEndpointIds: []string{id},
	}
	output, err := findInstanceConnectEndpoint(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.Ec2InstanceConnectEndpointStateDeleteComplete {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.InstanceConnectEndpointId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVerifiedAccessGroupPolicyByID(ctx context.Context, conn *ec2.Client, id string) (*ec2.GetVerifiedAccessGroupPolicyOutput, error) {
	input := &ec2.GetVerifiedAccessGroupPolicyInput{
		VerifiedAccessGroupId: &id,
	}
	output, err := conn.GetVerifiedAccessGroupPolicy(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessGroupIdNotFound) {
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

func findVerifiedAccessEndpointPolicyByID(ctx context.Context, conn *ec2.Client, id string) (*ec2.GetVerifiedAccessEndpointPolicyOutput, error) {
	input := &ec2.GetVerifiedAccessEndpointPolicyInput{
		VerifiedAccessEndpointId: &id,
	}
	output, err := conn.GetVerifiedAccessEndpointPolicy(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessEndpointIdNotFound) {
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

func findVerifiedAccessGroup(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVerifiedAccessGroupsInput) (*awstypes.VerifiedAccessGroup, error) {
	output, err := findVerifiedAccessGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVerifiedAccessGroups(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVerifiedAccessGroupsInput) ([]awstypes.VerifiedAccessGroup, error) {
	var output []awstypes.VerifiedAccessGroup

	pages := ec2.NewDescribeVerifiedAccessGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessGroupIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.VerifiedAccessGroups...)
	}

	return output, nil
}

func findVerifiedAccessGroupByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VerifiedAccessGroup, error) {
	input := &ec2.DescribeVerifiedAccessGroupsInput{
		VerifiedAccessGroupIds: []string{id},
	}
	output, err := findVerifiedAccessGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.VerifiedAccessGroupId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVerifiedAccessInstance(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVerifiedAccessInstancesInput) (*awstypes.VerifiedAccessInstance, error) {
	output, err := findVerifiedAccessInstances(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVerifiedAccessInstances(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVerifiedAccessInstancesInput) ([]awstypes.VerifiedAccessInstance, error) {
	var output []awstypes.VerifiedAccessInstance

	pages := ec2.NewDescribeVerifiedAccessInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessInstanceIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.VerifiedAccessInstances...)
	}

	return output, nil
}

func findVerifiedAccessInstanceByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VerifiedAccessInstance, error) {
	input := &ec2.DescribeVerifiedAccessInstancesInput{
		VerifiedAccessInstanceIds: []string{id},
	}
	output, err := findVerifiedAccessInstance(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.VerifiedAccessInstanceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVerifiedAccessInstanceLoggingConfiguration(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVerifiedAccessInstanceLoggingConfigurationsInput) (*awstypes.VerifiedAccessInstanceLoggingConfiguration, error) {
	output, err := findVerifiedAccessInstanceLoggingConfigurations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVerifiedAccessInstanceLoggingConfigurations(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVerifiedAccessInstanceLoggingConfigurationsInput) ([]awstypes.VerifiedAccessInstanceLoggingConfiguration, error) {
	var output []awstypes.VerifiedAccessInstanceLoggingConfiguration

	pages := ec2.NewDescribeVerifiedAccessInstanceLoggingConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessInstanceIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.LoggingConfigurations...)
	}

	return output, nil
}

func findVerifiedAccessInstanceLoggingConfigurationByInstanceID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VerifiedAccessInstanceLoggingConfiguration, error) {
	input := &ec2.DescribeVerifiedAccessInstanceLoggingConfigurationsInput{
		VerifiedAccessInstanceIds: []string{id},
	}
	output, err := findVerifiedAccessInstanceLoggingConfiguration(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.VerifiedAccessInstanceId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVerifiedAccessInstanceTrustProviderAttachmentExists(ctx context.Context, conn *ec2.Client, vaiID, vatpID string) error {
	output, err := findVerifiedAccessInstanceByID(ctx, conn, vaiID)

	if err != nil {
		return err
	}

	for _, v := range output.VerifiedAccessTrustProviders {
		if aws.ToString(v.VerifiedAccessTrustProviderId) == vatpID {
			return nil
		}
	}

	return &retry.NotFoundError{
		LastError: fmt.Errorf("Verified Access Instance (%s) Trust Provider (%s) Attachment not found", vaiID, vatpID),
	}
}

func findVerifiedAccessTrustProvider(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVerifiedAccessTrustProvidersInput) (*awstypes.VerifiedAccessTrustProvider, error) {
	output, err := findVerifiedAccessTrustProviders(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVerifiedAccessTrustProviders(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVerifiedAccessTrustProvidersInput) ([]awstypes.VerifiedAccessTrustProvider, error) {
	var output []awstypes.VerifiedAccessTrustProvider

	pages := ec2.NewDescribeVerifiedAccessTrustProvidersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessTrustProviderIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.VerifiedAccessTrustProviders...)
	}

	return output, nil
}

func findVerifiedAccessTrustProviderByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VerifiedAccessTrustProvider, error) {
	input := &ec2.DescribeVerifiedAccessTrustProvidersInput{
		VerifiedAccessTrustProviderIds: []string{id},
	}
	output, err := findVerifiedAccessTrustProvider(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.VerifiedAccessTrustProviderId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findVerifiedAccessEndpoint(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVerifiedAccessEndpointsInput) (*awstypes.VerifiedAccessEndpoint, error) {
	output, err := findVerifiedAccessEndpoints(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findVerifiedAccessEndpoints(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVerifiedAccessEndpointsInput) ([]awstypes.VerifiedAccessEndpoint, error) {
	var output []awstypes.VerifiedAccessEndpoint

	pages := ec2.NewDescribeVerifiedAccessEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessEndpointIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.VerifiedAccessEndpoints...)
	}

	return output, nil
}

func findVerifiedAccessEndpointByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VerifiedAccessEndpoint, error) {
	input := &ec2.DescribeVerifiedAccessEndpointsInput{
		VerifiedAccessEndpointIds: []string{id},
	}
	output, err := findVerifiedAccessEndpoint(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status != nil && status.Code == awstypes.VerifiedAccessEndpointStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status.Code),
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.ToString(output.VerifiedAccessEndpointId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findFastSnapshotRestore(ctx context.Context, conn *ec2.Client, input *ec2.DescribeFastSnapshotRestoresInput) (*awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	output, err := findFastSnapshotRestores(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFastSnapshotRestores(ctx context.Context, conn *ec2.Client, input *ec2.DescribeFastSnapshotRestoresInput) ([]awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	var output []awstypes.DescribeFastSnapshotRestoreSuccessItem

	pages := ec2.NewDescribeFastSnapshotRestoresPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.FastSnapshotRestores...)
	}

	return output, nil
}

func findFastSnapshotRestoreByTwoPartKey(ctx context.Context, conn *ec2.Client, availabilityZone, snapshotID string) (*awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	input := &ec2.DescribeFastSnapshotRestoresInput{
		Filters: newAttributeFilterList(map[string]string{
			"availability-zone": availabilityZone,
			"snapshot-id":       snapshotID,
		}),
	}

	output, err := findFastSnapshotRestore(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := output.State; state == awstypes.FastSnapshotRestoreStateCodeDisabled {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func findTrafficMirrorFilter(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTrafficMirrorFiltersInput) (*awstypes.TrafficMirrorFilter, error) {
	output, err := findTrafficMirrorFilters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTrafficMirrorFilters(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTrafficMirrorFiltersInput) ([]awstypes.TrafficMirrorFilter, error) {
	var output []awstypes.TrafficMirrorFilter

	pages := ec2.NewDescribeTrafficMirrorFiltersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorFilterIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TrafficMirrorFilters...)
	}

	return output, nil
}

func findTrafficMirrorFilterByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TrafficMirrorFilter, error) {
	input := &ec2.DescribeTrafficMirrorFiltersInput{
		TrafficMirrorFilterIds: []string{id},
	}

	output, err := findTrafficMirrorFilter(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.TrafficMirrorFilterId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTrafficMirrorFilterRuleByTwoPartKey(ctx context.Context, conn *ec2.Client, filterID, ruleID string) (*awstypes.TrafficMirrorFilterRule, error) {
	output, err := findTrafficMirrorFilterByID(ctx, conn, filterID)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(slices.Concat(output.IngressFilterRules, output.EgressFilterRules), func(v awstypes.TrafficMirrorFilterRule) bool {
		return aws.ToString(v.TrafficMirrorFilterRuleId) == ruleID
	}))
}

func findTrafficMirrorSession(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTrafficMirrorSessionsInput) (*awstypes.TrafficMirrorSession, error) {
	output, err := findTrafficMirrorSessions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTrafficMirrorSessions(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTrafficMirrorSessionsInput) ([]awstypes.TrafficMirrorSession, error) {
	var output []awstypes.TrafficMirrorSession

	pages := ec2.NewDescribeTrafficMirrorSessionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorSessionIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TrafficMirrorSessions...)
	}

	return output, nil
}

func findTrafficMirrorSessionByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TrafficMirrorSession, error) {
	input := &ec2.DescribeTrafficMirrorSessionsInput{
		TrafficMirrorSessionIds: []string{id},
	}

	output, err := findTrafficMirrorSession(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.TrafficMirrorSessionId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTrafficMirrorTarget(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTrafficMirrorTargetsInput) (*awstypes.TrafficMirrorTarget, error) {
	output, err := findTrafficMirrorTargets(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTrafficMirrorTargets(ctx context.Context, conn *ec2.Client, input *ec2.DescribeTrafficMirrorTargetsInput) ([]awstypes.TrafficMirrorTarget, error) {
	var output []awstypes.TrafficMirrorTarget

	pages := ec2.NewDescribeTrafficMirrorTargetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidTrafficMirrorTargetIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TrafficMirrorTargets...)
	}

	return output, nil
}

func findTrafficMirrorTargetByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TrafficMirrorTarget, error) {
	input := &ec2.DescribeTrafficMirrorTargetsInput{
		TrafficMirrorTargetIds: []string{id},
	}

	output, err := findTrafficMirrorTarget(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.TrafficMirrorTargetId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findNetworkInsightsPath(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkInsightsPathsInput) (*awstypes.NetworkInsightsPath, error) {
	output, err := findNetworkInsightsPaths(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findNetworkInsightsAnalysis(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkInsightsAnalysesInput) (*awstypes.NetworkInsightsAnalysis, error) {
	output, err := findNetworkInsightsAnalyses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findNetworkInsightsAnalyses(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkInsightsAnalysesInput) ([]awstypes.NetworkInsightsAnalysis, error) {
	var output []awstypes.NetworkInsightsAnalysis

	pages := ec2.NewDescribeNetworkInsightsAnalysesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInsightsAnalysisIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.NetworkInsightsAnalyses...)
	}

	return output, nil
}

func findNetworkInsightsAnalysisByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkInsightsAnalysis, error) {
	input := &ec2.DescribeNetworkInsightsAnalysesInput{
		NetworkInsightsAnalysisIds: []string{id},
	}

	output, err := findNetworkInsightsAnalysis(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.NetworkInsightsAnalysisId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findNetworkInsightsPaths(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkInsightsPathsInput) ([]awstypes.NetworkInsightsPath, error) {
	var output []awstypes.NetworkInsightsPath

	pages := ec2.NewDescribeNetworkInsightsPathsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInsightsPathIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.NetworkInsightsPaths...)
	}

	return output, nil
}

func findNetworkInsightsPathByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.NetworkInsightsPath, error) {
	input := &ec2.DescribeNetworkInsightsPathsInput{
		NetworkInsightsPathIds: []string{id},
	}

	output, err := findNetworkInsightsPath(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.NetworkInsightsPathId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findCapacityBlockOffering(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCapacityBlockOfferingsInput) (*awstypes.CapacityBlockOffering, error) {
	output, err := findCapacityBlockOfferings(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCapacityBlockOfferings(ctx context.Context, conn *ec2.Client, input *ec2.DescribeCapacityBlockOfferingsInput) ([]awstypes.CapacityBlockOffering, error) {
	var output []awstypes.CapacityBlockOffering

	pages := ec2.NewDescribeCapacityBlockOfferingsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInsightsAnalysisIdNotFound) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.CapacityBlockOfferings...)
	}

	return output, nil
}
