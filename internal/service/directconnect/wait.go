// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	connectionDisassociatedTimeout = 1 * time.Minute
	hostedConnectionDeletedTimeout = 10 * time.Minute
	lagDeletedTimeout              = 10 * time.Minute
)

func waitGatewayAssociationCreated(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectConnectGatewayAssociationStateAssociating),
		Target:  enum.Slice(awstypes.DirectConnectGatewayAssociationStateAssociated),
		Refresh: statusGatewayAssociationState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGatewayAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayAssociationUpdated(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectConnectGatewayAssociationStateUpdating),
		Target:  enum.Slice(awstypes.DirectConnectGatewayAssociationStateAssociated),
		Refresh: statusGatewayAssociationState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGatewayAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayAssociationDeleted(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.DirectConnectGatewayAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectConnectGatewayAssociationStateDisassociating),
		Target:  []string{},
		Refresh: statusGatewayAssociationState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectConnectGatewayAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitHostedConnectionDeleted(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Connection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectionStatePending, awstypes.ConnectionStateOrdering, awstypes.ConnectionStateAvailable, awstypes.ConnectionStateRequested, awstypes.ConnectionStateDeleting),
		Target:  []string{},
		Refresh: statusHostedConnectionState(ctx, conn, id),
		Timeout: hostedConnectionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Connection); ok {
		return output, err
	}

	return nil, err
}

func waitLagDeleted(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Lag, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LagStateAvailable, awstypes.LagStateRequested, awstypes.LagStatePending, awstypes.LagStateDeleting),
		Target:  []string{},
		Refresh: statusLagState(ctx, conn, id),
		Timeout: lagDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Lag); ok {
		return output, err
	}

	return nil, err
}
