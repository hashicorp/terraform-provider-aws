// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

// Exports for use in tests only.
var (
	ResourceEventSubscription       = resourceEventSubscription
	ResourceProxy                   = resourceProxy
	ResourceProxyDefaultTargetGroup = resourceProxyDefaultTargetGroup
	ResourceProxyEndpoint           = resourceProxyEndpoint
	ResourceProxyTarget             = resourceProxyTarget
	ResourceSubnetGroup             = resourceSubnetGroup

	FindDBInstanceByID                         = findDBInstanceByIDSDKv1
	FindDBProxyByName                          = findDBProxyByName
	FindDBProxyEndpointByTwoPartKey            = findDBProxyEndpointByTwoPartKey
	FindDBProxyTargetByFourPartKey             = findDBProxyTargetByFourPartKey
	FindDBProxyAuthItemByArn                   = findDBProxyAuthItemByArn
	FindDBSubnetGroupByName                    = findDBSubnetGroupByName
	FindDefaultDBProxyTargetGroupByDBProxyName = findDefaultDBProxyTargetGroupByDBProxyName
	FindEventSubscriptionByID                  = findEventSubscriptionByID
	ListTags                                   = listTags
	NewBlueGreenOrchestrator                   = newBlueGreenOrchestrator
	ParseDBInstanceARN                         = parseDBInstanceARN
	ProxyTargetParseResourceID                 = proxyTargetParseResourceID
	WaitBlueGreenDeploymentDeleted             = waitBlueGreenDeploymentDeleted
	WaitBlueGreenDeploymentAvailable           = waitBlueGreenDeploymentAvailable
	WaitDBInstanceAvailable                    = waitDBInstanceAvailableSDKv2
	WaitDBInstanceDeleted                      = waitDBInstanceDeleted

	ErrCodeInvalidAction               = errCodeInvalidAction
	ErrCodeInvalidParameterCombination = errCodeInvalidParameterCombination
	ErrCodeInvalidParameterValue       = errCodeInvalidParameterValue
)
