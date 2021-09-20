package waiter

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

const (
	ConnectionDeletedTimeout       = 10 * time.Minute
	ConnectionDisassociatedTimeout = 1 * time.Minute
	LagDeletedTimeout              = 10 * time.Minute
)

func ConnectionDeleted(conn *directconnect.DirectConnect, id string) (*directconnect.Connection, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.ConnectionStatePending, directconnect.ConnectionStateOrdering, directconnect.ConnectionStateAvailable, directconnect.ConnectionStateRequested, directconnect.ConnectionStateDeleting},
		Target:  []string{},
		Refresh: ConnectionState(conn, id),
		Timeout: ConnectionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.Connection); ok {
		return output, err
	}

	return nil, err
}

func GatewayCreated(conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.Gateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayStatePending},
		Target:  []string{directconnect.GatewayStateAvailable},
		Refresh: GatewayState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.Gateway); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func GatewayDeleted(conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.Gateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayStatePending, directconnect.GatewayStateAvailable, directconnect.GatewayStateDeleting},
		Target:  []string{},
		Refresh: GatewayState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.Gateway); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func GatewayAssociationCreated(conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.GatewayAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayAssociationStateAssociating},
		Target:  []string{directconnect.GatewayAssociationStateAssociated},
		Refresh: GatewayAssociationState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.GatewayAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func GatewayAssociationUpdated(conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.GatewayAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayAssociationStateUpdating},
		Target:  []string{directconnect.GatewayAssociationStateAssociated},
		Refresh: GatewayAssociationState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.GatewayAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func GatewayAssociationDeleted(conn *directconnect.DirectConnect, id string, timeout time.Duration) (*directconnect.GatewayAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.GatewayAssociationStateDisassociating},
		Target:  []string{},
		Refresh: GatewayAssociationState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.GatewayAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateChangeError)))

		return output, err
	}

	return nil, err
}

func LagDeleted(conn *directconnect.DirectConnect, id string) (*directconnect.Lag, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directconnect.LagStateAvailable, directconnect.LagStateRequested, directconnect.LagStatePending, directconnect.LagStateDeleting},
		Target:  []string{},
		Refresh: LagState(conn, id),
		Timeout: LagDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directconnect.Lag); ok {
		return output, err
	}

	return nil, err
}
