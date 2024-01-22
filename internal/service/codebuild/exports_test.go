// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

// Exports for use in tests only.
var (
	ResourceReportGroup      = resourceReportGroup
	ResourceResourcePolicy   = resourceResourcePolicy
	ResourceSourceCredential = resourceSourceCredential
	ResourceWebhook          = resourceWebhook

	FindReportGroupByARN       = findReportGroupByARN
	FindResourcePolicyByARN    = findResourcePolicyByARN
	FindSourceCredentialsByARN = findSourceCredentialsByARN
	FindWebhookByProjectName   = findWebhookByProjectName
)
