// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

// Exports for use in tests only.
var (
	ResourceComputeEnvironment = resourceComputeEnvironment
	ResourceJobQueue           = newJobQueueResource
	ResourceSchedulingPolicy   = resourceSchedulingPolicy

	ExpandEC2ConfigurationsUpdate           = expandEC2ConfigurationsUpdate
	ExpandLaunchTemplateSpecificationUpdate = expandLaunchTemplateSpecificationUpdate
	FindComputeEnvironmentDetailByName      = findComputeEnvironmentDetailByName
	FindJobQueueByID                        = findJobQueueByID
	FindSchedulingPolicyByARN               = findSchedulingPolicyByARN
)
