// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

// Exports for use in tests only.
var (
	ResourceAccountDefaultProtectConfiguration = newAccountDefaultProtectConfigurationResource
	ResourceConfigurationSet                   = newConfigurationSetResource
	ResourceEventDestination                   = newEventDestinationResource
	ResourceKeyword                            = newKeywordResource
	ResourceOptOutList                         = newOptOutListResource
	ResourcePhoneNumber                        = newPhoneNumberResource
	ResourcePool                               = newPoolResource
	ResourceProtectConfiguration               = newProtectConfigurationResource
	ResourceProtectConfigurationAssociation    = newProtectConfigurationAssociationResource
	ResourceResourcePolicy                     = newResourcePolicyResource
	ResourceSenderID                           = newSenderIDResource

	FindAccountDefaultProtectConfiguration                    = findAccountDefaultProtectConfiguration
	FindConfigurationSetByID                                  = findConfigurationSetByID
	FindEventDestinationByTwoPartKey                          = findEventDestinationByTwoPartKey
	FindKeywordByTwoPartKey                                   = findKeywordByTwoPartKey
	FindOptOutListByID                                        = findOptOutListByID
	FindPhoneNumberByID                                       = findPhoneNumberByID
	FindPoolByID                                              = findPoolByID
	FindProtectConfigurationAssociationByConfigurationSetName = findProtectConfigurationAssociationByConfigurationSetName
	FindProtectConfigurationByID                              = findProtectConfigurationByID
	FindProtectConfigurationCountryRuleSet                    = findProtectConfigurationCountryRuleSet
	FindResourcePolicyByARN                                   = findResourcePolicyByARN
	FindSenderIDByTwoPartKey                                  = findSenderIDByTwoPartKey

	ProtectConfigurationCountryRuleSetUpdates = protectConfigurationCountryRuleSetUpdates
	ValidatePhoneIdentity                     = validatePhoneIdentity
	ValidateSenderIdentity                    = validateSenderIdentity
)

type IntendedIdentityConfig = intendedIdentityConfig
