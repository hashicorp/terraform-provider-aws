// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudtrail

// Exports for use in tests only.
var (
	ResourceEventDataStore                    = resourceEventDataStore
	ResourceInsightSelectors                  = resourceInsightSelectors
	ResourceOrganizationDelegatedAdminAccount = newOrganizationDelegatedAdminAccountResource
	ResourceTrail                             = resourceTrail

	FindEventDataStoreByARN              = findEventDataStoreByARN
	FindInsightSelectorsByEventDataStore = findInsightSelectorsByEventDataStore
	FindInsightSelectorsByTrailName      = findInsightSelectorsByTrailName
	FindTrailByARN                       = findTrailByARN
	ServiceAccountPerRegionMap           = serviceAccountPerRegionMap
	ServicePrincipal                     = servicePrincipal
)
