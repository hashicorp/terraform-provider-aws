// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr

// Exports for use in tests only.
var (
	ResourceLifecyclePolicy               = resourceLifecyclePolicy
	ResourcePullThroughCacheRule          = resourcePullThroughCacheRule
	ResourcePullTimeUpdateExclusion       = newPullTimeUpdateExclusionResource
	ResourceRegistryPolicy                = resourceRegistryPolicy
	ResourceRegistryScanningConfiguration = resourceRegistryScanningConfiguration
	ResourceReplicationConfiguration      = resourceReplicationConfiguration
	ResourceRepository                    = resourceRepository
	ResourceRepositoryCreationTemplate    = resourceRepositoryCreationTemplate
	ResourceRepositoryPolicy              = resourceRepositoryPolicy

	FindAccountSettingByName                         = findAccountSettingByName
	FindLifecyclePolicyByRepositoryName              = findLifecyclePolicyByRepositoryName
	FindPullThroughCacheRuleByRepositoryPrefix       = findPullThroughCacheRuleByRepositoryPrefix
	FindPullTimeUpdateExclusionByPrincipalARN        = findPullTimeUpdateExclusionByPrincipalARN
	FindRegistryPolicy                               = findRegistryPolicy
	FindRegistryScanningConfiguration                = findRegistryScanningConfiguration
	FindReplicationConfiguration                     = findReplicationConfiguration
	FindRepositoryByName                             = findRepositoryByName
	FindRepositoryCreationTemplateByRepositoryPrefix = findRepositoryCreationTemplateByRepositoryPrefix
	FindRepositoryPolicyByRepositoryName             = findRepositoryPolicyByRepositoryName
)
