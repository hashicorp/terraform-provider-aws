package waiter

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	connectionDeletedTimeout       = 10 * time.Minute
	connectionDisassociatedTimeout = 1 * time.Minute
	lagDeletedTimeout              = 10 * time.Minute
)

func waitConnectionDeleted(conn *directconnect.DirectConnect, id string) (*directconnect.Connection, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.ConnectionStatePending, directconnect.ConnectionStateOrdering, directconnect.ConnectionStateAvailable, directconnect.ConnectionStateRequested, directconnect.ConnectionStateDeleting},
		Target:  []string{},
		Refresh: statusConnectionState(conn, id),
		Timeout: connectionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.Connection); ok {
		return output, err
	}

	return nil, err
}

func waitGatewayCreated(conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.Gateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayStatePending},
		Target:  []string{directconnect.GatewayStateAvailable},
		Refresh: statusGatewayState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.Gateway); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayDeleted(conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.Gateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayStatePending, directconnect.GatewayStateAvailable, directconnect.GatewayStateDeleting},
		Target:  []string{},
		Refresh: statusGatewayState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.Gateway); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayAssociationCreated(conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.GatewayAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayAssociationStateAssociating},
		Target:  []string{directconnect.GatewayAssociationStateAssociated},
		Refresh: statusGatewayAssociationState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.GatewayAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayAssociationUpdated(conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.GatewayAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayAssociationStateUpdating},
		Target:  []string{directconnect.GatewayAssociationStateAssociated},
		Refresh: statusGatewayAssociationState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.GatewayAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitGatewayAssociationDeleted(conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.GatewayAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayAssociationStateDisassociating},
		Target:  []string{},
		Refresh: statusGatewayAssociationState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.GatewayAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func waitLagDeleted(conn *directconnect.DirectConnect, id string) (*directconnect.Lag, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.LagStateAvailable, directconnect.LagStateRequested, directconnect.LagStatePending, directconnect.LagStateDeleting},
		Target:  []string{},
		Refresh: statusLagState(conn, id),
		Timeout: lagDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.Lag); ok {
		return output, err
	}

	return nil, err
}
