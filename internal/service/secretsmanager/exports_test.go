// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

// Exports for use in tests only.
var (
	ResourceSecret         = resourceSecret
	ResourceSecretPolicy   = resourceSecretPolicy
	ResourceSecretRotation = resourceSecretRotation
	ResourceSecretVersion  = resourceSecretVersion
	ResourceTag            = resourceTag

	FindSecretByID                     = findSecretByID
	FindSecretPolicyByID               = findSecretPolicyByID
	FindSecretVersionByTwoPartKey      = findSecretVersionByTwoPartKey
	FindSecretVersionEntryByTwoPartKey = findSecretVersionEntryByTwoPartKey
	FindSecretTag                      = findSecretTag
)
