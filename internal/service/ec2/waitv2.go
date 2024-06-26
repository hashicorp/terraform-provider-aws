// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	AvailabilityZoneGroupOptInStatusTimeout = 10 * time.Minute
)

func waitAvailabilityZoneGroupOptedIn(ctx context.Context, conn *ec2.Client, name string) (*awstypes.AvailabilityZone, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AvailabilityZoneOptInStatusNotOptedIn),
		Target:  enum.Slice(awstypes.AvailabilityZoneOptInStatusOptedIn),
		Refresh: statusAvailabilityZoneGroupOptInStatus(ctx, conn, name),
		Timeout: AvailabilityZoneGroupOptInStatusTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AvailabilityZone); ok {
		return output, err
	}

	return nil, err
}

func waitAvailabilityZoneGroupNotOptedIn(ctx context.Context, conn *ec2.Client, name string) (*awstypes.AvailabilityZone, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AvailabilityZoneOptInStatusOptedIn),
		Target:  enum.Slice(awstypes.AvailabilityZoneOptInStatusNotOptedIn),
		Refresh: statusAvailabilityZoneGroupOptInStatus(ctx, conn, name),
		Timeout: AvailabilityZoneGroupOptInStatusTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AvailabilityZone); ok {
		return output, err
	}

	return nil, err
}

const (
	CapacityReservationActiveTimeout  = 2 * time.Minute
	CapacityReservationDeletedTimeout = 2 * time.Minute
)

func waitCapacityReservationActive(ctx context.Context, conn *ec2.Client, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapacityReservationStatePending),
		Target:  enum.Slice(awstypes.CapacityReservationStateActive),
		Refresh: statusCapacityReservation(ctx, conn, id),
		Timeout: CapacityReservationActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitCapacityReservationDeleted(ctx context.Context, conn *ec2.Client, id string) (*awstypes.CapacityReservation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapacityReservationStateActive),
		Target:  []string{},
		Refresh: statusCapacityReservation(ctx, conn, id),
		Timeout: CapacityReservationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CapacityReservation); ok {
		return output, err
	}

	return nil, err
}

func waitFleet(ctx context.Context, conn *ec2.Client, id string, pending, target []string, timeout, delay time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    statusFleet(ctx, conn, id),
		Timeout:    timeout,
		Delay:      delay,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitHostCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Host, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AllocationStatePending),
		Target:  enum.Slice(awstypes.AllocationStateAvailable),
		Timeout: timeout,
		Refresh: statusHost(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Host); ok {
		return output, err
	}

	return nil, err
}

func waitHostUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Host, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AllocationStatePending),
		Target:  enum.Slice(awstypes.AllocationStateAvailable),
		Timeout: timeout,
		Refresh: statusHost(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Host); ok {
		return output, err
	}

	return nil, err
}

func waitHostDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Host, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AllocationStateAvailable),
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusHost(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Host); ok {
		return output, err
	}

	return nil, err
}

func waitInstanceIAMInstanceProfileUpdated(ctx context.Context, conn *ec2.Client, instanceID string, expectedValue string) (*awstypes.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Target:     enum.Slice(expectedValue),
		Refresh:    statusInstanceIAMInstanceProfile(ctx, conn, instanceID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Instance); ok {
		return output, err
	}

	return nil, err
}

func waitInstanceCapacityReservationSpecificationUpdated(ctx context.Context, conn *ec2.Client, instanceID string, expectedValue *awstypes.CapacityReservationSpecification) (*awstypes.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Target:     enum.Slice(strconv.FormatBool(true)),
		Refresh:    statusInstanceCapacityReservationSpecificationEquals(ctx, conn, instanceID, expectedValue),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Instance); ok {
		return output, err
	}

	return nil, err
}

func waitInstanceMaintenanceOptionsAutoRecoveryUpdated(ctx context.Context, conn *ec2.Client, id, expectedValue string, timeout time.Duration) (*awstypes.InstanceMaintenanceOptions, error) {
	stateConf := &retry.StateChangeConf{
		Target:     enum.Slice(expectedValue),
		Refresh:    statusInstanceMaintenanceOptionsAutoRecovery(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.InstanceMaintenanceOptions); ok {
		return output, err
	}

	return nil, err
}

func waitInstanceMetadataOptionsApplied(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.InstanceMetadataOptionsResponse, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.InstanceMetadataOptionsStatePending),
		Target:     enum.Slice(awstypes.InstanceMetadataOptionsStateApplied),
		Refresh:    statusInstanceMetadataOptions(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.InstanceMetadataOptionsResponse); ok {
		return output, err
	}

	return nil, err
}

func waitInstanceRootBlockDeviceDeleteOnTerminationUpdated(ctx context.Context, conn *ec2.Client, id string, expectedValue bool, timeout time.Duration) (*awstypes.EbsInstanceBlockDevice, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    statusInstanceRootBlockDeviceDeleteOnTermination(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.EbsInstanceBlockDevice); ok {
		return output, err
	}

	return nil, err
}

const (
	PlacementGroupCreatedTimeout = 5 * time.Minute
	PlacementGroupDeletedTimeout = 5 * time.Minute
)

func waitPlacementGroupCreated(ctx context.Context, conn *ec2.Client, name string) (*awstypes.PlacementGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PlacementGroupStatePending),
		Target:  enum.Slice(awstypes.PlacementGroupStateAvailable),
		Timeout: PlacementGroupCreatedTimeout,
		Refresh: statusPlacementGroup(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PlacementGroup); ok {
		return output, err
	}

	return nil, err
}

func waitPlacementGroupDeleted(ctx context.Context, conn *ec2.Client, name string) (*awstypes.PlacementGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PlacementGroupStateDeleting),
		Target:  []string{},
		Timeout: PlacementGroupDeletedTimeout,
		Refresh: statusPlacementGroup(ctx, conn, name),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PlacementGroup); ok {
		return output, err
	}

	return nil, err
}

func waitSpotInstanceRequestFulfilled(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.SpotInstanceRequest, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{spotInstanceRequestStatusCodePendingEvaluation, spotInstanceRequestStatusCodePendingFulfillment},
		Target:     []string{spotInstanceRequestStatusCodeFulfilled},
		Refresh:    statusSpotInstanceRequest(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SpotInstanceRequest); ok {
		if fault := output.Fault; fault != nil {
			errFault := fmt.Errorf("%s: %s", aws.ToString(fault.Code), aws.ToString(fault.Message))
			tfresource.SetLastError(err, fmt.Errorf("%s %w", aws.ToString(output.Status.Message), errFault))
		} else {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCCreated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Vpc, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VpcStatePending),
		Target:  enum.Slice(awstypes.VpcStateAvailable),
		Refresh: statusVPC(ctx, conn, id),
		Timeout: vpcCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Vpc); ok {
		return output, err
	}

	return nil, err
}

