// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

// Exports for use in tests only.
var (
	ResourceIdentityProvider         = resourceIdentityProvider
	ResourceLogDeliveryConfiguration = newLogDeliveryConfigurationResource
	ResourceManagedLoginBranding     = newManagedLoginBrandingResource
	ResourceManagedUserPoolClient    = newManagedUserPoolClientResource
	ResourceResourceServer           = resourceResourceServer
	ResourceRiskConfiguration        = resourceRiskConfiguration
	ResourceUser                     = resourceUser
	ResourceUserGroup                = resourceUserGroup
	ResourceUserInGroup              = resourceUserInGroup
	ResourceUserPool                 = resourceUserPool
	ResourceUserPoolClient           = newUserPoolClientResource
	ResourceUserPoolDomain           = resourceUserPoolDomain
	ResourceUserPoolUICustomization  = resourceUserPoolUICustomization

	FindGroupByTwoPartKey                    = findGroupByTwoPartKey
	FindGroupUserByThreePartKey              = findGroupUserByThreePartKey
	FindIdentityProviderByTwoPartKey         = findIdentityProviderByTwoPartKey
	FindLogDeliveryConfigurationByUserPoolID = findLogDeliveryConfigurationByUserPoolID
	FindManagedLoginBrandingByThreePartKey   = findManagedLoginBrandingByThreePartKey
	FindResourceServerByTwoPartKey           = findResourceServerByTwoPartKey
	FindRiskConfigurationByTwoPartKey        = findRiskConfigurationByTwoPartKey
	FindUserByTwoPartKey                     = findUserByTwoPartKey
	FindUserPoolByID                         = findUserPoolByID
	FindUserPoolClientByName                 = findUserPoolClientByName
	FindUserPoolClientByTwoPartKey           = findUserPoolClientByTwoPartKey
	FindUserPoolDomain                       = findUserPoolDomain
	FindUserPoolUICustomizationByTwoPartKey  = findUserPoolUICustomizationByTwoPartKey
)
