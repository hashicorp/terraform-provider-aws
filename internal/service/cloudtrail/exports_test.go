// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudtrail

// Exports for use in tests only.
var (
	ResourceEventDataStore                    = resourceEventDataStore
	ResourceOrganizationDelegatedAdminAccount = newOrganizationDelegatedAdminAccountResource
	ResourceTrail                             = resourceTrail

	FindEventDataStoreByARN    = findEventDataStoreByARN
	FindTrailByARN             = findTrailByARN
	ServiceAccountPerRegionMap = serviceAccountPerRegionMap
	ServicePrincipal           = servicePrincipal
)