func waitVPCIPv6CIDRBlockAssociationCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcCidrBlockState, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VpcCidrBlockStateCodeAssociating, awstypes.VpcCidrBlockStateCodeDisassociated, awstypes.VpcCidrBlockStateCodeFailing),
		Target:     enum.Slice(awstypes.VpcCidrBlockStateCodeAssociated),
		Refresh:    statusVPCIPv6CIDRBlockAssociation(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcCidrBlockState); ok {
		if state := output.State; state == awstypes.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCIPv6CIDRBlockAssociationDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VpcCidrBlockStateCodeAssociated, awstypes.VpcCidrBlockStateCodeDisassociating, awstypes.VpcCidrBlockStateCodeFailing),
		Target:     []string{},
		Refresh:    statusVPCIPv6CIDRBlockAssociation(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcCidrBlockState); ok {
		if state := output.State; state == awstypes.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCAttributeUpdated(ctx context.Context, conn *ec2.Client, vpcID string, attribute awstypes.VpcAttributeName, expectedValue bool) (*awstypes.Vpc, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    statusVPCAttributeValue(ctx, conn, vpcID, attribute),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Vpc); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceAvailableAfterUse(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.NetworkInterface, error) {
	// Hyperplane attached ENI.
	// Wait for it to be moved into a removable state.
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.NetworkInterfaceStatusInUse),
		Target:     enum.Slice(awstypes.NetworkInterfaceStatusAvailable),
		Timeout:    timeout,
		Refresh:    statusNetworkInterface(ctx, conn, id),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
		// Handle EC2 ENI eventual consistency. It can take up to 3 minutes.
		ContinuousTargetOccurence: 18,
		NotFoundChecks:            1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.NetworkInterface, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{NetworkInterfaceStatusPending},
		Target:  enum.Slice(awstypes.NetworkInterfaceStatusAvailable),
		Timeout: timeout,
		Refresh: statusNetworkInterface(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceAttached(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.NetworkInterfaceAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStatusAttaching),
		Target:  enum.Slice(awstypes.AttachmentStatusAttached),
		Timeout: timeout,
		Refresh: statusNetworkInterfaceAttachment(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceDetached(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.NetworkInterfaceAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStatusAttached, awstypes.AttachmentStatusDetaching),
		Target:  enum.Slice(awstypes.AttachmentStatusDetached),
		Timeout: timeout,
		Refresh: statusNetworkInterfaceAttachment(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVolumeCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Volume, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VolumeStateCreating),
		Target:     enum.Slice(awstypes.VolumeStateAvailable),
		Refresh:    statusVolume(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Volume); ok {
		return output, err
	}

	return nil, err
}

func waitVolumeDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Volume, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VolumeStateDeleting),
		Target:     []string{},
		Refresh:    statusVolume(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Volume); ok {
		return output, err
	}

	return nil, err
}

func waitVolumeUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Volume, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VolumeStateCreating, awstypes.VolumeState(awstypes.VolumeModificationStateModifying)),
		Target:     enum.Slice(awstypes.VolumeStateAvailable, awstypes.VolumeStateInUse),
		Refresh:    statusVolume(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Volume); ok {
		return output, err
	}

	return nil, err
}

func waitVolumeAttachmentCreated(ctx context.Context, conn *ec2.Client, volumeID, instanceID, deviceName string, timeout time.Duration) (*awstypes.VolumeAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VolumeAttachmentStateAttaching),
		Target:     enum.Slice(awstypes.VolumeAttachmentStateAttached),
		Refresh:    statusVolumeAttachment(ctx, conn, volumeID, instanceID, deviceName),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VolumeAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVolumeModificationComplete(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VolumeModification, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VolumeModificationStateModifying),
		// The volume is useable once the state is "optimizing", but will not be at full performance.
		// Optimization can take hours. e.g. a full 1 TiB drive takes approximately 6 hours to optimize,
		// according to https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/monitoring-volume-modifications.html.
		Target:     enum.Slice(awstypes.VolumeModificationStateCompleted, awstypes.VolumeModificationStateOptimizing),
		Refresh:    statusVolumeModification(ctx, conn, id),
		Timeout:    timeout,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VolumeModification); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitVPCEndpointAccepted(ctx context.Context, conn *ec2.Client, vpcEndpointID string, timeout time.Duration) (*awstypes.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpcEndpointStatePendingAcceptance),
		Target:     enum.Slice(vpcEndpointStateAvailable),
		Timeout:    timeout,
		Refresh:    statusVPCEndpoint(ctx, conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcEndpoint); ok {
		if state, lastError := output.State, output.LastError; state == awstypes.StateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(lastError.Code), aws.ToString(lastError.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCEndpointAvailable(ctx context.Context, conn *ec2.Client, vpcEndpointID string, timeout time.Duration) (*awstypes.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpcEndpointStatePending),
		Target:     enum.Slice(vpcEndpointStateAvailable, vpcEndpointStatePendingAcceptance),
		Timeout:    timeout,
		Refresh:    statusVPCEndpoint(ctx, conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcEndpoint); ok {
		if state, lastError := output.State, output.LastError; state == awstypes.StateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(lastError.Code), aws.ToString(lastError.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCEndpointDeleted(ctx context.Context, conn *ec2.Client, vpcEndpointID string, timeout time.Duration) (*awstypes.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpcEndpointStateDeleting, vpcEndpointStateDeleted),
		Target:     []string{},
		Refresh:    statusVPCEndpoint(ctx, conn, vpcEndpointID),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcEndpoint); ok {
		return output, err
	}

	return nil, err
}

func waitRouteDeleted(ctx context.Context, conn *ec2.Client, routeFinder routeFinder, routeTableID, destination string, timeout time.Duration) (*awstypes.Route, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{routeStatusReady},
		Target:                    []string{},
		Refresh:                   statusRoute(ctx, conn, routeFinder, routeTableID, destination),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Route); ok {
		return output, err
	}

	return nil, err
}

func waitRouteReady(ctx context.Context, conn *ec2.Client, routeFinder routeFinder, routeTableID, destination string, timeout time.Duration) (*awstypes.Route, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{routeStatusReady},
		Refresh:                   statusRoute(ctx, conn, routeFinder, routeTableID, destination),
		Timeout:                   timeout,
		NotFoundChecks:            RouteNotFoundChecks,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Route); ok {
		return output, err
	}

	return nil, err
}

func waitRouteTableReady(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{routeTableStatusReady},
		Refresh:                   statusRouteTable(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            RouteTableNotFoundChecks,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.RouteTable); ok {
		return output, err
	}

	return nil, err
}

func waitRouteTableDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{routeTableStatusReady},
		Target:                    []string{},
		Refresh:                   statusRouteTable(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.RouteTable); ok {
		return output, err
	}

	return nil, err
}

func waitRouteTableAssociationCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteTableAssociationState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.RouteTableAssociationStateCodeAssociating),
		Target:         enum.Slice(awstypes.RouteTableAssociationStateCodeAssociated),
		Refresh:        statusRouteTableAssociation(ctx, conn, id),
		Timeout:        timeout,
		NotFoundChecks: RouteTableAssociationCreatedNotFoundChecks,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.RouteTableAssociationState); ok {
		if output.State == awstypes.RouteTableAssociationStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitRouteTableAssociationDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteTableAssociationState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RouteTableAssociationStateCodeDisassociating, awstypes.RouteTableAssociationStateCodeAssociated),
		Target:  []string{},
		Refresh: statusRouteTableAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.RouteTableAssociationState); ok {
		if output.State == awstypes.RouteTableAssociationStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitRouteTableAssociationUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteTableAssociationState, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RouteTableAssociationStateCodeAssociating),
		Target:  enum.Slice(awstypes.RouteTableAssociationStateCodeAssociated),
		Refresh: statusRouteTableAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.RouteTableAssociationState); ok {
		if output.State == awstypes.RouteTableAssociationStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitSpotFleetRequestCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.SpotFleetRequestConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.BatchStateSubmitted),
		Target:     enum.Slice(awstypes.BatchStateActive),
		Refresh:    statusSpotFleetRequest(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SpotFleetRequestConfig); ok {
		return output, err
	}

	return nil, err
}

func waitSpotFleetRequestFulfilled(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.SpotFleetRequestConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ActivityStatusPendingFulfillment),
		Target:     enum.Slice(awstypes.ActivityStatusFulfilled),
		Refresh:    statusSpotFleetActivityStatus(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SpotFleetRequestConfig); ok {
		if output.ActivityStatus == awstypes.ActivityStatusError {
			var errs []error

			input := &ec2.DescribeSpotFleetRequestHistoryInput{
				SpotFleetRequestId: aws.String(id),
				StartTime:          aws.Time(time.UnixMilli(0)),
			}

			if output, err := findSpotFleetRequestHistoryRecords(ctx, conn, input); err == nil {
				for _, v := range output {
					if eventType := v.EventType; eventType == awstypes.EventTypeError || eventType == awstypes.EventTypeInformation {
						errs = append(errs, errors.New(aws.ToString(v.EventInformation.EventDescription)))
					}
				}
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return output, err
	}

	return nil, err
}

func waitSpotFleetRequestUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.SpotFleetRequestConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.BatchStateModifying),
		Target:     enum.Slice(awstypes.BatchStateActive),
		Refresh:    statusSpotFleetRequest(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SpotFleetRequestConfig); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointServiceAvailable(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.ServiceConfiguration, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ServiceStatePending),
		Target:     enum.Slice(awstypes.ServiceStateAvailable),
		Refresh:    statusVPCEndpointServiceAvailable(ctx, conn, id),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ServiceConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointServiceDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.ServiceConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ServiceStateAvailable, awstypes.ServiceStateDeleting),
		Target:     []string{},
		Timeout:    timeout,
		Refresh:    fetchVPCEndpointServiceDeletionStatus(ctx, conn, id),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ServiceConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointRouteTableAssociationReady(ctx context.Context, conn *ec2.Client, vpcEndpointID, routeTableID string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(VPCEndpointRouteTableAssociationStatusReady),
		Refresh:                   statusVPCEndpointRouteTableAssociation(ctx, conn, vpcEndpointID, routeTableID),
		Timeout:                   ec2PropagationTimeout,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitVPCEndpointRouteTableAssociationDeleted(ctx context.Context, conn *ec2.Client, vpcEndpointID, routeTableID string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(VPCEndpointRouteTableAssociationStatusReady),
		Target:                    []string{},
		Refresh:                   statusVPCEndpointRouteTableAssociation(ctx, conn, vpcEndpointID, routeTableID),
		Timeout:                   ec2PropagationTimeout,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitVPCEndpointConnectionAccepted(ctx context.Context, conn *ec2.Client, serviceID, vpcEndpointID string, timeout time.Duration) (*awstypes.VpcEndpointConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{vpcEndpointStatePendingAcceptance, vpcEndpointStatePending},
		Target:     []string{vpcEndpointStateAvailable},
		Refresh:    statusVPCEndpointConnectionVPCEndpoint(ctx, conn, serviceID, vpcEndpointID),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcEndpointConnection); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointServicePrivateDNSNameVerified(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.PrivateDnsNameConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.DnsNameStatePendingVerification),
		Target:                    enum.Slice(awstypes.DnsNameStateVerified),
		Refresh:                   statusVPCEndpointServicePrivateDNSNameConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.PrivateDnsNameConfiguration); ok {
		return out, err
	}

	return nil, err
}

func waitClientVPNEndpointDeleted(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ClientVpnEndpoint, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ClientVpnEndpointStatusCodeDeleting),
		Target:  []string{},
		Refresh: statusClientVPNEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ClientVpnEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitClientVPNEndpointClientConnectResponseOptionsUpdated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.ClientConnectResponseOptions, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ClientVpnEndpointAttributeStatusCodeApplying),
		Target:  enum.Slice(awstypes.ClientVpnEndpointAttributeStatusCodeApplied),
		Refresh: statusClientVPNEndpointClientConnectResponseOptions(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ClientConnectResponseOptions); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitClientVPNAuthorizationRuleCreated(ctx context.Context, conn *ec2.Client, endpointID, targetNetworkCIDR, accessGroupID string, timeout time.Duration) (*awstypes.AuthorizationRule, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ClientVpnAuthorizationRuleStatusCodeAuthorizing),
		Target:  enum.Slice(awstypes.ClientVpnAuthorizationRuleStatusCodeActive),
		Refresh: statusClientVPNAuthorizationRule(ctx, conn, endpointID, targetNetworkCIDR, accessGroupID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AuthorizationRule); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitClientVPNAuthorizationRuleDeleted(ctx context.Context, conn *ec2.Client, endpointID, targetNetworkCIDR, accessGroupID string, timeout time.Duration) (*awstypes.AuthorizationRule, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ClientVpnAuthorizationRuleStatusCodeRevoking),
		Target:  []string{},
		Refresh: statusClientVPNAuthorizationRule(ctx, conn, endpointID, targetNetworkCIDR, accessGroupID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AuthorizationRule); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitClientVPNNetworkAssociationCreated(ctx context.Context, conn *ec2.Client, associationID, endpointID string, timeout time.Duration) (*awstypes.TargetNetwork, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(awstypes.AssociationStatusCodeAssociating),
		Target:       enum.Slice(awstypes.AssociationStatusCodeAssociated),
		Refresh:      statusClientVPNNetworkAssociation(ctx, conn, associationID, endpointID),
		Timeout:      timeout,
		Delay:        4 * time.Minute,
		PollInterval: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TargetNetwork); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitClientVPNNetworkAssociationDeleted(ctx context.Context, conn *ec2.Client, associationID, endpointID string, timeout time.Duration) (*awstypes.TargetNetwork, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(awstypes.AssociationStatusCodeDisassociating),
		Target:       []string{},
		Refresh:      statusClientVPNNetworkAssociation(ctx, conn, associationID, endpointID),
		Timeout:      timeout,
		Delay:        4 * time.Minute,
		PollInterval: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TargetNetwork); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitClientVPNRouteCreated(ctx context.Context, conn *ec2.Client, endpointID, targetSubnetID, destinationCIDR string, timeout time.Duration) (*awstypes.ClientVpnRoute, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ClientVpnRouteStatusCodeCreating),
		Target:  enum.Slice(awstypes.ClientVpnRouteStatusCodeActive),
		Refresh: statusClientVPNRoute(ctx, conn, endpointID, targetSubnetID, destinationCIDR),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ClientVpnRoute); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitClientVPNRouteDeleted(ctx context.Context, conn *ec2.Client, endpointID, targetSubnetID, destinationCIDR string, timeout time.Duration) (*awstypes.ClientVpnRoute, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ClientVpnRouteStatusCodeActive, awstypes.ClientVpnRouteStatusCodeDeleting),
		Target:  []string{},
		Refresh: statusClientVPNRoute(ctx, conn, endpointID, targetSubnetID, destinationCIDR),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ClientVpnRoute); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitCarrierGatewayCreated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.CarrierGateway, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CarrierGatewayStatePending),
		Target:  enum.Slice(awstypes.CarrierGatewayStateAvailable),
		Refresh: statusCarrierGateway(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CarrierGateway); ok {
		return output, err
	}

	return nil, err
}

