// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

// Exports for use in tests only.
var (
	FindAccessLogSubscriptionByID            = findAccessLogSubscriptionByID
	FindListenerByTwoPartKey                 = findListenerByTwoPartKey
	FindServiceByID                          = findServiceByID
	FindServiceNetworkByID                   = findServiceNetworkByID
	FindServiceNetworkServiceAssociationByID = findServiceNetworkServiceAssociationByID
	FindServiceNetworkVPCAssociationByID     = findServiceNetworkVPCAssociationByID
	FindTargetByThreePartKey                 = findTargetByThreePartKey

	IDFromIDOrARN                               = idFromIDOrARN
	SuppressEquivalentCloudWatchLogsLogGroupARN = suppressEquivalentCloudWatchLogsLogGroupARN
	SuppressEquivalentIDOrARN                   = suppressEquivalentIDOrARN

	ResourceAccessLogSubscription            = resourceAccessLogSubscription
	ResourceListener                         = resourceListener
	ResourceService                          = resourceService
	ResourceServiceNetwork                   = resourceServiceNetwork
	ResourceServiceNetworkServiceAssociation = resourceServiceNetworkServiceAssociation
	ResourceServiceNetworkVPCAssociation     = resourceServiceNetworkVPCAssociation
	ResourceTargetGroupAttachment            = resourceTargetGroupAttachment
)
