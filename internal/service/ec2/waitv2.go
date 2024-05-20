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
	ec2_sdkv2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	AvailabilityZoneGroupOptInStatusTimeout = 10 * time.Minute
)

func waitAvailabilityZoneGroupOptedIn(ctx context.Context, conn *ec2.Client, name string) (*types.AvailabilityZone, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AvailabilityZoneOptInStatusNotOptedIn),
		Target:  enum.Slice(types.AvailabilityZoneOptInStatusOptedIn),
		Refresh: statusAvailabilityZoneGroupOptInStatus(ctx, conn, name),
		Timeout: AvailabilityZoneGroupOptInStatusTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AvailabilityZone); ok {
		return output, err
	}

	return nil, err
}

func waitAvailabilityZoneGroupNotOptedIn(ctx context.Context, conn *ec2.Client, name string) (*types.AvailabilityZone, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AvailabilityZoneOptInStatusOptedIn),
		Target:  enum.Slice(types.AvailabilityZoneOptInStatusNotOptedIn),
		Refresh: statusAvailabilityZoneGroupOptInStatus(ctx, conn, name),
		Timeout: AvailabilityZoneGroupOptInStatusTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AvailabilityZone); ok {
		return output, err
	}

	return nil, err
}

const (
	CapacityReservationActiveTimeout  = 2 * time.Minute
	CapacityReservationDeletedTimeout = 2 * time.Minute
)

func waitCapacityReservationActive(ctx context.Context, conn *ec2.Client, id string) (*types.CapacityReservation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.CapacityReservationStatePending),
		Target:  enum.Slice(types.CapacityReservationStateActive),
		Refresh: statusCapacityReservationState(ctx, conn, id),
		Timeout: CapacityReservationActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.CapacityReservation); ok {
		return output, err
	}

	return nil, err
}

func waitCapacityReservationDeleted(ctx context.Context, conn *ec2.Client, id string) (*types.CapacityReservation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.CapacityReservationStateActive),
		Target:  []string{},
		Refresh: statusCapacityReservationState(ctx, conn, id),
		Timeout: CapacityReservationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.CapacityReservation); ok {
		return output, err
	}

	return nil, err
}

func waitFleet(ctx context.Context, conn *ec2.Client, id string, pending, target []string, timeout, delay time.Duration) (*awstypes.FleetData, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    statusFleetState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      delay,
		MinTimeout: 1 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FleetData); ok {
		return output, err
	}

	return nil, err
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

func waitInstanceIAMInstanceProfileUpdated(ctx context.Context, conn *ec2.Client, instanceID string, expectedValue string) (*types.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Target:     enum.Slice(expectedValue),
		Refresh:    statusInstanceIAMInstanceProfile(ctx, conn, instanceID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Instance); ok {
		return output, err
	}

	return nil, err
}

func waitInstanceCapacityReservationSpecificationUpdated(ctx context.Context, conn *ec2.Client, instanceID string, expectedValue *types.CapacityReservationSpecification) (*types.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Target:     enum.Slice(strconv.FormatBool(true)),
		Refresh:    statusInstanceCapacityReservationSpecificationEquals(ctx, conn, instanceID, expectedValue),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Instance); ok {
		return output, err
	}

	return nil, err
}

func waitInstanceMaintenanceOptionsAutoRecoveryUpdated(ctx context.Context, conn *ec2.Client, id, expectedValue string, timeout time.Duration) (*types.InstanceMaintenanceOptions, error) {
	stateConf := &retry.StateChangeConf{
		Target:     enum.Slice(expectedValue),
		Refresh:    statusInstanceMaintenanceOptionsAutoRecovery(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.InstanceMaintenanceOptions); ok {
		return output, err
	}

	return nil, err
}

func waitInstanceMetadataOptionsApplied(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.InstanceMetadataOptionsResponse, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.InstanceMetadataOptionsStatePending),
		Target:     enum.Slice(types.InstanceMetadataOptionsStateApplied),
		Refresh:    statusInstanceMetadataOptionsState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.InstanceMetadataOptionsResponse); ok {
		return output, err
	}

	return nil, err
}