func waitCarrierGatewayDeleted(ctx context.Context, conn *ec2.Client, id string) (*awstypes.CarrierGateway, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CarrierGatewayStateDeleting),
		Target:  []string{},
		Refresh: statusCarrierGateway(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CarrierGateway); ok {
		return output, err
	}

	return nil, err
}

func waitImageAvailable(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Image, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ImageStatePending),
		Target:     enum.Slice(awstypes.ImageStateAvailable),
		Refresh:    statusImage(ctx, conn, id),
		Timeout:    timeout,
		Delay:      amiRetryDelay,
		MinTimeout: amiRetryMinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Image); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitImageDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Image, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ImageStateAvailable, awstypes.ImageStateFailed, awstypes.ImageStatePending),
		Target:     []string{},
		Refresh:    statusImage(ctx, conn, id),
		Timeout:    timeout,
		Delay:      amiRetryDelay,
		MinTimeout: amiRetryMinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Image); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitImageBlockPublicAccessState(ctx context.Context, conn *ec2.Client, target string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Target:  []string{target},
		Refresh: statusImageBlockPublicAccess(ctx, conn),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitVPNConnectionCreated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpnConnection, error) {
	const (
		timeout = 40 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VpnStatePending),
		Target:     enum.Slice(awstypes.VpnStateAvailable),
		Refresh:    statusVPNConnection(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpnConnection); ok {
		return output, err
	}

	return nil, err
}

func waitVPNConnectionUpdated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpnConnection, error) { //nolint:unparam
	const (
		timeout = 30 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpnStateModifying),
		Target:     enum.Slice(awstypes.VpnStateAvailable),
		Refresh:    statusVPNConnection(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpnConnection); ok {
		return output, err
	}

	return nil, err
}

func waitVPNConnectionDeleted(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpnConnection, error) {
	const (
		timeout = 30 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VpnStateDeleting),
		Target:     []string{},
		Refresh:    statusVPNConnection(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpnConnection); ok {
		return output, err
	}

	return nil, err
}

func waitVPNConnectionRouteCreated(ctx context.Context, conn *ec2.Client, vpnConnectionID, cidrBlock string) (*awstypes.VpnStaticRoute, error) {
	const (
		timeout = 15 * time.Second
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VpnStatePending),
		Target:  enum.Slice(awstypes.VpnStateAvailable),
		Refresh: statusVPNConnectionRoute(ctx, conn, vpnConnectionID, cidrBlock),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpnStaticRoute); ok {
		return output, err
	}

	return nil, err
}

func waitVPNConnectionRouteDeleted(ctx context.Context, conn *ec2.Client, vpnConnectionID, cidrBlock string) (*awstypes.VpnStaticRoute, error) {
	const (
		timeout = 15 * time.Second
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VpnStatePending, awstypes.VpnStateAvailable, awstypes.VpnStateDeleting),
		Target:  []string{},
		Refresh: statusVPNConnectionRoute(ctx, conn, vpnConnectionID, cidrBlock),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpnStaticRoute); ok {
		return output, err
	}

	return nil, err
}

func waitVPNGatewayCreated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpnGateway, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VpnStatePending),
		Target:     enum.Slice(awstypes.VpnStateAvailable),
		Refresh:    statusVPNGateway(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpnGateway); ok {
		return output, err
	}

	return nil, err
}

func waitVPNGatewayDeleted(ctx context.Context, conn *ec2.Client, id string) (*awstypes.VpnGateway, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VpnStateDeleting),
		Target:     []string{},
		Refresh:    statusVPNGateway(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpnGateway); ok {
		return output, err
	}

	return nil, err
}

func waitVPNGatewayVPCAttachmentAttached(ctx context.Context, conn *ec2.Client, vpnGatewayID, vpcID string) (*awstypes.VpcAttachment, error) { //nolint:unparam
	const (
		timeout = 15 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStatusAttaching),
		Target:  enum.Slice(awstypes.AttachmentStatusAttached),
		Refresh: statusVPNGatewayVPCAttachment(ctx, conn, vpnGatewayID, vpcID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVPNGatewayVPCAttachmentDetached(ctx context.Context, conn *ec2.Client, vpnGatewayID, vpcID string) (*awstypes.VpcAttachment, error) { //nolint:unparam
	const (
		timeout = 30 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStatusAttached, awstypes.AttachmentStatusDetaching),
		Target:  []string{},
		Refresh: statusVPNGatewayVPCAttachment(ctx, conn, vpnGatewayID, vpcID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitCustomerGatewayCreated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.CustomerGateway, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(CustomerGatewayStatePending),
		Target:     enum.Slice(CustomerGatewayStateAvailable),
		Refresh:    statusCustomerGateway(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CustomerGateway); ok {
		return output, err
	}

	return nil, err
}

func waitCustomerGatewayDeleted(ctx context.Context, conn *ec2.Client, id string) (*awstypes.CustomerGateway, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(CustomerGatewayStateAvailable, CustomerGatewayStateDeleting),
		Target:  []string{},
		Refresh: statusCustomerGateway(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.CustomerGateway); ok {
		return output, err
	}

	return nil, err
}

func waitIPAMCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Ipam, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamStateCreateInProgress),
		Target:  enum.Slice(awstypes.IpamStateCreateComplete),
		Refresh: statusIPAM(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Ipam); ok {
		return output, err
	}

	return nil, err
}

func waitIPAMUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Ipam, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamStateModifyInProgress),
		Target:  enum.Slice(awstypes.IpamStateModifyComplete),
		Refresh: statusIPAM(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Ipam); ok {
		return output, err
	}

	return nil, err
}

func waitIPAMDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Ipam, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamStateCreateComplete, awstypes.IpamStateModifyComplete, awstypes.IpamStateDeleteInProgress),
		Target:  []string{},
		Refresh: statusIPAM(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Ipam); ok {
		return output, err
	}

	return nil, err
}

func waitIPAMPoolCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.IpamPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamPoolStateCreateInProgress),
		Target:  enum.Slice(awstypes.IpamPoolStateCreateComplete),
		Refresh: statusIPAMPool(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamPool); ok {
		if state := output.State; state == awstypes.IpamPoolStateCreateFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitIPAMPoolUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.IpamPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamPoolStateModifyInProgress),
		Target:  enum.Slice(awstypes.IpamPoolStateModifyComplete),
		Refresh: statusIPAMPool(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamPool); ok {
		if state := output.State; state == awstypes.IpamPoolStateModifyFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitIPAMPoolDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.IpamPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamPoolStateDeleteInProgress),
		Target:  []string{},
		Refresh: statusIPAMPool(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamPool); ok {
		if state := output.State; state == awstypes.IpamPoolStateDeleteFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitIPAMPoolCIDRCreated(ctx context.Context, conn *ec2.Client, poolCIDRID, poolID, cidrBlock string, timeout time.Duration) (*awstypes.IpamPoolCidr, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.IpamPoolCidrStatePendingProvision),
		Target:         enum.Slice(awstypes.IpamPoolCidrStateProvisioned),
		Refresh:        statusIPAMPoolCIDR(ctx, conn, cidrBlock, poolID, poolCIDRID),
		Timeout:        timeout,
		Delay:          5 * time.Second,
		NotFoundChecks: 1000, // Should exceed any reasonable custom timeout value.
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamPoolCidr); ok {
		if state, failureReason := output.State, output.FailureReason; state == awstypes.IpamPoolCidrStateFailedProvision && failureReason != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", string(failureReason.Code), aws.ToString(failureReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitIPAMPoolCIDRDeleted(ctx context.Context, conn *ec2.Client, poolCIDRID, poolID, cidrBlock string, timeout time.Duration) (*awstypes.IpamPoolCidr, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamPoolCidrStatePendingDeprovision, awstypes.IpamPoolCidrStateProvisioned),
		Target:  []string{},
		Refresh: statusIPAMPoolCIDR(ctx, conn, cidrBlock, poolID, poolCIDRID),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamPoolCidr); ok {
		if state, failureReason := output.State, output.FailureReason; state == awstypes.IpamPoolCidrStateFailedDeprovision && failureReason != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", string(failureReason.Code), aws.ToString(failureReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitIPAMResourceDiscoveryCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.IpamResourceDiscovery, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamResourceDiscoveryStateCreateInProgress),
		Target:  enum.Slice(awstypes.IpamResourceDiscoveryStateCreateComplete),
		Refresh: statusIPAMResourceDiscovery(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamResourceDiscovery); ok {
		return output, err
	}

	return nil, err
}

func waitIPAMResourceDiscoveryUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.IpamResourceDiscovery, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamResourceDiscoveryStateModifyInProgress),
		Target:  enum.Slice(awstypes.IpamResourceDiscoveryStateModifyComplete),
		Refresh: statusIPAMResourceDiscovery(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamResourceDiscovery); ok {
		return output, err
	}

	return nil, err
}

func waitIPAMResourceDiscoveryDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.IpamResourceDiscovery, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamResourceDiscoveryStateCreateComplete, awstypes.IpamResourceDiscoveryStateModifyComplete, awstypes.IpamResourceDiscoveryStateDeleteInProgress),
		Target:  []string{},
		Refresh: statusIPAMResourceDiscovery(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamResourceDiscovery); ok {
		return output, err
	}

	return nil, err
}

func waitIPAMResourceDiscoveryAssociationCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.IpamResourceDiscoveryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamResourceDiscoveryAssociationStateAssociateInProgress),
		Target:  enum.Slice(awstypes.IpamResourceDiscoveryAssociationStateAssociateComplete),
		Refresh: statusIPAMResourceDiscoveryAssociation(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamResourceDiscoveryAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitIPAMResourceDiscoveryAssociationDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.IpamResourceDiscoveryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamResourceDiscoveryAssociationStateAssociateComplete, awstypes.IpamResourceDiscoveryAssociationStateDisassociateInProgress),
		Target:  []string{},
		Refresh: statusIPAMResourceDiscoveryAssociation(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamResourceDiscoveryAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitIPAMScopeCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.IpamScope, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamScopeStateCreateInProgress),
		Target:  enum.Slice(awstypes.IpamScopeStateCreateComplete),
		Refresh: statusIPAMScope(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamScope); ok {
		return output, err
	}

	return nil, err
}

func waitIPAMScopeUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.IpamScope, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamScopeStateModifyInProgress),
		Target:  enum.Slice(awstypes.IpamScopeStateModifyComplete),
		Refresh: statusIPAMScope(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamScope); ok {
		return output, err
	}

	return nil, err
}

func waitIPAMScopeDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.IpamScope, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IpamScopeStateCreateComplete, awstypes.IpamScopeStateModifyComplete, awstypes.IpamScopeStateDeleteInProgress),
		Target:  []string{},
		Refresh: statusIPAMScope(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.IpamScope); ok {
		return output, err
	}

	return nil, err
}

func waitLocalGatewayRouteDeleted(ctx context.Context, conn *ec2.Client, localGatewayRouteTableID, destinationCIDRBlock string) (*awstypes.LocalGatewayRoute, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LocalGatewayRouteStateDeleting),
		Target:  []string{},
		Refresh: statusLocalGatewayRoute(ctx, conn, localGatewayRouteTableID, destinationCIDRBlock),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.LocalGatewayRoute); ok {
		return output, err
	}

	return nil, err
}

