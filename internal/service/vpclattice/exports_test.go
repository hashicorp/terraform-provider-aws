// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

// Exports for use in tests only.
var (
	FindServiceNetworkServiceAssociationByID = findServiceNetworkServiceAssociationByID
	FindServiceNetworkVPCAssociationByID     = findServiceNetworkVPCAssociationByID
	FindTargetByThreePartKey                 = findTargetByThreePartKey

	IDFromIDOrARN             = idFromIDOrARN
	SuppressEquivalentIDOrARN = suppressEquivalentIDOrARN

	ResourceServiceNetworkServiceAssociation = resourceServiceNetworkServiceAssociation
	ResourceServiceNetworkVPCAssociation     = resourceServiceNetworkVPCAssociation
	ResourceTargetGroupAttachment            = resourceTargetGroupAttachment
)
