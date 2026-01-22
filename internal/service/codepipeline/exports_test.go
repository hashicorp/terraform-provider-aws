// Copyright IBM Corp. 2014, 2026
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
