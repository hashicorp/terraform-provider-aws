// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

// Exports for use in tests only.
var (
	ResourceConfigurationSet     = newConfigurationSetResource
	ResourceEventDestination     = newEventDestinationResource
	ResourceKeyword              = newKeywordResource
	ResourceOptOutList           = newOptOutListResource
	ResourcePhoneNumber          = newPhoneNumberResource
	ResourcePool                 = newPoolResource
	ResourceProtectConfiguration = newProtectConfigurationResource
	ResourceResourcePolicy       = newResourcePolicyResource
	ResourceSenderID             = newSenderIDResource

	FindConfigurationSetByID             = findConfigurationSetByID
	FindEventDestinationByTwoPartKey     = findEventDestinationByTwoPartKey
	FindKeywordByTwoPartKey              = findKeywordByTwoPartKey
	FindOptOutListByID                   = findOptOutListByID
	FindPhoneNumberByID                  = findPhoneNumberByID
	FindPoolByID                         = findPoolByID
	FindProtectConfigurationByID         = findProtectConfigurationByID
	FindProtectConfigurationCountryRules = findProtectConfigurationCountryRules
	FindResourcePolicyByARN              = findResourcePolicyByARN
	FindSenderIDByTwoPartKey             = findSenderIDByTwoPartKey

	ValidatePhoneIdentity  = validatePhoneIdentity
	ValidateSenderIdentity = validateSenderIdentity
)

type IntendedIdentityConfig = intendedIdentityConfig
