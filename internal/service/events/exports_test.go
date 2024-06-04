// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

// Exports for use in tests only.
var (
	ResourceAPIDestination = resourceAPIDestination
	ResourceArchive        = resourceArchive
	ResourceBus            = resourceBus
	ResourceBusPolicy      = resourceBusPolicy
	ResourceConnection     = resourceConnection
	ResourceEndpoint       = resourceEndpoint
	ResourcePermission     = resourcePermission
	ResourceRule           = resourceRule
	ResourceTarget         = resourceTarget

	FindAPIDestinationByName    = findAPIDestinationByName
	FindArchiveByName           = findArchiveByName
	FindConnectionByName        = findConnectionByName
	FindEndpointByName          = findEndpointByName
	FindEventBusByName          = findEventBusByName
	FindEventBusPolicyByName    = findEventBusPolicyByName
	FindPermissionByTwoPartKey  = findPermissionByTwoPartKey
	FindRuleByTwoPartKey        = findRuleByTwoPartKey
	FindTargetByThreePartKey    = findTargetByThreePartKey
	RuleEventPatternJSONDecoder = ruleEventPatternJSONDecoder
	RuleCreateResourceID        = ruleCreateResourceID
	RuleParseResourceID         = ruleParseResourceID
	TargetParseImportID         = targetParseImportID
	TargetStateUpgradeV0        = targetStateUpgradeV0
)
