// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

// Exports for use in tests only.
var (
	ResourceAccessLogSubscription            = resourceAccessLogSubscription
	ResourceListener                         = resourceListener
	ResourceResourceGateway                  = newResourceGatewayResource
	ResourceService                          = resourceService
	ResourceServiceNetwork                   = resourceServiceNetwork
	ResourceServiceNetworkServiceAssociation = resourceServiceNetworkServiceAssociation
	ResourceServiceNetworkVPCAssociation     = resourceServiceNetworkVPCAssociation
	ResourceTargetGroupAttachment            = resourceTargetGroupAttachment

	FindAccessLogSubscriptionByID            = findAccessLogSubscriptionByID
	FindListenerByTwoPartKey                 = findListenerByTwoPartKey
	FindResourceGatewayByID                  = findResourceGatewayByID
	FindServiceByID                          = findServiceByID
	FindServiceNetworkByID                   = findServiceNetworkByID
	FindServiceNetworkServiceAssociationByID = findServiceNetworkServiceAssociationByID
	FindServiceNetworkVPCAssociationByID     = findServiceNetworkVPCAssociationByID
	FindTargetByThreePartKey                 = findTargetByThreePartKey

	IDFromIDOrARN                               = idFromIDOrARN
	SuppressEquivalentCloudWatchLogsLogGroupARN = suppressEquivalentCloudWatchLogsLogGroupARN
	SuppressEquivalentIDOrARN                   = suppressEquivalentIDOrARN
)
