// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

// Exports for use in tests only.
var (
	ResourceLifecyclePolicy               = resourceLifecyclePolicy
	ResourcePullThroughCacheRule          = resourcePullThroughCacheRule
	ResourceRegistryPolicy                = resourceRegistryPolicy
	ResourceRegistryScanningConfiguration = resourceRegistryScanningConfiguration
	ResourceReplicationConfiguration      = resourceReplicationConfiguration
	ResourceRepository                    = resourceRepository
	ResourceRepositoryPolicy              = resourceRepositoryPolicy

	FindLifecyclePolicyByRepositoryName        = findLifecyclePolicyByRepositoryName
	FindPullThroughCacheRuleByRepositoryPrefix = findPullThroughCacheRuleByRepositoryPrefix
	FindRegistryPolicy                         = findRegistryPolicy
	FindRegistryScanningConfiguration          = findRegistryScanningConfiguration
	FindReplicationConfiguration               = findReplicationConfiguration
	FindRepositoryByName                       = findRepositoryByName
	FindRepositoryPolicyByRepositoryName       = findRepositoryPolicyByRepositoryName
)
