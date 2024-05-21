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
		Refresh: statusCapacityReservationState(ctx, conn, id),
		Timeout: CapacityReservationActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitCapacityReservationDeleted(ctx context.Context, conn *ec2.Client, id string) (*awstypes.CapacityReservation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapacityReservationStateActive),
		Target:  []string{},
		Refresh: statusCapacityReservationState(ctx, conn, id),
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
		Refresh:    statusFleetState(ctx, conn, id),
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
		Refresh: statusHostState(ctx, conn, id),
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
		Refresh: statusHostState(ctx, conn, id),
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
		Refresh: statusHostState(ctx, conn, id),
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
		Refresh:    statusInstanceMetadataOptionsState(ctx, conn, id),
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
		Refresh: statusPlacementGroupState(ctx, conn, name),
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
		Refresh: statusPlacementGroupState(ctx, conn, name),
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

func waitVPCCreatedV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Vpc, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.VpcStatePending),
		Target:  enum.Slice(awstypes.VpcStateAvailable),
		Refresh: statusVPCStateV2(ctx, conn, id),
		Timeout: vpcCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Vpc); ok {
		return output, err
	}

	return nil, err
}

func waitVPCIPv6CIDRBlockAssociationCreatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcCidrBlockState, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VpcCidrBlockStateCodeAssociating, awstypes.VpcCidrBlockStateCodeDisassociated, awstypes.VpcCidrBlockStateCodeFailing),
		Target:     enum.Slice(awstypes.VpcCidrBlockStateCodeAssociated),
		Refresh:    statusVPCIPv6CIDRBlockAssociationStateV2(ctx, conn, id),
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

func waitVPCAttributeUpdatedV2(ctx context.Context, conn *ec2.Client, vpcID string, attribute awstypes.VpcAttributeName, expectedValue bool) (*awstypes.Vpc, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    statusVPCAttributeValueV2(ctx, conn, vpcID, attribute),
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

func waitVPCIPv6CIDRBlockAssociationDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.VpcCidrBlockStateCodeAssociated, awstypes.VpcCidrBlockStateCodeDisassociating, awstypes.VpcCidrBlockStateCodeFailing),
		Target:     []string{},
		Refresh:    statusVPCIPv6CIDRBlockAssociationStateV2(ctx, conn, id),
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

func waitNetworkInterfaceAvailableAfterUseV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.NetworkInterface, error) {
	// Hyperplane attached ENI.
	// Wait for it to be moved into a removable state.
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.NetworkInterfaceStatusInUse),
		Target:     enum.Slice(awstypes.NetworkInterfaceStatusAvailable),
		Timeout:    timeout,
		Refresh:    statusNetworkInterfaceV2(ctx, conn, id),
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

func waitNetworkInterfaceCreatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.NetworkInterface, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{NetworkInterfaceStatusPending},
		Target:  enum.Slice(awstypes.NetworkInterfaceStatusAvailable),
		Timeout: timeout,
		Refresh: statusNetworkInterfaceV2(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceAttachedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.NetworkInterfaceAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStatusAttaching),
		Target:  enum.Slice(awstypes.AttachmentStatusAttached),
		Timeout: timeout,
		Refresh: statusNetworkInterfaceAttachmentV2(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceDetachedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.NetworkInterfaceAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStatusAttached, awstypes.AttachmentStatusDetaching),
		Target:  enum.Slice(awstypes.AttachmentStatusDetached),
		Timeout: timeout,
		Refresh: statusNetworkInterfaceAttachmentV2(ctx, conn, id),
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
		Refresh:    statusVolumeState(ctx, conn, id),
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
		Refresh:    statusVolumeState(ctx, conn, id),
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
		Refresh:    statusVolumeState(ctx, conn, id),
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
		Refresh:    statusVolumeAttachmentState(ctx, conn, volumeID, instanceID, deviceName),
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
		Refresh:    statusVolumeModificationState(ctx, conn, id),
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

func waitVPCEndpointAcceptedV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, timeout time.Duration) (*awstypes.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpcEndpointStatePendingAcceptance),
		Target:     enum.Slice(vpcEndpointStateAvailable),
		Timeout:    timeout,
		Refresh:    statusVPCEndpointStateV2(ctx, conn, vpcEndpointID),
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

func waitVPCEndpointAvailableV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, timeout time.Duration) (*awstypes.VpcEndpoint, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpcEndpointStatePending),
		Target:     enum.Slice(vpcEndpointStateAvailable, vpcEndpointStatePendingAcceptance),
		Timeout:    timeout,
		Refresh:    statusVPCEndpointStateV2(ctx, conn, vpcEndpointID),
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

func waitVPCEndpointDeletedV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, timeout time.Duration) (*awstypes.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpcEndpointStateDeleting, vpcEndpointStateDeleted),
		Target:     []string{},
		Refresh:    statusVPCEndpointStateV2(ctx, conn, vpcEndpointID),
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

