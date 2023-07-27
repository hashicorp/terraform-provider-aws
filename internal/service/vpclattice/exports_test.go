// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

// Exports for use in tests only.
var (
	FindAccessLogSubscriptionByID            = findAccessLogSubscriptionByID
	FindServiceNetworkServiceAssociationByID = findServiceNetworkServiceAssociationByID
	FindServiceNetworkVPCAssociationByID     = findServiceNetworkVPCAssociationByID
	FindTargetByThreePartKey                 = findTargetByThreePartKey

	IDFromIDOrARN             = idFromIDOrARN
	SuppressEquivalentIDOrARN = suppressEquivalentIDOrARN

	ResourceAccessLogSubscription            = resourceAccessLogSubscription
	ResourceServiceNetworkServiceAssociation = resourceServiceNetworkServiceAssociation
	ResourceServiceNetworkVPCAssociation     = resourceServiceNetworkVPCAssociation
	ResourceTargetGroupAttachment            = resourceTargetGroupAttachment
)
