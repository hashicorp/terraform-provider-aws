// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

// Exports for use in tests only.
var (
	ResourceAccessLogSubscription             = resourceAccessLogSubscription
	ResourceListener                          = resourceListener
	ResourceResourceConfiguration             = newResourceConfigurationResource
	ResourceResourceGateway                   = newResourceGatewayResource
	ResourceService                           = resourceService
	ResourceServiceNetwork                    = resourceServiceNetwork
	ResourceServiceNetworkResourceAssociation = newServiceNetworkResourceAssociationResource
	ResourceServiceNetworkServiceAssociation  = resourceServiceNetworkServiceAssociation
	ResourceServiceNetworkVPCAssociation      = resourceServiceNetworkVPCAssociation
	ResourceTargetGroupAttachment             = resourceTargetGroupAttachment

	FindAccessLogSubscriptionByID             = findAccessLogSubscriptionByID
	FindListenerByTwoPartKey                  = findListenerByTwoPartKey
	FindResourceConfigurationByID             = findResourceConfigurationByID
	FindResourceGatewayByID                   = findResourceGatewayByID
	FindServiceByID                           = findServiceByID
	FindServiceNetworkByID                    = findServiceNetworkByID
	FindServiceNetworkResourceAssociationByID = findServiceNetworkResourceAssociationByID
	FindServiceNetworkServiceAssociationByID  = findServiceNetworkServiceAssociationByID
	FindServiceNetworkVPCAssociationByID      = findServiceNetworkVPCAssociationByID
	FindTargetByThreePartKey                  = findTargetByThreePartKey

	IDFromIDOrARN                               = idFromIDOrARN
	SuppressEquivalentCloudWatchLogsLogGroupARN = suppressEquivalentCloudWatchLogsLogGroupARN
	SuppressEquivalentIDOrARN                   = suppressEquivalentIDOrARN
)
