// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

// Exports for use in tests only.
var (
	ResourceConfigurationSet = newConfigurationSetResource
	ResourceOptOutList       = newOptOutListResource
	ResourcePhoneNumber      = newPhoneNumberResource

	FindConfigurationSetByID = findConfigurationSetByID
	FindOptOutListByID       = findOptOutListByID
	FindPhoneNumberByID      = findPhoneNumberByID
)
