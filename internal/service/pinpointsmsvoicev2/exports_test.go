// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

// Exports for use in tests only.
var (
	ResourceConfigurationSet    = newConfigurationSetResource
	ResourceEventDestination    = newEventDestinationResource
	ResourceKeyword             = newKeywordResource
	ResourceNotifyConfiguration = newNotifyConfigurationResource
	ResourceOptOutList          = newOptOutListResource
	ResourcePhoneNumber         = newPhoneNumberResource
	ResourcePool                = newPoolResource
	ResourceResourcePolicy      = newResourcePolicyResource
	ResourceSenderID            = newSenderIDResource

	FindConfigurationSetByID         = findConfigurationSetByID
	FindEventDestinationByTwoPartKey = findEventDestinationByTwoPartKey
	FindKeywordByTwoPartKey          = findKeywordByTwoPartKey
	FindNotifyConfigurationByID      = findNotifyConfigurationByID
	FindOptOutListByID               = findOptOutListByID
	FindPhoneNumberByID              = findPhoneNumberByID
	FindPoolByID                     = findPoolByID
	FindResourcePolicyByARN          = findResourcePolicyByARN
	FindSenderIDByTwoPartKey         = findSenderIDByTwoPartKey

	ValidatePhoneIdentity  = validatePhoneIdentity
	ValidateSenderIdentity = validateSenderIdentity
)

type IntendedIdentityConfig = intendedIdentityConfig