func waitInstanceRootBlockDeviceDeleteOnTerminationUpdated(ctx context.Context, conn *ec2.Client, id string, expectedValue bool, timeout time.Duration) (*types.EbsInstanceBlockDevice, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    statusInstanceRootBlockDeviceDeleteOnTermination(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.EbsInstanceBlockDevice); ok {
		return output, err
	}

	return nil, err
}

func waitSpotInstanceRequestFulfilled(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.SpotInstanceRequest, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{spotInstanceRequestStatusCodePendingEvaluation, spotInstanceRequestStatusCodePendingFulfillment},
		Target:     []string{spotInstanceRequestStatusCodeFulfilled},
		Refresh:    statusSpotInstanceRequest(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.SpotInstanceRequest); ok {
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

func waitVPCCreatedV2(ctx context.Context, conn *ec2.Client, id string) (*types.Vpc, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.VpcStatePending),
		Target:  enum.Slice(types.VpcStateAvailable),
		Refresh: statusVPCStateV2(ctx, conn, id),
		Timeout: vpcCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Vpc); ok {
		return output, err
	}

	return nil, err
}

func waitVPCIPv6CIDRBlockAssociationCreatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.VpcCidrBlockState, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VpcCidrBlockStateCodeAssociating, types.VpcCidrBlockStateCodeDisassociated, types.VpcCidrBlockStateCodeFailing),
		Target:     enum.Slice(types.VpcCidrBlockStateCodeAssociated),
		Refresh:    statusVPCIPv6CIDRBlockAssociationStateV2(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcCidrBlockState); ok {
		if state := output.State; state == types.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCAttributeUpdatedV2(ctx context.Context, conn *ec2.Client, vpcID string, attribute types.VpcAttributeName, expectedValue bool) (*types.Vpc, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    statusVPCAttributeValueV2(ctx, conn, vpcID, attribute),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Vpc); ok {
		return output, err
	}

	return nil, err
}

func waitVPCIPv6CIDRBlockAssociationDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VpcCidrBlockStateCodeAssociated, types.VpcCidrBlockStateCodeDisassociating, types.VpcCidrBlockStateCodeFailing),
		Target:     []string{},
		Refresh:    statusVPCIPv6CIDRBlockAssociationStateV2(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcCidrBlockState); ok {
		if state := output.State; state == types.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceAvailableAfterUseV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.NetworkInterface, error) {
	// Hyperplane attached ENI.
	// Wait for it to be moved into a removable state.
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.NetworkInterfaceStatusInUse),
		Target:     enum.Slice(types.NetworkInterfaceStatusAvailable),
		Timeout:    timeout,
		Refresh:    statusNetworkInterfaceV2(ctx, conn, id),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
		// Handle EC2 ENI eventual consistency. It can take up to 3 minutes.
		ContinuousTargetOccurence: 18,
		NotFoundChecks:            1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceCreatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.NetworkInterface, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{NetworkInterfaceStatusPending},
		Target:  enum.Slice(types.NetworkInterfaceStatusAvailable),
		Timeout: timeout,
		Refresh: statusNetworkInterfaceV2(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceAttachedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.NetworkInterfaceAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AttachmentStatusAttaching),
		Target:  enum.Slice(types.AttachmentStatusAttached),
		Timeout: timeout,
		Refresh: statusNetworkInterfaceAttachmentV2(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitNetworkInterfaceDetachedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.NetworkInterfaceAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AttachmentStatusAttached, types.AttachmentStatusDetaching),
		Target:  enum.Slice(types.AttachmentStatusDetached),
		Timeout: timeout,
		Refresh: statusNetworkInterfaceAttachmentV2(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVolumeCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.Volume, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VolumeStateCreating),
		Target:     enum.Slice(types.VolumeStateAvailable),
		Refresh:    statusVolumeState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Volume); ok {
		return output, err
	}

	return nil, err
}

func waitVolumeDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.Volume, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VolumeStateDeleting),
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

