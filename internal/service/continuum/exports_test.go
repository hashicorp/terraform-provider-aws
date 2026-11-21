// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

// Exports for use in tests only.
var (
	ResourceIntegration          = newIntegrationResource
	ResourceCollector            = newCollectorResource
	ResourceEnvironment          = newEnvironmentResource
	ResourceCollectorAssociation = newCollectorAssociationResource

	FindIntegrationByID     = findIntegrationByID
	FindCollectorByID       = findCollectorByID
	FindEnvironmentByID     = findEnvironmentByID
	FindAssociatedCollector = findAssociatedCollector
)
