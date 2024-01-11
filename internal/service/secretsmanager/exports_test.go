// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

// Exports for use in tests only.
var (
	ResourceSecret       = resourceSecret
	ResourceSecretPolicy = resourceSecretPolicy

	FindSecretByID       = findSecretByID
	FindSecretPolicyByID = findSecretPolicyByID

	ErrCodeResourceNotFoundException = errCodeResourceNotFoundException
	ErrCodeInvalidRequestException   = errCodeInvalidRequestException
)
