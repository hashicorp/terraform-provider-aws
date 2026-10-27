// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

// Exports for use in tests only.
var (
	ResourceCollector            = newCollectorResource
	ResourceCollectorAssociation = newCollectorAssociationResource
	ResourceEnvironment          = newEnvironmentResource
	ResourceIntegration          = newIntegrationResource
	ResourceResourcePolicy       = newResourcePolicyResource

	FindAssociatedCollector = findAssociatedCollector
	FindCollectorByID       = findCollectorByID
	FindEnvironmentByID     = findEnvironmentByID
	FindIntegrationByID     = findIntegrationByID
	FindResourcePolicyByARN = findResourcePolicyByARN
)
