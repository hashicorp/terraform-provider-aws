// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

// Exports for use in tests only.
var (
	ResourceOptOutList  = newOptOutListResource
	ResourcePhoneNumber = newPhoneNumberResource

	FindOptOutListByID  = findOptOutListByID
	FindPhoneNumberByID = findPhoneNumberByID
)
