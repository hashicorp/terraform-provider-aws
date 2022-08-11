package directconnect

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindConnectionByID(conn *directconnect.DirectConnect, id string) (*directconnect.Connection, error) {
	input := &directconnect.DescribeConnectionsInput{
		ConnectionId: aws.String(id),
	}

	output, err := conn.DescribeConnections(input)

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Could not find Connection with ID") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Connections) == 0 || output.Connections[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Connections); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	connection := output.Connections[0]

	if state := aws.StringValue(connection.ConnectionState); state == directconnect.ConnectionStateDeleted || state == directconnect.ConnectionStateRejected {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return connection, nil
}

func FindConnectionAssociationExists(conn *directconnect.DirectConnect, connectionID, lagID string) error {
	connection, err := FindConnectionByID(conn, connectionID)

	if err != nil {
		return err
	}

	if lagID != aws.StringValue(connection.LagId) {
		return &resource.NotFoundError{}
	}

	return nil
}

func FindGatewayByID(conn *directconnect.DirectConnect, id string) (*directconnect.Gateway, error) {
	input := &directconnect.DescribeDirectConnectGatewaysInput{
		DirectConnectGatewayId: aws.String(id),
	}

	output, err := conn.DescribeDirectConnectGateways(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DirectConnectGateways) == 0 || output.DirectConnectGateways[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.DirectConnectGateways); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	gateway := output.DirectConnectGateways[0]

	if state := aws.StringValue(gateway.DirectConnectGatewayState); state == directconnect.GatewayStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return gateway, nil
}

func FindGatewayAssociationByID(conn *directconnect.DirectConnect, id string) (*directconnect.GatewayAssociation, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
		AssociationId: aws.String(id),
	}

	return FindGatewayAssociation(conn, input)
}

func FindGatewayAssociationByGatewayIDAndAssociatedGatewayID(conn *directconnect.DirectConnect, directConnectGatewayID, associatedGatewayID string) (*directconnect.GatewayAssociation, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
		AssociatedGatewayId:    aws.String(associatedGatewayID),
		DirectConnectGatewayId: aws.String(directConnectGatewayID),
	}

	return FindGatewayAssociation(conn, input)
}

func FindGatewayAssociationByGatewayIDAndVirtualGatewayID(conn *directconnect.DirectConnect, directConnectGatewayID, virtualGatewayID string) (*directconnect.GatewayAssociation, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
		DirectConnectGatewayId: aws.String(directConnectGatewayID),
		VirtualGatewayId:       aws.String(virtualGatewayID),
	}

	return FindGatewayAssociation(conn, input)
}

func FindGatewayAssociation(conn *directconnect.DirectConnect, input *directconnect.DescribeDirectConnectGatewayAssociationsInput) (*directconnect.GatewayAssociation, error) {
	output, err := conn.DescribeDirectConnectGatewayAssociations(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DirectConnectGatewayAssociations) == 0 || output.DirectConnectGatewayAssociations[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.DirectConnectGatewayAssociations); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	association := output.DirectConnectGatewayAssociations[0]

	if state := aws.StringValue(association.AssociationState); state == directconnect.GatewayAssociationStateDisassociated {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	if association.AssociatedGateway == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty AssociatedGateway",
			LastRequest: input,
		}
	}

	return association, nil
}

func FindGatewayAssociationProposalByID(conn *directconnect.DirectConnect, id string) (*directconnect.GatewayAssociationProposal, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationProposalsInput{
		ProposalId: aws.String(id),
	}

	output, err := conn.DescribeDirectConnectGatewayAssociationProposals(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DirectConnectGatewayAssociationProposals) == 0 || output.DirectConnectGatewayAssociationProposals[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.DirectConnectGatewayAssociationProposals); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	proposal := output.DirectConnectGatewayAssociationProposals[0]

	if state := aws.StringValue(proposal.ProposalState); state == directconnect.GatewayAssociationProposalStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	if proposal.AssociatedGateway == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty AssociatedGateway",
			LastRequest: input,
		}
	}

	return proposal, nil
}

func FindHostedConnectionByID(conn *directconnect.DirectConnect, id string) (*directconnect.Connection, error) {
	input := &directconnect.DescribeHostedConnectionsInput{
		ConnectionId: aws.String(id),
	}

	output, err := conn.DescribeHostedConnections(input)

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Could not find Connection with ID") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Connections) == 0 || output.Connections[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Connections); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	connection := output.Connections[0]

	if state := aws.StringValue(connection.ConnectionState); state == directconnect.ConnectionStateDeleted || state == directconnect.ConnectionStateRejected {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return connection, nil
}

func FindLagByID(conn *directconnect.DirectConnect, id string) (*directconnect.Lag, error) {
	input := &directconnect.DescribeLagsInput{
		LagId: aws.String(id),
	}

	output, err := conn.DescribeLags(input)

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Could not find Lag with ID") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Lags) == 0 || output.Lags[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Lags); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	lag := output.Lags[0]

	if state := aws.StringValue(lag.LagState); state == directconnect.LagStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return lag, nil
}

func FindLocationByCode(conn *directconnect.DirectConnect, code string) (*directconnect.Location, error) {
	input := &directconnect.DescribeLocationsInput{}

	locations, err := FindLocations(conn, input)

	if err != nil {
		return nil, err
	}

	for _, location := range locations {
		if aws.StringValue(location.LocationCode) == code {
			return location, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func FindLocations(conn *directconnect.DirectConnect, input *directconnect.DescribeLocationsInput) ([]*directconnect.Location, error) {
	output, err := conn.DescribeLocations(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Locations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Locations, nil
}
