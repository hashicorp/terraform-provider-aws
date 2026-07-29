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

	FindConfigurationSetByID         = findConfigurationSetByID
	FindEventDestinationByTwoPartKey = findEventDestinationByTwoPartKey
	FindOptOutListByID               = findOptOutListByID
	FindPhoneNumberByID              = findPhoneNumberByID
	FindPoolByID                     = findPoolByID

	ValidatePhoneIdentity  = validatePhoneIdentity
	ValidateSenderIdentity = validateSenderIdentity
)

type IntendedIdentityConfig = intendedIdentityConfig
