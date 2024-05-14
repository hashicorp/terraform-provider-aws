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
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

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

func WaitIPAMCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.Ipam, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamStateCreateInProgress),
		Target:  enum.Slice(types.IpamStateCreateComplete),
		Refresh: StatusIPAMState(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Ipam); ok {
		return output, err
	}

	return nil, err
}

func WaitIPAMDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.Ipam, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamStateCreateComplete, types.IpamStateModifyComplete, types.IpamStateDeleteInProgress),
		Target:  []string{},
		Refresh: StatusIPAMState(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Ipam); ok {
		return output, err
	}

	return nil, err
}

func WaitIPAMUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.Ipam, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamStateModifyInProgress),
		Target:  enum.Slice(types.IpamStateModifyComplete),
		Refresh: StatusIPAMState(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Ipam); ok {
		return output, err
	}

	return nil, err
}

func WaitIPAMPoolCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.IpamPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamPoolStateCreateInProgress),
		Target:  enum.Slice(types.IpamPoolStateCreateComplete),
		Refresh: StatusIPAMPoolState(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamPool); ok {
		if state := output.State; state == types.IpamPoolStateCreateFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitIPAMPoolDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.IpamPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamPoolStateDeleteInProgress),
		Target:  []string{},
		Refresh: StatusIPAMPoolState(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamPool); ok {
		if state := output.State; state == types.IpamPoolStateDeleteFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitIPAMPoolUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.IpamPool, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamPoolStateModifyInProgress),
		Target:  enum.Slice(types.IpamPoolStateModifyComplete),
		Refresh: StatusIPAMPoolState(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamPool); ok {
		if state := output.State; state == types.IpamPoolStateModifyFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitIPAMPoolCIDRIdCreated(ctx context.Context, conn *ec2.Client, poolCidrId, poolID, cidrBlock string, timeout time.Duration) (*types.IpamPoolCidr, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(types.IpamPoolCidrStatePendingProvision),
		Target:         enum.Slice(types.IpamPoolCidrStateProvisioned),
		Refresh:        StatusIPAMPoolCIDRState(ctx, conn, cidrBlock, poolID, poolCidrId),
		Timeout:        timeout,
		Delay:          5 * time.Second,
		NotFoundChecks: IPAMPoolCIDRNotFoundChecks,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamPoolCidr); ok {
		if state, failureReason := output.State, output.FailureReason; state == types.IpamPoolCidrStateFailedProvision && failureReason != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", string(failureReason.Code), aws.ToString(failureReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitIPAMPoolCIDRDeleted(ctx context.Context, conn *ec2.Client, cidrBlock, poolID, poolCidrId string, timeout time.Duration) (*types.IpamPoolCidr, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamPoolCidrStatePendingDeprovision, types.IpamPoolCidrStateProvisioned),
		Target:  []string{},
		Refresh: StatusIPAMPoolCIDRState(ctx, conn, cidrBlock, poolID, poolCidrId),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamPoolCidr); ok {
		if state, failureReason := output.State, output.FailureReason; state == types.IpamPoolCidrStateFailedDeprovision && failureReason != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", string(failureReason.Code), aws.ToString(failureReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitIPAMPoolCIDRAllocationCreated(ctx context.Context, conn *ec2.Client, allocationID, poolID string, timeout time.Duration) (*types.IpamPoolAllocation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(IpamPoolCIDRAllocationCreateComplete),
		Refresh: StatusIPAMPoolCIDRAllocationState(ctx, conn, allocationID, poolID),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamPoolAllocation); ok {
		return output, err
	}

	return nil, err
}

func WaitIPAMResourceDiscoveryAvailable(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.Ipam, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamResourceDiscoveryStateCreateInProgress),
		Target:  enum.Slice(types.IpamResourceDiscoveryStateCreateComplete),
		Refresh: StatusIPAMResourceDiscoveryState(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Ipam); ok {
		return output, err
	}

	return nil, err
}

func WaitIPAMResourceDiscoveryAssociationAvailable(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.IpamResourceDiscoveryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamResourceDiscoveryAssociationStateAssociateInProgress),
		Target:  enum.Slice(types.IpamResourceDiscoveryAssociationStateAssociateComplete),
		Refresh: StatusIPAMResourceDiscoveryAssociationStatus(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamResourceDiscoveryAssociation); ok {
		return output, err
	}

	return nil, err
}

func WaitIPAMResourceDiscoveryAssociationDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.IpamResourceDiscoveryAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamResourceDiscoveryAssociationStateAssociateComplete, types.IpamResourceDiscoveryAssociationStateDisassociateInProgress),
		Target:  []string{},
		Refresh: StatusIPAMResourceDiscoveryAssociationStatus(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamResourceDiscoveryAssociation); ok {
		return output, err
	}

	return nil, err
}

func WaitIPAMResourceDiscoveryDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.IpamResourceDiscovery, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamResourceDiscoveryStateCreateComplete, types.IpamResourceDiscoveryStateModifyComplete, types.IpamResourceDiscoveryStateDeleteInProgress),
		Target:  []string{},
		Refresh: StatusIPAMResourceDiscoveryState(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamResourceDiscovery); ok {
		return output, err
	}

	return nil, err
}

func WaitIPAMResourceDiscoveryUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.IpamResourceDiscovery, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamResourceDiscoveryStateModifyInProgress),
		Target:  enum.Slice(types.IpamResourceDiscoveryStateModifyComplete),
		Refresh: StatusIPAMResourceDiscoveryState(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamResourceDiscovery); ok {
		return output, err
	}

	return nil, err
}

func WaitIPAMScopeCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.IpamScope, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamScopeStateCreateInProgress),
		Target:  enum.Slice(types.IpamScopeStateCreateComplete),
		Refresh: StatusIPAMScopeState(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamScope); ok {
		return output, err
	}

	return nil, err
}

func WaitIPAMScopeDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.IpamScope, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamScopeStateCreateComplete, types.IpamScopeStateModifyComplete, types.IpamScopeStateDeleteInProgress),
		Target:  []string{},
		Refresh: StatusIPAMScopeState(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamScope); ok {
		return output, err
	}

	return nil, err
}

func WaitIPAMScopeUpdated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*types.IpamScope, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.IpamScopeStateModifyInProgress),
		Target:  enum.Slice(types.IpamScopeStateModifyComplete),
		Refresh: StatusIPAMScopeState(ctx, conn, id),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.IpamScope); ok {
		return output, err
	}

	return nil, err
}
