// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

// Exports for use in tests only.
var (
	ResourceAccessEntry             = resourceAccessEntry
	ResourceAccessPolicyAssociation = resourceAccessPolicyAssociation
	ResourceAddon                   = resourceAddon
	ResourceCluster                 = resourceCluster
	ResourceFargateProfile          = resourceFargateProfile
	ResourceIdentityProviderConfig  = resourceIdentityProviderConfig
	ResourceNodeGroup               = resourceNodeGroup
	ResourcePodIdentityAssociation  = newPodIdentityAssociationResource

	ClusterStateUpgradeV0                      = clusterStateUpgradeV0
	FindAccessEntryByTwoPartKey                = findAccessEntryByTwoPartKey
	FindAccessPolicyAssociationByThreePartKey  = findAccessPolicyAssociationByThreePartKey
	FindAddonByTwoPartKey                      = findAddonByTwoPartKey
	FindClusterByName                          = findClusterByName
	FindFargateProfileByTwoPartKey             = findFargateProfileByTwoPartKey
	FindNodegroupByTwoPartKey                  = findNodegroupByTwoPartKey
	FindOIDCIdentityProviderConfigByTwoPartKey = findOIDCIdentityProviderConfigByTwoPartKey
	FindPodIdentityAssociationByTwoPartKey     = findPodIdentityAssociationByTwoPartKey
)
