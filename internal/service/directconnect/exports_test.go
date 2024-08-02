// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

// Exports for use in tests only.
var (
	ResourceBGPPeer    = resourceBGPPeer
	ResourceConnection = resourceConnection

	FindBGPPeerByThreePartKey = findBGPPeerByThreePartKey
	FindConnectionByID        = findConnectionByID
	FindVirtualInterfaceByID  = findVirtualInterfaceByID
	ValidConnectionBandWidth  = validConnectionBandWidth
)
