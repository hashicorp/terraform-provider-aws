package directconnect

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	connectionConfirmedTimeout     = 10 * time.Minute
	connectionDeletedTimeout       = 10 * time.Minute
	connectionDisassociatedTimeout = 1 * time.Minute
	hostedConnectionDeletedTimeout = 10 * time.Minute
	lagDeletedTimeout              = 10 * time.Minute
)

func waitConnectionConfirmed(ctx context.Context, conn *directconnect.DirectConnect, id string) (*directconnect.Connection, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.ConnectionStatePending, directconnect.ConnectionStateOrdering, directconnect.ConnectionStateRequested},
		Target:  []string{directconnect.ConnectionStateAvailable},
		Refresh: statusConnectionState(ctx, conn, id),
		Timeout: connectionConfirmedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directconnect.Connection); ok {
		return output, err
	}

	return nil, err
}

func waitConnectionDeleted(ctx context.Context, conn *directconnect.DirectConnect, id string) (*directconnect.Connection, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.ConnectionStatePending, directconnect.ConnectionStateOrdering, directconnect.ConnectionStateAvailable, directconnect.ConnectionStateRequested, directconnect.ConnectionStateDeleting},
		Target:  []string{},
		Refresh: statusConnectionState(ctx, conn, id),
		Timeout: connectionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directconnect.Connection); ok {
		return output, err
	}

	return nil, err
}

func waitGatewayCreated(ctx context.Context, conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.Gateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayStatePending},
		Target:  []string{directconnect.GatewayStateAvailable},
		Refresh: statusGatewayState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directconnect.Gateway); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayDeleted(ctx context.Context, conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.Gateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayStatePending, directconnect.GatewayStateAvailable, directconnect.GatewayStateDeleting},
		Target:  []string{},
		Refresh: statusGatewayState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directconnect.Gateway); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayAssociationCreated(ctx context.Context, conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.GatewayAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayAssociationStateAssociating},
		Target:  []string{directconnect.GatewayAssociationStateAssociated},
		Refresh: statusGatewayAssociationState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directconnect.GatewayAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayAssociationUpdated(ctx context.Context, conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.GatewayAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayAssociationStateUpdating},
		Target:  []string{directconnect.GatewayAssociationStateAssociated},
		Refresh: statusGatewayAssociationState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directconnect.GatewayAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayAssociationDeleted(ctx context.Context, conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.GatewayAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayAssociationStateDisassociating},
		Target:  []string{},
		Refresh: statusGatewayAssociationState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directconnect.GatewayAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitHostedConnectionDeleted(ctx context.Context, conn *directconnect.DirectConnect, id string) (*directconnect.Connection, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.ConnectionStatePending, directconnect.ConnectionStateOrdering, directconnect.ConnectionStateAvailable, directconnect.ConnectionStateRequested, directconnect.ConnectionStateDeleting},
		Target:  []string{},
		Refresh: statusHostedConnectionState(ctx, conn, id),
		Timeout: hostedConnectionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directconnect.Connection); ok {
		return output, err
	}

	return nil, err
}

func waitLagDeleted(ctx context.Context, conn *directconnect.DirectConnect, id string) (*directconnect.Lag, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.LagStateAvailable, directconnect.LagStateRequested, directconnect.LagStatePending, directconnect.LagStateDeleting},
		Target:  []string{},
		Refresh: statusLagState(ctx, conn, id),
		Timeout: lagDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directconnect.Lag); ok {
		return output, err
	}

	return nil, err
}