func waitLocalGatewayRouteTableVPCAssociationAssociated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.LocalGatewayRouteTableVpcAssociation, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RouteTableAssociationStateCodeAssociating),
		Target:  enum.Slice(awstypes.RouteTableAssociationStateCodeAssociated),
		Refresh: statusLocalGatewayRouteTableVPCAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.LocalGatewayRouteTableVpcAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitLocalGatewayRouteTableVPCAssociationDisassociated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.LocalGatewayRouteTableVpcAssociation, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RouteTableAssociationStateCodeDisassociating),
		Target:  []string{},
		Refresh: statusLocalGatewayRouteTableVPCAssociation(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.LocalGatewayRouteTableVpcAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.TransitGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayStatePending),
		Target:  enum.Slice(awstypes.TransitGatewayStateAvailable),
		Refresh: statusTransitGateway(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGateway); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.TransitGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.TransitGatewayStateAvailable, awstypes.TransitGatewayStateDeleting),
		Target:         []string{},
		Refresh:        statusTransitGateway(ctx, conn, id),
		Timeout:        timeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGateway); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.TransitGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayStateModifying),
		Target:  enum.Slice(awstypes.TransitGatewayStateAvailable),
		Refresh: statusTransitGateway(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGateway); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayConnectCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayConnect, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayAttachmentStatePending),
		Target:  enum.Slice(awstypes.TransitGatewayAttachmentStateAvailable),
		Refresh: statusTransitGatewayConnect(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayConnect); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayConnectDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayConnect, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.TransitGatewayAttachmentStateAvailable, awstypes.TransitGatewayAttachmentStateDeleting),
		Target:         []string{},
		Refresh:        statusTransitGatewayConnect(ctx, conn, id),
		Timeout:        timeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayConnect); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayConnectPeerCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayConnectPeer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayConnectPeerStatePending),
		Target:  enum.Slice(awstypes.TransitGatewayConnectPeerStateAvailable),
		Refresh: statusTransitGatewayConnectPeer(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayConnectPeer); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayConnectPeerDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayConnectPeer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayConnectPeerStateAvailable, awstypes.TransitGatewayConnectPeerStateDeleting),
		Target:  []string{},
		Refresh: statusTransitGatewayConnectPeer(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayConnectPeer); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayMulticastDomainCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayMulticastDomain, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayMulticastDomainStatePending),
		Target:  enum.Slice(awstypes.TransitGatewayMulticastDomainStateAvailable),
		Refresh: statusTransitGatewayMulticastDomain(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayMulticastDomain); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayMulticastDomainDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.TransitGatewayMulticastDomain, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayMulticastDomainStateAvailable, awstypes.TransitGatewayMulticastDomainStateDeleting),
		Target:  []string{},
		Refresh: statusTransitGatewayMulticastDomain(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayMulticastDomain); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayMulticastDomainAssociationCreated(ctx context.Context, conn *ec2.Client, multicastDomainID, attachmentID, subnetID string, timeout time.Duration) (*awstypes.TransitGatewayMulticastDomainAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AssociationStatusCodeAssociating),
		Target:  enum.Slice(awstypes.AssociationStatusCodeAssociated),
		Refresh: statusTransitGatewayMulticastDomainAssociation(ctx, conn, multicastDomainID, attachmentID, subnetID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayMulticastDomainAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayMulticastDomainAssociationDeleted(ctx context.Context, conn *ec2.Client, multicastDomainID, attachmentID, subnetID string, timeout time.Duration) (*awstypes.TransitGatewayMulticastDomainAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AssociationStatusCodeAssociated, awstypes.AssociationStatusCodeDisassociating),
		Target:  []string{},
		Refresh: statusTransitGatewayMulticastDomainAssociation(ctx, conn, multicastDomainID, attachmentID, subnetID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayMulticastDomainAssociation); ok {
		return output, err
	}

	return nil, err
}

const (
	TransitGatewayPeeringAttachmentCreatedTimeout = 10 * time.Minute
	TransitGatewayPeeringAttachmentDeletedTimeout = 10 * time.Minute
	TransitGatewayPeeringAttachmentUpdatedTimeout = 10 * time.Minute
)

func waitTransitGatewayPeeringAttachmentAccepted(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayPeeringAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayAttachmentStatePending, awstypes.TransitGatewayAttachmentStatePendingAcceptance),
		Target:  enum.Slice(awstypes.TransitGatewayAttachmentStateAvailable),
		Timeout: TransitGatewayPeeringAttachmentUpdatedTimeout,
		Refresh: statusTransitGatewayPeeringAttachment(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayPeeringAttachment); ok {
		if status := output.Status; status != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(status.Code), aws.ToString(status.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitTransitGatewayPeeringAttachmentCreated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayPeeringAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayAttachmentStateFailing, awstypes.TransitGatewayAttachmentStateInitiatingRequest, awstypes.TransitGatewayAttachmentStatePending),
		Target:  enum.Slice(awstypes.TransitGatewayAttachmentStateAvailable, awstypes.TransitGatewayAttachmentStatePendingAcceptance),
		Timeout: TransitGatewayPeeringAttachmentCreatedTimeout,
		Refresh: statusTransitGatewayPeeringAttachment(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayPeeringAttachment); ok {
		if status := output.Status; status != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(status.Code), aws.ToString(status.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitTransitGatewayPeeringAttachmentDeleted(ctx context.Context, conn *ec2.Client, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.TransitGatewayAttachmentStateAvailable,
			awstypes.TransitGatewayAttachmentStateDeleting,
			awstypes.TransitGatewayAttachmentStatePendingAcceptance,
			awstypes.TransitGatewayAttachmentStateRejecting,
		),
		Target:  enum.Slice(awstypes.TransitGatewayAttachmentStateDeleted),
		Timeout: TransitGatewayPeeringAttachmentDeletedTimeout,
		Refresh: statusTransitGatewayPeeringAttachment(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayPeeringAttachment); ok {
		if status := output.Status; status != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(status.Code), aws.ToString(status.Message)))
		}
	}

	return err
}

const (
	TransitGatewayPrefixListReferenceTimeout = 5 * time.Minute
)

func waitTransitGatewayPrefixListReferenceStateCreated(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID string, prefixListID string) (*awstypes.TransitGatewayPrefixListReference, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayPrefixListReferenceStatePending),
		Target:  enum.Slice(awstypes.TransitGatewayPrefixListReferenceStateAvailable),
		Timeout: TransitGatewayPrefixListReferenceTimeout,
		Refresh: statusTransitGatewayPrefixListReference(ctx, conn, transitGatewayRouteTableID, prefixListID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayPrefixListReference); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayPrefixListReferenceStateDeleted(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID string, prefixListID string) (*awstypes.TransitGatewayPrefixListReference, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayPrefixListReferenceStateDeleting),
		Target:  []string{},
		Timeout: TransitGatewayPrefixListReferenceTimeout,
		Refresh: statusTransitGatewayPrefixListReference(ctx, conn, transitGatewayRouteTableID, prefixListID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayPrefixListReference); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayPrefixListReferenceStateUpdated(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID string, prefixListID string) (*awstypes.TransitGatewayPrefixListReference, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayPrefixListReferenceStateModifying),
		Target:  enum.Slice(awstypes.TransitGatewayPrefixListReferenceStateAvailable),
		Timeout: TransitGatewayPrefixListReferenceTimeout,
		Refresh: statusTransitGatewayPrefixListReference(ctx, conn, transitGatewayRouteTableID, prefixListID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayPrefixListReference); ok {
		return output, err
	}

	return nil, err
}

