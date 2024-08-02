// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

// Exports for use in tests only.
var (
	ResourceBGPPeer               = resourceBGPPeer
	ResourceConnection            = resourceConnection
	ResourceConnectionAssociation = resourceConnectionAssociation

	FindBGPPeerByThreePartKey    = findBGPPeerByThreePartKey
	FindConnectionByID           = findConnectionByID
	FindConnectionLAGAssociation = findConnectionLAGAssociation
	FindVirtualInterfaceByID     = findVirtualInterfaceByID
	ValidConnectionBandWidth     = validConnectionBandWidth
)
