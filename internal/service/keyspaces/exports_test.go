// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package keyspaces

// Exports for use in tests only.
var (
	ResourceKeyspace = resourceKeyspace
	ResourceTable    = resourceTable

	FindKeyspaceByName                       = findKeyspaceByName
	FindTableByTwoPartKey                    = findTableByTwoPartKey
	FindTableAutoScalingSettingsByTwoPartKey = findTableAutoScalingSettingsByTwoPartKey

	TableParseResourceID = tableParseResourceID

	ExpandAutoScalingSpecification                 = expandAutoScalingSpecification
	ExpandAutoScalingSpecificationDisabled         = expandAutoScalingSpecificationDisabled
	ExpandAutoScalingSettings                      = expandAutoScalingSettings
	ExpandTargetTrackingScalingPolicyConfiguration = expandTargetTrackingScalingPolicyConfiguration

	FlattenAutoScalingSpecification                 = flattenAutoScalingSpecification
	FlattenAutoScalingSettings                      = flattenAutoScalingSettings
	FlattenTargetTrackingScalingPolicyConfiguration = flattenTargetTrackingScalingPolicyConfiguration
)
