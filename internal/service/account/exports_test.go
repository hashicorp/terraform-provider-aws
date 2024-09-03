// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account

// Exports for use in tests only.
var (
	AlternateContactParseResourceID  = alternateContactParseResourceID
	FindAlternateContactByTwoPartKey = findAlternateContactByTwoPartKey
	FindContactInformation           = findContactInformation

	ResourceAlternateContact = resourceAlternateContact
	ResourcePrimaryContact   = resourcePrimaryContact
)
