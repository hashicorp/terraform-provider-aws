// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains

// Exports for use in tests only.
var (
	ResourceDelegationSignerRecord = newDelegationSignerRecordResource
	ResourceDomain                 = newDomainResource
	ResourceRegisteredDomain       = resourceRegisteredDomain

	FindDNSSECKeyByTwoPartKey = findDNSSECKeyByTwoPartKey
	FindDomainDetailByName    = findDomainDetailByName
)
