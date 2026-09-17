// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
)

// Exports for use in tests only.
var (
	ResourceBGPPeer                               = resourceBGPPeer
	ResourceConnection                            = resourceConnection
	ResourceConnectionAssociation                 = resourceConnectionAssociation
	ResourceConnectionConfirmation                = resourceConnectionConfirmation
	ResourceGateway                               = resourceGateway
	ResourceGatewayAssociation                    = resourceGatewayAssociation
	ResourceGatewayAssociationProposal            = resourceGatewayAssociationProposal
	ResourceHostedConnection                      = resourceHostedConnection
	ResourceHostedPrivateVirtualInterface         = resourceHostedPrivateVirtualInterface
	ResourceHostedPrivateVirtualInterfaceAccepter = resourceHostedPrivateVirtualInterfaceAccepter
	ResourceHostedPublicVirtualInterface          = resourceHostedPublicVirtualInterface
	ResourceHostedPublicVirtualInterfaceAccepter  = resourceHostedPublicVirtualInterfaceAccepter
	ResourceHostedTransitVirtualInterface         = resourceHostedTransitVirtualInterface
	ResourceHostedTransitVirtualInterfaceAccepter = resourceHostedTransitVirtualInterfaceAccepter
	ResourceLag                                   = resourceLag
	ResourceMacSecKeyAssociation                  = resourceMacSecKeyAssociation
	ResourcePrivateVirtualInterface               = resourcePrivateVirtualInterface
	ResourcePublicVirtualInterface                = resourcePublicVirtualInterface
	ResourceTransitVirtualInterface               = resourceTransitVirtualInterface

	FindBGPPeerByThreePartKey = func(ctx context.Context, conn *directconnect.Client, vifID string, addrFamily awstypes.AddressFamily, asn int32) (*awstypes.BGPPeer, error) {
		return findBGPPeerByThreePartKey(ctx, conn, vifID, addrFamily, int64(asn))
	}
	FindBGPPeerByThreePartKeyWithInt64ASN = findBGPPeerByThreePartKey
	FindConnectionByID                    = findConnectionByID
	FindConnectionLAGAssociation          = findConnectionLAGAssociation
	FindGatewayAssociationByID            = findGatewayAssociationByID
	FindGatewayAssociationProposalByID    = findGatewayAssociationProposalByID
	FindGatewayByID                       = findGatewayByID
	FindHostedConnectionByID              = findHostedConnectionByID
	FindLagByID                           = findLagByID
	FindMacSecKeyByTwoPartKey             = findMacSecKeyByTwoPartKey
	FindVirtualInterfaceByID              = findVirtualInterfaceByID
	GatewayAssociationStateUpgradeV0      = gatewayAssociationStateUpgradeV0
	GatewayAssociationStateUpgradeV1      = gatewayAssociationStateUpgradeV1
)
