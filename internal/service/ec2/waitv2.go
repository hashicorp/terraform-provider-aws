// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
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

const (
	ClientVPNEndpointDeletedTimeout          = 5 * time.Minute
	ClientVPNEndpointAttributeUpdatedTimeout = 5 * time.Minute
)

func WaitClientVPNEndpointDeleted(ctx context.Context, conn *ec2.Client, id string) (*types.ClientVpnEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ClientVpnEndpointStatusCodeDeleting),
		Target:  []string{},
		Refresh: StatusClientVPNEndpointState(ctx, conn, id),
		Timeout: ClientVPNEndpointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ClientVpnEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitClientVPNEndpointClientConnectResponseOptionsUpdated(ctx context.Context, conn *ec2.Client, id string) (*types.ClientConnectResponseOptions, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ClientVpnEndpointAttributeStatusCodeApplying),
		Target:  enum.Slice(types.ClientVpnEndpointAttributeStatusCodeApplied),
		Refresh: StatusClientVPNEndpointClientConnectResponseOptionsState(ctx, conn, id),
		Timeout: ClientVPNEndpointAttributeUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ClientConnectResponseOptions); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

const (
	ClientVPNAuthorizationRuleCreatedTimeout = 10 * time.Minute
	ClientVPNAuthorizationRuleDeletedTimeout = 10 * time.Minute
)

func WaitClientVPNAuthorizationRuleCreated(ctx context.Context, conn *ec2.Client, endpointID, targetNetworkCIDR, accessGroupID string, timeout time.Duration) (*types.AuthorizationRule, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ClientVpnAuthorizationRuleStatusCodeAuthorizing),
		Target:  enum.Slice(types.ClientVpnAuthorizationRuleStatusCodeActive),
		Refresh: StatusClientVPNAuthorizationRule(ctx, conn, endpointID, targetNetworkCIDR, accessGroupID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AuthorizationRule); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitClientVPNAuthorizationRuleDeleted(ctx context.Context, conn *ec2.Client, endpointID, targetNetworkCIDR, accessGroupID string, timeout time.Duration) (*types.AuthorizationRule, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ClientVpnAuthorizationRuleStatusCodeRevoking),
		Target:  []string{},
		Refresh: StatusClientVPNAuthorizationRule(ctx, conn, endpointID, targetNetworkCIDR, accessGroupID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AuthorizationRule); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

const (
	ClientVPNNetworkAssociationCreatedTimeout     = 30 * time.Minute
	ClientVPNNetworkAssociationCreatedDelay       = 4 * time.Minute
	ClientVPNNetworkAssociationDeletedTimeout     = 30 * time.Minute
	ClientVPNNetworkAssociationDeletedDelay       = 4 * time.Minute
	ClientVPNNetworkAssociationStatusPollInterval = 10 * time.Second
)

func WaitClientVPNNetworkAssociationCreated(ctx context.Context, conn *ec2.Client, associationID, endpointID string, timeout time.Duration) (*types.TargetNetwork, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(types.AssociationStatusCodeAssociating),
		Target:       enum.Slice(types.AssociationStatusCodeAssociated),
		Refresh:      StatusClientVPNNetworkAssociation(ctx, conn, associationID, endpointID),
		Timeout:      timeout,
		Delay:        ClientVPNNetworkAssociationCreatedDelay,
		PollInterval: ClientVPNNetworkAssociationStatusPollInterval,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.TargetNetwork); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitClientVPNNetworkAssociationDeleted(ctx context.Context, conn *ec2.Client, associationID, endpointID string, timeout time.Duration) (*types.TargetNetwork, error) {
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(types.AssociationStatusCodeDisassociating),
		Target:       []string{},
		Refresh:      StatusClientVPNNetworkAssociation(ctx, conn, associationID, endpointID),
		Timeout:      timeout,
		Delay:        ClientVPNNetworkAssociationDeletedDelay,
		PollInterval: ClientVPNNetworkAssociationStatusPollInterval,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.TargetNetwork); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitClientVPNRouteCreated(ctx context.Context, conn *ec2.Client, endpointID, targetSubnetID, destinationCIDR string, timeout time.Duration) (*types.ClientVpnRoute, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ClientVpnRouteStatusCodeCreating),
		Target:  enum.Slice(types.ClientVpnRouteStatusCodeActive),
		Refresh: StatusClientVPNRoute(ctx, conn, endpointID, targetSubnetID, destinationCIDR),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ClientVpnRoute); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitClientVPNRouteDeleted(ctx context.Context, conn *ec2.Client, endpointID, targetSubnetID, destinationCIDR string, timeout time.Duration) (*types.ClientVpnRoute, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ClientVpnRouteStatusCodeActive, types.ClientVpnRouteStatusCodeDeleting),
		Target:  []string{},
		Refresh: StatusClientVPNRoute(ctx, conn, endpointID, targetSubnetID, destinationCIDR),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ClientVpnRoute); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.Message)))

		return output, err
	}

	return nil, err
}