const (
	TransitGatewayRouteCreatedTimeout = 2 * time.Minute
	TransitGatewayRouteDeletedTimeout = 2 * time.Minute
)

func waitTransitGatewayRouteCreated(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, destination string) (*awstypes.TransitGatewayRoute, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayRouteStatePending),
		Target:  enum.Slice(awstypes.TransitGatewayRouteStateActive, awstypes.TransitGatewayRouteStateBlackhole),
		Timeout: TransitGatewayRouteCreatedTimeout,
		Refresh: statusTransitGatewayStaticRoute(ctx, conn, transitGatewayRouteTableID, destination),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayRoute); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayRouteDeleted(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, destination string) (*awstypes.TransitGatewayRoute, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayRouteStateActive, awstypes.TransitGatewayRouteStateBlackhole, awstypes.TransitGatewayRouteStateDeleting),
		Target:  []string{},
		Timeout: TransitGatewayRouteDeletedTimeout,
		Refresh: statusTransitGatewayStaticRoute(ctx, conn, transitGatewayRouteTableID, destination),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayRoute); ok {
		return output, err
	}

	return nil, err
}

const (
	TransitGatewayRouteTableCreatedTimeout  = 10 * time.Minute
	TransitGatewayRouteTableDeletedTimeout  = 10 * time.Minute
	TransitGatewayPolicyTableCreatedTimeout = 10 * time.Minute
	TransitGatewayPolicyTableDeletedTimeout = 10 * time.Minute
)

func waitTransitGatewayPolicyTableCreated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayPolicyTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayPolicyTableStatePending),
		Target:  enum.Slice(awstypes.TransitGatewayPolicyTableStateAvailable),
		Timeout: TransitGatewayPolicyTableCreatedTimeout,
		Refresh: statusTransitGatewayPolicyTable(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayPolicyTable); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayRouteTableCreated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayRouteTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayRouteTableStatePending),
		Target:  enum.Slice(awstypes.TransitGatewayRouteTableStateAvailable),
		Timeout: TransitGatewayRouteTableCreatedTimeout,
		Refresh: statusTransitGatewayRouteTable(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayRouteTable); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayPolicyTableDeleted(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayPolicyTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayPolicyTableStateAvailable, awstypes.TransitGatewayPolicyTableStateDeleting),
		Target:  []string{},
		Timeout: TransitGatewayPolicyTableDeletedTimeout,
		Refresh: statusTransitGatewayPolicyTable(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayPolicyTable); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayRouteTableDeleted(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayRouteTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayRouteTableStateAvailable, awstypes.TransitGatewayRouteTableStateDeleting),
		Target:  []string{},
		Timeout: TransitGatewayRouteTableDeletedTimeout,
		Refresh: statusTransitGatewayRouteTable(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayRouteTable); ok {
		return output, err
	}

	return nil, err
}

const (
	TransitGatewayPolicyTableAssociationCreatedTimeout = 5 * time.Minute
	TransitGatewayPolicyTableAssociationDeletedTimeout = 10 * time.Minute
	TransitGatewayRouteTableAssociationCreatedTimeout  = 5 * time.Minute
	TransitGatewayRouteTableAssociationDeletedTimeout  = 10 * time.Minute
)

func waitTransitGatewayPolicyTableAssociationCreated(ctx context.Context, conn *ec2.Client, transitGatewayPolicyTableID, transitGatewayAttachmentID string) (*awstypes.TransitGatewayPolicyTableAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayAssociationStateAssociating),
		Target:  enum.Slice(awstypes.TransitGatewayAssociationStateAssociated),
		Timeout: TransitGatewayPolicyTableAssociationCreatedTimeout,
		Refresh: statusTransitGatewayPolicyTableAssociation(ctx, conn, transitGatewayPolicyTableID, transitGatewayAttachmentID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayPolicyTableAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayPolicyTableAssociationDeleted(ctx context.Context, conn *ec2.Client, transitGatewayPolicyTableID, transitGatewayAttachmentID string) (*awstypes.TransitGatewayPolicyTableAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.TransitGatewayAssociationStateAssociated, awstypes.TransitGatewayAssociationStateDisassociating),
		Target:         []string{},
		Timeout:        TransitGatewayPolicyTableAssociationDeletedTimeout,
		Refresh:        statusTransitGatewayPolicyTableAssociation(ctx, conn, transitGatewayPolicyTableID, transitGatewayAttachmentID),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayPolicyTableAssociation); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayRouteTableAssociationCreated(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, transitGatewayAttachmentID string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayAssociationStateAssociating),
		Target:  enum.Slice(awstypes.TransitGatewayAssociationStateAssociated),
		Timeout: TransitGatewayRouteTableAssociationCreatedTimeout,
		Refresh: statusTransitGatewayRouteTableAssociation(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitTransitGatewayRouteTableAssociationDeleted(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID, transitGatewayAttachmentID string) error {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.TransitGatewayAssociationStateAssociated, awstypes.TransitGatewayAssociationStateDisassociating),
		Target:         []string{},
		Timeout:        TransitGatewayRouteTableAssociationDeletedTimeout,
		Refresh:        statusTransitGatewayRouteTableAssociation(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID),
		NotFoundChecks: 1,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

const (
	TransitGatewayRouteTablePropagationCreatedTimeout = 5 * time.Minute
	TransitGatewayRouteTablePropagationDeletedTimeout = 5 * time.Minute
)

func waitTransitGatewayRouteTablePropagationCreated(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID string, transitGatewayAttachmentID string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayPropagationStateEnabling),
		Target:  enum.Slice(awstypes.TransitGatewayPropagationStateEnabled),
		Timeout: TransitGatewayRouteTablePropagationCreatedTimeout,
		Refresh: statusTransitGatewayRouteTablePropagation(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitTransitGatewayRouteTablePropagationDeleted(ctx context.Context, conn *ec2.Client, transitGatewayRouteTableID string, transitGatewayAttachmentID string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayPropagationStateDisabling),
		Target:  []string{},
		Timeout: TransitGatewayRouteTablePropagationDeletedTimeout,
		Refresh: statusTransitGatewayRouteTablePropagation(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil
	}

	return err
}

const (
	TransitGatewayVPCAttachmentCreatedTimeout = 10 * time.Minute
	TransitGatewayVPCAttachmentDeletedTimeout = 10 * time.Minute
	TransitGatewayVPCAttachmentUpdatedTimeout = 10 * time.Minute
)

func waitTransitGatewayVPCAttachmentAccepted(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayVpcAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayAttachmentStatePending, awstypes.TransitGatewayAttachmentStatePendingAcceptance),
		Target:  enum.Slice(awstypes.TransitGatewayAttachmentStateAvailable),
		Timeout: TransitGatewayVPCAttachmentUpdatedTimeout,
		Refresh: statusTransitGatewayVPCAttachment(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayVpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayVPCAttachmentCreated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayVpcAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayAttachmentStateFailing, awstypes.TransitGatewayAttachmentStatePending),
		Target:  enum.Slice(awstypes.TransitGatewayAttachmentStateAvailable, awstypes.TransitGatewayAttachmentStatePendingAcceptance),
		Timeout: TransitGatewayVPCAttachmentCreatedTimeout,
		Refresh: statusTransitGatewayVPCAttachment(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayVpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitTransitGatewayVPCAttachmentDeleted(ctx context.Context, conn *ec2.Client, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.TransitGatewayAttachmentStateAvailable,
			awstypes.TransitGatewayAttachmentStateDeleting,
			awstypes.TransitGatewayAttachmentStatePendingAcceptance,
			awstypes.TransitGatewayAttachmentStateRejecting,
		),
		Target:  enum.Slice(awstypes.TransitGatewayAttachmentStateDeleted),
		Timeout: TransitGatewayVPCAttachmentDeletedTimeout,
		Refresh: statusTransitGatewayVPCAttachment(ctx, conn, id),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitTransitGatewayVPCAttachmentUpdated(ctx context.Context, conn *ec2.Client, id string) (*awstypes.TransitGatewayVpcAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransitGatewayAttachmentStateModifying),
		Target:  enum.Slice(awstypes.TransitGatewayAttachmentStateAvailable),
		Timeout: TransitGatewayVPCAttachmentUpdatedTimeout,
		Refresh: statusTransitGatewayVPCAttachment(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TransitGatewayVpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitEIPDomainNameAttributeUpdated(ctx context.Context, conn *ec2.Client, allocationID string, timeout time.Duration) (*awstypes.AddressAttribute, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{PTRUpdateStatusPending},
		Target:  []string{""},
		Timeout: timeout,
		Refresh: statusEIPDomainNameAttribute(ctx, conn, allocationID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AddressAttribute); ok {
		if v := output.PtrRecordUpdate; v != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(v.Reason)))
		}

		return output, err
	}

	return nil, err
}

