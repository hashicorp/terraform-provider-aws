// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

// Exports for use in tests only.
var (
	ResourceSecret = resourceSecret

	FindSecretByID = findSecretByID

	ErrCodeResourceNotFoundException = errCodeResourceNotFoundException
	ErrCodeInvalidRequestException   = errCodeInvalidRequestException
)
