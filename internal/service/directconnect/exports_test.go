// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

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

	FindBGPPeerByThreePartKey          = findBGPPeerByThreePartKey
	FindConnectionByID                 = findConnectionByID
	FindConnectionLAGAssociation       = findConnectionLAGAssociation
	FindGatewayAssociationByID         = findGatewayAssociationByID
	FindGatewayAssociationProposalByID = findGatewayAssociationProposalByID
	FindGatewayByID                    = findGatewayByID
	FindHostedConnectionByID           = findHostedConnectionByID
	FindLagByID                        = findLagByID
	FindMacSecKeyByTwoPartKey          = findMacSecKeyByTwoPartKey
	FindVirtualInterfaceByID           = findVirtualInterfaceByID
	GatewayAssociationStateUpgradeV0   = gatewayAssociationStateUpgradeV0
)
