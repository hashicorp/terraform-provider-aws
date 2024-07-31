// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline

// Exports for use in tests only.
var (
	ResourceCustomActionType = resourceCustomActionType
	ResourcePipeline         = resourcePipeline
	ResourceWebhook          = resourceWebhook

	FindCustomActionTypeByThreePartKey = findCustomActionTypeByThreePartKey
	FindPipelineByName                 = findPipelineByName
	FindWebhookByARN                   = findWebhookByARN
)
