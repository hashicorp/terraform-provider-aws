// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

// Exports for use in tests only.
var (
	ResourceConfigurationSet = newConfigurationSetResource
	ResourceEventDestination = newEventDestinationResource
	ResourceOptOutList       = newOptOutListResource
	ResourcePhoneNumber      = newPhoneNumberResource

	FindConfigurationSetByID         = findConfigurationSetByID
	FindEventDestinationByTwoPartKey = findEventDestinationByTwoPartKey
	FindOptOutListByID               = findOptOutListByID
	FindPhoneNumberByID              = findPhoneNumberByID
)
