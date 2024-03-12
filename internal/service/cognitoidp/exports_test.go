// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

// Exports for use in tests only.
var (
	ResourceIdentityProvider      = resourceIdentityProvider
	ResourceManagedUserPoolClient = newResourceManagedUserPoolClient
	ResourceUserGroup             = resourceUserGroup
	ResourceUserPoolClient        = newResourceUserPoolClient

	FindGroupByTwoPartKey            = findGroupByTwoPartKey
	FindIdentityProviderByTwoPartKey = findIdentityProviderByTwoPartKey
)