func waitEIPDomainNameAttributeDeleted(ctx context.Context, conn *ec2.Client, allocationID string, timeout time.Duration) (*awstypes.AddressAttribute, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{PTRUpdateStatusPending},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusEIPDomainNameAttribute(ctx, conn, allocationID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AddressAttribute); ok {
		if v := output.PtrRecordUpdate; v != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(v.Reason)))
		}

		return output, err
	}

	return nil, err
}

func waitEBSSnapshotImportComplete(ctx context.Context, conn *ec2.Client, importTaskID string, timeout time.Duration) (*awstypes.SnapshotTaskDetail, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			EBSSnapshotImportStateActive,
			EBSSnapshotImportStateUpdating,
			EBSSnapshotImportStateValidating,
			EBSSnapshotImportStateValidated,
			EBSSnapshotImportStateConverting,
		},
		Target:  []string{EBSSnapshotImportStateCompleted},
		Refresh: statusEBSSnapshotImport(ctx, conn, importTaskID),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SnapshotTaskDetail); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

const (
	ebsSnapshotArchivedTimeout = 60 * time.Minute
)

func waitEBSSnapshotTierArchive(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.SnapshotTierStatus, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(TargetStorageTierStandard),
		Target:  enum.Slice(awstypes.TargetStorageTierArchive),
		Refresh: statusSnapshotStorageTier(ctx, conn, id),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SnapshotTierStatus); ok {
		tfresource.SetLastError(err, fmt.Errorf("%s: %s", string(output.LastTieringOperationStatus), aws.ToString(output.LastTieringOperationStatusDetail)))

		return output, err
	}

	return nil, err
}

func waitInstanceConnectEndpointCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Ec2InstanceConnectEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.Ec2InstanceConnectEndpointStateCreateInProgress),
		Target:  enum.Slice(awstypes.Ec2InstanceConnectEndpointStateCreateComplete),
		Refresh: statusInstanceConnectEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Ec2InstanceConnectEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))

		return output, err
	}

	return nil, err
}

func waitInstanceConnectEndpointDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Ec2InstanceConnectEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.Ec2InstanceConnectEndpointStateDeleteInProgress),
		Target:  []string{},
		Refresh: statusInstanceConnectEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Ec2InstanceConnectEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))

		return output, err
	}

	return nil, err
}

func waitVerifiedAccessEndpointCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VerifiedAccessEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.VerifiedAccessEndpointStatusCodePending),
		Target:                    enum.Slice(awstypes.VerifiedAccessEndpointStatusCodeActive),
		Refresh:                   statusVerifiedAccessEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VerifiedAccessEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitVerifiedAccessEndpointUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VerifiedAccessEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.VerifiedAccessEndpointStatusCodeUpdating),
		Target:                    enum.Slice(awstypes.VerifiedAccessEndpointStatusCodeActive),
		Refresh:                   statusVerifiedAccessEndpoint(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VerifiedAccessEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitVerifiedAccessEndpointDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VerifiedAccessEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VerifiedAccessEndpointStatusCodeDeleting, awstypes.VerifiedAccessEndpointStatusCodeActive, awstypes.VerifiedAccessEndpointStatusCodeDeleted),
		Target:  []string{},
		Refresh: statusVerifiedAccessEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VerifiedAccessEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func waitFastSnapshotRestoreCreated(ctx context.Context, conn *ec2.Client, availabilityZone, snapshotID string, timeout time.Duration) (*awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FastSnapshotRestoreStateCodeEnabling, awstypes.FastSnapshotRestoreStateCodeOptimizing),
		Target:  enum.Slice(awstypes.FastSnapshotRestoreStateCodeEnabled),
		Refresh: statusFastSnapshotRestore(ctx, conn, availabilityZone, snapshotID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribeFastSnapshotRestoreSuccessItem); ok {
		return output, err
	}

	return nil, err
}

func waitFastSnapshotRestoreDeleted(ctx context.Context, conn *ec2.Client, availabilityZone, snapshotID string, timeout time.Duration) (*awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FastSnapshotRestoreStateCodeDisabling, awstypes.FastSnapshotRestoreStateCodeOptimizing, awstypes.FastSnapshotRestoreStateCodeEnabled),
		Target:  []string{},
		Refresh: statusFastSnapshotRestore(ctx, conn, availabilityZone, snapshotID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribeFastSnapshotRestoreSuccessItem); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInsightsAnalysisCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.NetworkInsightsAnalysis, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.AnalysisStatusRunning),
		Target:     enum.Slice(awstypes.AnalysisStatusSucceeded),
		Timeout:    timeout,
		Refresh:    statusNetworkInsightsAnalysis(ctx, conn, id),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NetworkInsightsAnalysis); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
