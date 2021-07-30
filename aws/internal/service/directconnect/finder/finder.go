package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
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
