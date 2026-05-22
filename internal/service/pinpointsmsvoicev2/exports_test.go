// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

// Exports for use in tests only.
var (
	ResourceConfigurationSet                 = newConfigurationSetResource
	ResourceConfigurationSetEventDestination = newConfigurationSetEventDestinationResource
	ResourceOptOutList                       = newOptOutListResource
	ResourcePhoneNumber                      = newPhoneNumberResource

	FindConfigurationSetByID                         = findConfigurationSetByID
	FindConfigurationSetEventDestinationByTwoPartKey = findConfigurationSetEventDestinationByTwoPartKey
	FindOptOutListByID                               = findOptOutListByID
	FindPhoneNumberByID                              = findPhoneNumberByID
)
