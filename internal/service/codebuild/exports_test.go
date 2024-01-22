// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

// Exports for use in tests only.
var (
	ResourceResourcePolicy   = resourceResourcePolicy
	ResourceSourceCredential = resourceSourceCredential

	FindResourcePolicyByARN    = findResourcePolicyByARN
	FindSourceCredentialsByARN = findSourceCredentialsByARN
)