func waitVolumeUpdated(ctx context.Context, conn *ec2_sdkv2.Client, id string, timeout time.Duration) (*types.Volume, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VolumeStateCreating, types.VolumeState(types.VolumeModificationStateModifying)),
		Target:     enum.Slice(types.VolumeStateAvailable, types.VolumeStateInUse),
		Refresh:    statusVolumeState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Volume); ok {
		return output, err
	}

	return nil, err
}

func waitVolumeAttachmentCreated(ctx context.Context, conn *ec2.Client, volumeID, instanceID, deviceName string, timeout time.Duration) (*types.VolumeAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VolumeAttachmentStateAttaching),
		Target:     enum.Slice(types.VolumeAttachmentStateAttached),
		Refresh:    statusVolumeAttachmentState(ctx, conn, volumeID, instanceID, deviceName),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VolumeAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVolumeModificationComplete(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.VolumeModification, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.VolumeModificationStateModifying),
		// The volume is useable once the state is "optimizing", but will not be at full performance.
		// Optimization can take hours. e.g. a full 1 TiB drive takes approximately 6 hours to optimize,
		// according to https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/monitoring-volume-modifications.html.
		Target:     enum.Slice(types.VolumeModificationStateCompleted, types.VolumeModificationStateOptimizing),
		Refresh:    statusVolumeModificationState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VolumeModification); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitVPCEndpointAcceptedV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, timeout time.Duration) (*types.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpcEndpointStatePendingAcceptance),
		Target:     enum.Slice(vpcEndpointStateAvailable),
		Timeout:    timeout,
		Refresh:    statusVPCEndpointStateV2(ctx, conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcEndpoint); ok {
		if state, lastError := output.State, output.LastError; state == types.StateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(lastError.Code), aws.ToString(lastError.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCEndpointAvailableV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, timeout time.Duration) (*types.VpcEndpoint, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpcEndpointStatePending),
		Target:     enum.Slice(vpcEndpointStateAvailable, vpcEndpointStatePendingAcceptance),
		Timeout:    timeout,
		Refresh:    statusVPCEndpointStateV2(ctx, conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcEndpoint); ok {
		if state, lastError := output.State, output.LastError; state == types.StateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(lastError.Code), aws.ToString(lastError.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVPCEndpointDeletedV2(ctx context.Context, conn *ec2.Client, vpcEndpointID string, timeout time.Duration) (*types.VpcEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpcEndpointStateDeleting, vpcEndpointStateDeleted),
		Target:     []string{},
		Refresh:    statusVPCEndpointStateV2(ctx, conn, vpcEndpointID),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcEndpoint); ok {
		return output, err
	}

	return nil, err
}

func waitRouteDeletedV2(ctx context.Context, conn *ec2.Client, routeFinder routeFinderV2, routeTableID, destination string, timeout time.Duration) (*types.Route, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{RouteStatusReady},
		Target:                    []string{},
		Refresh:                   statusRouteV2(ctx, conn, routeFinder, routeTableID, destination),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Route); ok {
		return output, err
	}

	return nil, err
}

func waitRouteReadyV2(ctx context.Context, conn *ec2.Client, routeFinder routeFinderV2, routeTableID, destination string, timeout time.Duration) (*types.Route, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{RouteStatusReady},
		Refresh:                   statusRouteV2(ctx, conn, routeFinder, routeTableID, destination),
		Timeout:                   timeout,
		NotFoundChecks:            RouteNotFoundChecks,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Route); ok {
		return output, err
	}

	return nil, err
}

func waitRouteTableReadyV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.RouteTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{RouteTableStatusReady},
		Refresh:                   statusRouteTableV2(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            RouteTableNotFoundChecks,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RouteTable); ok {
		return output, err
	}

	return nil, err
}

func waitRouteTableDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.RouteTable, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{RouteTableStatusReady},
		Target:                    []string{},
		Refresh:                   statusRouteTableV2(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RouteTable); ok {
		return output, err
	}

	return nil, err
}

func waitRouteTableAssociationCreatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.RouteTableAssociationState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(types.RouteTableAssociationStateCodeAssociating),
		Target:         enum.Slice(types.RouteTableAssociationStateCodeAssociated),
		Refresh:        statusRouteTableAssociationStateV2(ctx, conn, id),
		Timeout:        timeout,
		NotFoundChecks: RouteTableAssociationCreatedNotFoundChecks,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RouteTableAssociationState); ok {
		if output.State == types.RouteTableAssociationStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitRouteTableAssociationDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.RouteTableAssociationState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.RouteTableAssociationStateCodeDisassociating, types.RouteTableAssociationStateCodeAssociated),
		Target:  []string{},
		Refresh: statusRouteTableAssociationStateV2(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RouteTableAssociationState); ok {
		if output.State == types.RouteTableAssociationStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitRouteTableAssociationUpdatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.RouteTableAssociationState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.RouteTableAssociationStateCodeAssociating),
		Target:  enum.Slice(types.RouteTableAssociationStateCodeAssociated),
		Refresh: statusRouteTableAssociationStateV2(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RouteTableAssociationState); ok {
		if output.State == types.RouteTableAssociationStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitSpotFleetRequestCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.SpotFleetRequestConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.BatchStateSubmitted),
		Target:     enum.Slice(types.BatchStateActive),
		Refresh:    statusSpotFleetRequestState(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.SpotFleetRequestConfig); ok {
		return output, err
	}

	return nil, err
}

func waitSpotFleetRequestFulfilled(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.SpotFleetRequestConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ActivityStatusPendingFulfillment),
		Target:     enum.Slice(types.ActivityStatusFulfilled),
		Refresh:    statusSpotFleetActivityStatus(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.SpotFleetRequestConfig); ok {
		if output.ActivityStatus == types.ActivityStatusError {
			var errs []error

			input := &ec2.DescribeSpotFleetRequestHistoryInput{
				SpotFleetRequestId: aws.String(id),
				StartTime:          aws.Time(time.UnixMilli(0)),
			}

			if output, err := findSpotFleetRequestHistoryRecords(ctx, conn, input); err == nil {
				for _, v := range output {
					if eventType := v.EventType; eventType == types.EventTypeError || eventType == types.EventTypeInformation {
						errs = append(errs, errors.New(string(v.EventType)))
					}
				}
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return output, err
	}

	return nil, err
}

func waitSpotFleetRequestUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.SpotFleetRequestConfig, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.BatchStateModifying),
		Target:     enum.Slice(types.BatchStateActive),
		Refresh:    statusSpotFleetRequestState(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.SpotFleetRequestConfig); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointServiceAvailableV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.ServiceConfiguration, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ServiceStatePending),
		Target:     enum.Slice(types.ServiceStateAvailable),
		Refresh:    statusVPCEndpointServiceStateAvailableV2(ctx, conn, id),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ServiceConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitVPCEndpointServiceDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.ServiceConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ServiceStateAvailable, types.ServiceStateDeleting),
		Target:     []string{},
		Timeout:    timeout,
		Refresh:    statusVPCEndpointServiceStateDeletedV2(ctx, conn, id),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ServiceConfiguration); ok {
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

func waitVPCEndpointConnectionAcceptedV2(ctx context.Context, conn *ec2.Client, serviceID, vpcEndpointID string, timeout time.Duration) (*types.VpcEndpointConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{vpcEndpointStatePendingAcceptance, vpcEndpointStatePending},
		Target:     []string{vpcEndpointStateAvailable},
		Refresh:    statusVPCEndpointConnectionVPCEndpointStateV2(ctx, conn, serviceID, vpcEndpointID),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcEndpointConnection); ok {
		return output, err
	}

	return nil, err
}

func WaitImageAvailable(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.Image, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ImageStatePending),
		Target:     enum.Slice(types.ImageStateAvailable),
		Refresh:    StatusImageState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      amiRetryDelay,
		MinTimeout: amiRetryMinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Image); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitImageDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.Image, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.ImageStateAvailable, types.ImageStateFailed, types.ImageStatePending),
		Target:     []string{},
		Refresh:    StatusImageState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      amiRetryDelay,
		MinTimeout: amiRetryMinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Image); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}
