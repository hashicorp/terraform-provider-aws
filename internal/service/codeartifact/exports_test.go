// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package codeartifact

// Exports for use in tests only.
var (
	ResourceDomain                      = resourceDomain
	ResourceDomainPermissionsPolicy     = resourceDomainPermissionsPolicy
	ResourceRepository                  = resourceRepository
	ResourceRepositoryPermissionsPolicy = resourceRepositoryPermissionsPolicy

	FindDomainByTwoPartKey                        = findDomainByTwoPartKey
	FindDomainPermissionsPolicyByTwoPartKey       = findDomainPermissionsPolicyByTwoPartKey
	FindRepositoryByThreePartKey                  = findRepositoryByThreePartKey
	FindRepositoryPermissionsPolicyByThreePartKey = findRepositoryPermissionsPolicyByThreePartKey
)
