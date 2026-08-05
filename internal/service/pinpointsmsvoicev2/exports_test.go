// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

// Exports for use in tests only.
var (
	ResourceConfigurationSet = newConfigurationSetResource
	ResourceEventDestination = newEventDestinationResource
	ResourceOptOutList       = newOptOutListResource
	ResourcePhoneNumber      = newPhoneNumberResource
	ResourcePool             = newPoolResource
	ResourceResourcePolicy   = newResourcePolicyResource

	FindConfigurationSetByID         = findConfigurationSetByID
	FindEventDestinationByTwoPartKey = findEventDestinationByTwoPartKey
	FindOptOutListByID               = findOptOutListByID
	FindPhoneNumberByID              = findPhoneNumberByID
	FindPoolByID                     = findPoolByID
	FindResourcePolicyByARN          = findResourcePolicyByARN

	ValidatePhoneIdentity  = validatePhoneIdentity
	ValidateSenderIdentity = validateSenderIdentity
)

type IntendedIdentityConfig = intendedIdentityConfig
