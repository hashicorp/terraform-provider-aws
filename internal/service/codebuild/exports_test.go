// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

// Exports for use in tests only.
var (
	ResourceProject          = resourceProject
	ResourceReportGroup      = resourceReportGroup
	ResourceResourcePolicy   = resourceResourcePolicy
	ResourceSourceCredential = resourceSourceCredential
	ResourceWebhook          = resourceWebhook

	FindProjectByNameOrARN     = findProjectByNameOrARN
	FindReportGroupByARN       = findReportGroupByARN
	FindResourcePolicyByARN    = findResourcePolicyByARN
	FindSourceCredentialsByARN = findSourceCredentialsByARN
	FindSourceCredentials      = findSourceCredentials
	FindWebhookByProjectName   = findWebhookByProjectName
)
