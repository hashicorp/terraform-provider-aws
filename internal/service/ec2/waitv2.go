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
	vpnConnectionCreatedTimeout = 40 * time.Minute
	vpnConnectionDeletedTimeout = 30 * time.Minute
	vpnConnectionUpdatedTimeout = 30 * time.Minute
)

func WaitVPNConnectionCreated(ctx context.Context, conn *ec2.Client, id string) (*types.VpnConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VpnStatePending),
		Target:     enum.Slice(types.VpnStateAvailable),
		Refresh:    StatusVPNConnectionState(ctx, conn, id),
		Timeout:    vpnConnectionCreatedTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpnConnection); ok {
		return output, err
	}

	return nil, err
}

func WaitVPNConnectionDeleted(ctx context.Context, conn *ec2.Client, id string) (*types.VpnConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VpnStateDeleting),
		Target:     []string{},
		Refresh:    StatusVPNConnectionState(ctx, conn, id),
		Timeout:    vpnConnectionDeletedTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpnConnection); ok {
		return output, err
	}

	return nil, err
}

func WaitVPNConnectionUpdated(ctx context.Context, conn *ec2.Client, id string) (*types.VpnConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(vpnStateModifying),
		Target:     enum.Slice(types.VpnStateAvailable),
		Refresh:    StatusVPNConnectionState(ctx, conn, id),
		Timeout:    vpnConnectionUpdatedTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpnConnection); ok {
		return output, err
	}

	return nil, err
}

const (
	vpnConnectionRouteCreatedTimeout = 15 * time.Second
	vpnConnectionRouteDeletedTimeout = 15 * time.Second
)

func WaitVPNConnectionRouteCreated(ctx context.Context, conn *ec2.Client, vpnConnectionID, cidrBlock string) (*types.VpnStaticRoute, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.VpnStatePending),
		Target:  enum.Slice(types.VpnStateAvailable),
		Refresh: StatusVPNConnectionRouteState(ctx, conn, vpnConnectionID, cidrBlock),
		Timeout: vpnConnectionRouteCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpnStaticRoute); ok {
		return output, err
	}

	return nil, err
}

func WaitVPNConnectionRouteDeleted(ctx context.Context, conn *ec2.Client, vpnConnectionID, cidrBlock string) (*types.VpnStaticRoute, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.VpnStatePending, types.VpnStateAvailable, types.VpnStateDeleting),
		Target:  []string{},
		Refresh: StatusVPNConnectionRouteState(ctx, conn, vpnConnectionID, cidrBlock),
		Timeout: vpnConnectionRouteDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpnStaticRoute); ok {
		return output, err
	}

	return nil, err
}

const (
	vpnGatewayCreatedTimeout = 10 * time.Minute
	vpnGatewayDeletedTimeout = 10 * time.Minute
)

func WaitVPNGatewayCreated(ctx context.Context, conn *ec2.Client, id string) (*types.VpnGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VpnStatePending),
		Target:     enum.Slice(types.VpnStateAvailable),
		Refresh:    StatusVPNGatewayState(ctx, conn, id),
		Timeout:    vpnGatewayCreatedTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpnGateway); ok {
		return output, err
	}

	return nil, err
}

func WaitVPNGatewayDeleted(ctx context.Context, conn *ec2.Client, id string) (*types.VpnGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.VpnStateDeleting),
		Target:     []string{},
		Refresh:    StatusVPNGatewayState(ctx, conn, id),
		Timeout:    vpnGatewayDeletedTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpnGateway); ok {
		return output, err
	}

	return nil, err
}

const (
	VPNGatewayDeletedTimeout = 5 * time.Minute

	VPNGatewayVPCAttachmentAttachedTimeout = 15 * time.Minute
	VPNGatewayVPCAttachmentDetachedTimeout = 30 * time.Minute
)

func WaitVPNGatewayVPCAttachmentAttached(ctx context.Context, conn *ec2.Client, vpnGatewayID, vpcID string) (*types.VpcAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AttachmentStatusAttaching),
		Target:  enum.Slice(types.AttachmentStatusAttached),
		Refresh: StatusVPNGatewayVPCAttachmentState(ctx, conn, vpnGatewayID, vpcID),
		Timeout: VPNGatewayVPCAttachmentAttachedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func WaitVPNGatewayVPCAttachmentDetached(ctx context.Context, conn *ec2.Client, vpnGatewayID, vpcID string) (*types.VpcAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AttachmentStatusAttached, types.AttachmentStatusDetaching),
		Target:  []string{},
		Refresh: StatusVPNGatewayVPCAttachmentState(ctx, conn, vpnGatewayID, vpcID),
		Timeout: VPNGatewayVPCAttachmentDetachedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

const (
	customerGatewayCreatedTimeout = 10 * time.Minute
	customerGatewayDeletedTimeout = 5 * time.Minute
)

func WaitCustomerGatewayCreated(ctx context.Context, conn *ec2.Client, id string) (*types.CustomerGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(CustomerGatewayStatePending),
		Target:     enum.Slice(CustomerGatewayStateAvailable),
		Refresh:    StatusCustomerGatewayState(ctx, conn, id),
		Timeout:    customerGatewayCreatedTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.CustomerGateway); ok {
		return output, err
	}

	return nil, err
}

func WaitCustomerGatewayDeleted(ctx context.Context, conn *ec2.Client, id string) (*types.CustomerGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(CustomerGatewayStateAvailable, CustomerGatewayStateDeleting),
		Target:  []string{},
		Refresh: StatusCustomerGatewayState(ctx, conn, id),
		Timeout: customerGatewayDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.CustomerGateway); ok {
		return output, err
	}

	return nil, err
}
