// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindConnectionByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Connection, error) {
	input := &directconnect.DescribeConnectionsInput{
		ConnectionId: aws.String(id),
	}

	output, err := conn.DescribeConnections(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Could not find Connection with ID") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	connection, err := tfresource.AssertSingleValueResult(output.Connections)

	if err != nil {
		return nil, err
	}

	if connection.ConnectionState == awstypes.ConnectionStateDeleted || connection.ConnectionState == awstypes.ConnectionStateRejected {
		return nil, &retry.NotFoundError{
			Message:     string(connection.ConnectionState),
			LastRequest: input,
		}
	}

	return connection, nil
}

func FindConnectionAssociationExists(ctx context.Context, conn *directconnect.Client, connectionID, lagID string) error {
	connection, err := FindConnectionByID(ctx, conn, connectionID)

	if err != nil {
		return err
	}

	if lagID != aws.ToString(connection.LagId) {
		return &retry.NotFoundError{}
	}

	return nil
}

func FindGatewayByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.DirectConnectGateway, error) {
	input := &directconnect.DescribeDirectConnectGatewaysInput{
		DirectConnectGatewayId: aws.String(id),
	}

	output, err := conn.DescribeDirectConnectGateways(ctx, input)

	if err != nil {
		return nil, err
	}

	gateway, err := tfresource.AssertSingleValueResult(output.DirectConnectGateways)

	if err != nil {
		return nil, err
	}

	if gateway.DirectConnectGatewayState == awstypes.DirectConnectGatewayStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(gateway.DirectConnectGatewayState),
			LastRequest: input,
		}
	}

	return gateway, nil
}

func FindGatewayAssociationByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.DirectConnectGatewayAssociation, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
		AssociationId: aws.String(id),
	}

	return FindGatewayAssociation(ctx, conn, input)
}

func FindGatewayAssociationByGatewayIDAndAssociatedGatewayID(ctx context.Context, conn *directconnect.Client, directConnectGatewayID, associatedGatewayID string) (*awstypes.DirectConnectGatewayAssociation, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
		AssociatedGatewayId:    aws.String(associatedGatewayID),
		DirectConnectGatewayId: aws.String(directConnectGatewayID),
	}

	return FindGatewayAssociation(ctx, conn, input)
}

func FindGatewayAssociationByGatewayIDAndVirtualGatewayID(ctx context.Context, conn *directconnect.Client, directConnectGatewayID, virtualGatewayID string) (*awstypes.DirectConnectGatewayAssociation, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
		DirectConnectGatewayId: aws.String(directConnectGatewayID),
		VirtualGatewayId:       aws.String(virtualGatewayID),
	}

	return FindGatewayAssociation(ctx, conn, input)
}

func FindGatewayAssociation(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeDirectConnectGatewayAssociationsInput) (*awstypes.DirectConnectGatewayAssociation, error) {
	output, err := conn.DescribeDirectConnectGatewayAssociations(ctx, input)

	if err != nil {
		return nil, err
	}

	association, err := tfresource.AssertSingleValueResult(output.DirectConnectGatewayAssociations)

	if err != nil {
		return nil, err
	}

	if association.AssociationState == awstypes.DirectConnectGatewayAssociationStateDisassociated {
		return nil, &retry.NotFoundError{
			Message:     string(association.AssociationState),
			LastRequest: input,
		}
	}

	if association.AssociatedGateway == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty AssociatedGateway",
			LastRequest: input,
		}
	}

	return association, nil
}

func FindGatewayAssociationProposalByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.DirectConnectGatewayAssociationProposal, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationProposalsInput{
		ProposalId: aws.String(id),
	}

	output, err := conn.DescribeDirectConnectGatewayAssociationProposals(ctx, input)

	if err != nil {
		return nil, err
	}

	proposal, err := tfresource.AssertSingleValueResult(output.DirectConnectGatewayAssociationProposals)

	if err != nil {
		return nil, err
	}

	if proposal.ProposalState == awstypes.DirectConnectGatewayAssociationProposalStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(proposal.ProposalState),
			LastRequest: input,
		}
	}

	if proposal.AssociatedGateway == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty AssociatedGateway",
			LastRequest: input,
		}
	}

	return proposal, nil
}

func FindHostedConnectionByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Connection, error) {
	input := &directconnect.DescribeHostedConnectionsInput{
		ConnectionId: aws.String(id),
	}

	output, err := conn.DescribeHostedConnections(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Could not find Connection with ID") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	connection, err := tfresource.AssertSingleValueResult(output.Connections)

	if err != nil {
		return nil, err
	}

	if connection.ConnectionState == awstypes.ConnectionStateDeleted || connection.ConnectionState == awstypes.ConnectionStateRejected {
		return nil, &retry.NotFoundError{
			Message:     string(connection.ConnectionState),
			LastRequest: input,
		}
	}

	return connection, nil
}

func FindLagByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Lag, error) {
	input := &directconnect.DescribeLagsInput{
		LagId: aws.String(id),
	}

	output, err := conn.DescribeLags(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Could not find Lag with ID") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	lag, err := tfresource.AssertSingleValueResult(output.Lags)

	if err != nil {
		return nil, err
	}

	if lag.LagState == awstypes.LagStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(lag.LagState),
			LastRequest: input,
		}
	}

	return lag, nil
}

func FindLocationByCode(ctx context.Context, conn *directconnect.Client, code string) (awstypes.Location, error) {
	input := &directconnect.DescribeLocationsInput{}

	locations, err := FindLocations(ctx, conn, input)

	if err != nil {
		return awstypes.Location{}, err
	}

	for _, location := range locations {
		if aws.ToString(location.LocationCode) == code {
			return location, nil
		}
	}

	return awstypes.Location{}, tfresource.NewEmptyResultError(input)
}

func FindLocations(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeLocationsInput) ([]awstypes.Location, error) {
	output, err := conn.DescribeLocations(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Locations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Locations, nil
}
