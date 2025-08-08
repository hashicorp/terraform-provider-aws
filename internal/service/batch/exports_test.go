// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

// Exports for use in tests only.
var (
	ResourceComputeEnvironment = resourceComputeEnvironment
	ResourceJobDefinition      = resourceJobDefinition
	ResourceJobQueue           = newJobQueueResource
	ResourceSchedulingPolicy   = resourceSchedulingPolicy

	EquivalentContainerPropertiesJSON       = equivalentContainerPropertiesJSON
	EquivalentECSPropertiesJSON             = equivalentECSPropertiesJSON
	EquivalentEKSPropertiesJSON             = equivalentEKSPropertiesJSON
	EquivalentNodePropertiesJSON            = equivalentNodePropertiesJSON
	ExpandEC2ConfigurationsUpdate           = expandEC2ConfigurationsUpdate
	ExpandLaunchTemplateSpecificationUpdate = expandLaunchTemplateSpecificationUpdate
	FindComputeEnvironmentDetailByName      = findComputeEnvironmentDetailByName
	FindJobDefinitionByARN                  = findJobDefinitionByARN
	FindJobQueueByID                        = findJobQueueByID
	FindSchedulingPolicyByARN               = findSchedulingPolicyByARN

	ListTags = listTags

	ComputeEnvironmentStateUpgradeV0 = computeEnvironmentStateUpgradeV0
)
