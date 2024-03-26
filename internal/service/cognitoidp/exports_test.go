// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

// Exports for use in tests only.
var (
	ResourceIdentityProvider        = resourceIdentityProvider
	ResourceManagedUserPoolClient   = newManagedUserPoolClientResource
	ResourceResourceServer          = resourceResourceServer
	ResourceRiskConfiguration       = resourceRiskConfiguration
	ResourceUser                    = resourceUser
	ResourceUserGroup               = resourceUserGroup
	ResourceUserInGroup             = resourceUserInGroup
	ResourceUserPool                = resourceUserPool
	ResourceUserPoolClient          = newUserPoolClientResource
	ResourceUserPoolDomain          = resourceUserPoolDomain
	ResourceUserPoolUICustomization = resourceUserPoolUICustomization

	FindGroupByTwoPartKey                   = findGroupByTwoPartKey
	FindIdentityProviderByTwoPartKey        = findIdentityProviderByTwoPartKey
	FindUserByTwoPartKey                    = findUserByTwoPartKey
	FindUserPoolByID                        = findUserPoolByID
	FindUserPoolUICustomizationByTwoPartKey = findUserPoolUICustomizationByTwoPartKey
)
