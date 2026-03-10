// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package workmail

// Exports for use in tests only.
var (
	ResourceOrganization  = newOrganizationResource
	ResourceDomain        = newDomainResource
	ResourceDefaultDomain = newDefaultDomainResource

	FindOrganizationByID     = findOrganizationByID
	FindDomainByOrgAndName   = findDomainByOrgAndName
	FindDefaultDomainByOrgID = findDefaultDomainByOrgID
)
