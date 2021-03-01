package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func GatewayByID(conn *directconnect.DirectConnect, id string) (*directconnect.Gateway, error) {
	input := &directconnect.DescribeDirectConnectGatewaysInput{
		DirectConnectGatewayId: aws.String(id),
	}

	output, err := conn.DescribeDirectConnectGateways(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DirectConnectGateways) == 0 || output.DirectConnectGateways[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	// TODO Check for multiple results.
	// TODO https://github.com/hashicorp/terraform-provider-aws/pull/17613.

	gateway := output.DirectConnectGateways[0]

	if state := aws.StringValue(gateway.DirectConnectGatewayState); state == directconnect.GatewayStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return gateway, nil
}

func GatewayAssociationByID(conn *directconnect.DirectConnect, id string) (*directconnect.GatewayAssociation, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
		AssociationId: aws.String(id),
	}

	return GatewayAssociation(conn, input)
}

func GatewayAssociationByDirectConnectGatewayIDAndAssociatedGatewayID(conn *directconnect.DirectConnect, directConnectGatewayID, associatedGatewayID string) (*directconnect.GatewayAssociation, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
		AssociatedGatewayId:    aws.String(associatedGatewayID),
		DirectConnectGatewayId: aws.String(directConnectGatewayID),
	}

	return GatewayAssociation(conn, input)
}

func GatewayAssociationByDirectConnectGatewayIDAndVirtualGatewayID(conn *directconnect.DirectConnect, directConnectGatewayID, virtualGatewayID string) (*directconnect.GatewayAssociation, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
		DirectConnectGatewayId: aws.String(directConnectGatewayID),
		VirtualGatewayId:       aws.String(virtualGatewayID),
	}

	return GatewayAssociation(conn, input)
}

func GatewayAssociation(conn *directconnect.DirectConnect, input *directconnect.DescribeDirectConnectGatewayAssociationsInput) (*directconnect.GatewayAssociation, error) {
	output, err := conn.DescribeDirectConnectGatewayAssociations(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DirectConnectGatewayAssociations) == 0 || output.DirectConnectGatewayAssociations[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	// TODO Check for multiple results.
	// TODO https://github.com/hashicorp/terraform-provider-aws/pull/17613.

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

func GatewayAssociationProposalByID(conn *directconnect.DirectConnect, id string) (*directconnect.GatewayAssociationProposal, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationProposalsInput{
		ProposalId: aws.String(id),
	}

	output, err := conn.DescribeDirectConnectGatewayAssociationProposals(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DirectConnectGatewayAssociationProposals) == 0 || output.DirectConnectGatewayAssociationProposals[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	// TODO Check for multiple results.
	// TODO https://github.com/hashicorp/terraform-provider-aws/pull/17613.

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

// ConnectionByID returns the connections corresponding to the specified connection ID.
// Returns NotFoundError if no connection is found.
func ConnectionByID(conn *directconnect.DirectConnect, connectionID string) (*directconnect.Connection, error) {
	input := &directconnect.DescribeConnectionsInput{
		ConnectionId: aws.String(connectionID),
	}

	connections, err := Connections(conn, input)

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Could not find Connection with ID") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	// Handle any empty result.
	if len(connections) == 0 {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	if state := aws.StringValue(connections[0].ConnectionState); state == directconnect.ConnectionStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return connections[0], nil
}

// Connections returns the connections corresponding to the specified input.
// Returns an empty slice if no connections are found.
func Connections(conn *directconnect.DirectConnect, input *directconnect.DescribeConnectionsInput) ([]*directconnect.Connection, error) {
	output, err := conn.DescribeConnections(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return []*directconnect.Connection{}, nil
	}

	return output.Connections, nil
}

// LagByID returns the locations corresponding to the specified LAG ID.
// Returns NotFoundError if no LAG is found.
func LagByID(conn *directconnect.DirectConnect, lagID string) (*directconnect.Lag, error) {
	input := &directconnect.DescribeLagsInput{
		LagId: aws.String(lagID),
	}

	lags, err := Lags(conn, input)

	if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "Could not find Lag with ID") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	// Handle any empty result.
	if len(lags) == 0 {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	if state := aws.StringValue(lags[0].LagState); state == directconnect.LagStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return lags[0], nil
}

// Lags returns the LAGs corresponding to the specified input.
// Returns an empty slice if no LAGs are found.
func Lags(conn *directconnect.DirectConnect, input *directconnect.DescribeLagsInput) ([]*directconnect.Lag, error) {
	output, err := conn.DescribeLags(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return []*directconnect.Lag{}, nil
	}

	return output.Lags, nil
}

// LocationByCode returns the locations corresponding to the specified location code.
// Returns NotFoundError if no location is found.
func LocationByCode(conn *directconnect.DirectConnect, locationCode string) (*directconnect.Location, error) {
	input := &directconnect.DescribeLocationsInput{}

	locations, err := Locations(conn, input)

	if err != nil {
		return nil, err
	}

	for _, location := range locations {
		if aws.StringValue(location.LocationCode) == locationCode {
			return location, nil
		}
	}

	return nil, &resource.NotFoundError{
		Message:     "Empty result",
		LastRequest: input,
	}
}

// Locations returns the locations corresponding to the specified input.
// Returns an empty slice if no locations are found.
func Locations(conn *directconnect.DirectConnect, input *directconnect.DescribeLocationsInput) ([]*directconnect.Location, error) {
	output, err := conn.DescribeLocations(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return []*directconnect.Location{}, nil
	}

	return output.Locations, nil
}