func waitRouteDeletedV2(ctx context.Context, conn *ec2.Client, routeFinder routeFinderV2, routeTableID, destination string, timeout time.Duration) (*awstypes.Route, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{RouteStatusReady},
		Target:                    []string{},
		Refresh:                   statusRouteV2(ctx, conn, routeFinder, routeTableID, destination),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Route); ok {
		return output, err
	}

	return nil, err
}

func waitRouteReadyV2(ctx context.Context, conn *ec2.Client, routeFinder routeFinderV2, routeTableID, destination string, timeout time.Duration) (*awstypes.Route, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{RouteStatusReady},
		Refresh:                   statusRouteV2(ctx, conn, routeFinder, routeTableID, destination),
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

func waitRouteTableReadyV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{RouteTableStatusReady},
		Refresh:                   statusRouteTableV2(ctx, conn, id),
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

func waitRouteTableDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{RouteTableStatusReady},
		Target:                    []string{},
		Refresh:                   statusRouteTableV2(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.RouteTable); ok {
		return output, err
	}

	return nil, err
}

func waitRouteTableAssociationCreatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteTableAssociationState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.RouteTableAssociationStateCodeAssociating),
		Target:         enum.Slice(awstypes.RouteTableAssociationStateCodeAssociated),
		Refresh:        statusRouteTableAssociationStateV2(ctx, conn, id),
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

func waitRouteTableAssociationDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteTableAssociationState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RouteTableAssociationStateCodeDisassociating, awstypes.RouteTableAssociationStateCodeAssociated),
		Target:  []string{},
		Refresh: statusRouteTableAssociationStateV2(ctx, conn, id),
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

func waitRouteTableAssociationUpdatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.RouteTableAssociationState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RouteTableAssociationStateCodeAssociating),
		Target:  enum.Slice(awstypes.RouteTableAssociationStateCodeAssociated),
		Refresh: statusRouteTableAssociationStateV2(ctx, conn, id),
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
		Refresh:    statusSpotFleetRequestState(ctx, conn, id),
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
		Refresh:    statusSpotFleetRequestState(ctx, conn, id),
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

func waitVPCEndpointServiceAvailableV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.ServiceConfiguration, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ServiceStatePending),
		Target:     enum.Slice(awstypes.ServiceStateAvailable),
		Refresh:    statusVPCEndpointServiceStateAvailableV2(ctx, conn, id),
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

func waitVPCEndpointServiceDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.ServiceConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ServiceStateAvailable, awstypes.ServiceStateDeleting),
		Target:     []string{},
		Timeout:    timeout,
		Refresh:    statusVPCEndpointServiceStateDeletedV2(ctx, conn, id),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ServiceConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointRouteTableAssociationReadyV2(ctx context.Context, conn *ec2.Client, vpcEndpointID, routeTableID string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(VPCEndpointRouteTableAssociationStatusReady),
		Refresh:                   statusVPCEndpointRouteTableAssociationV2(ctx, conn, vpcEndpointID, routeTableID),
		Timeout:                   ec2PropagationTimeout,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitVPCEndpointRouteTableAssociationDeletedV2(ctx context.Context, conn *ec2.Client, vpcEndpointID, routeTableID string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(VPCEndpointRouteTableAssociationStatusReady),
		Target:                    []string{},
		Refresh:                   statusVPCEndpointRouteTableAssociationV2(ctx, conn, vpcEndpointID, routeTableID),
		Timeout:                   ec2PropagationTimeout,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitVPCEndpointConnectionAcceptedV2(ctx context.Context, conn *ec2.Client, serviceID, vpcEndpointID string, timeout time.Duration) (*awstypes.VpcEndpointConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{vpcEndpointStatePendingAcceptance, vpcEndpointStatePending},
		Target:     []string{vpcEndpointStateAvailable},
		Refresh:    statusVPCEndpointConnectionVPCEndpointStateV2(ctx, conn, serviceID, vpcEndpointID),
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

func waitVPCEndpointServicePrivateDNSNameVerifiedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.PrivateDnsNameConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.DnsNameStatePendingVerification),
		Target:                    enum.Slice(awstypes.DnsNameStateVerified),
		Refresh:                   statusVPCEndpointServicePrivateDNSNameConfigurationV2(ctx, conn, id),
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
		Refresh: statusClientVPNEndpointState(ctx, conn, id),
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
		Refresh: statusClientVPNEndpointClientConnectResponseOptionsState(ctx, conn, id),
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

func WaitImageAvailable(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Image, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ImageStatePending),
		Target:     enum.Slice(awstypes.ImageStateAvailable),
		Refresh:    StatusImageState(ctx, conn, id),
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

func WaitImageDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.Image, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.ImageStateAvailable, awstypes.ImageStateFailed, awstypes.ImageStatePending),
		Target:     []string{},
		Refresh:    StatusImageState(ctx, conn, id),
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
