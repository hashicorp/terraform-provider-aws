// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

// Exports for use in tests only.
var (
	ResourceJobQueue = newResourceJobQueue

	ExpandEC2ConfigurationsUpdate           = expandEC2ConfigurationsUpdate
	ExpandLaunchTemplateSpecificationUpdate = expandLaunchTemplateSpecificationUpdate
	FindComputeEnvironmentDetailByName      = findComputeEnvironmentDetailByName
	FindJobQueueByName                      = findJobQueueByName
)
