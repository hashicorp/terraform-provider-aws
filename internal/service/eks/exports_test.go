// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

// Exports for use in tests only.
var (
	ResourceAddon                  = resourceAddon
	ResourceCluster                = resourceCluster
	ResourceFargateProfile         = resourceFargateProfile
	ResourceIdentityProviderConfig = resourceIdentityProviderConfig
	ResourceNodeGroup              = resourceNodeGroup

	FindAddonByTwoPartKey                      = findAddonByTwoPartKey
	FindClusterByName                          = findClusterByName
	FindFargateProfileByTwoPartKey             = findFargateProfileByTwoPartKey
	FindNodegroupByTwoPartKey                  = findNodegroupByTwoPartKey
	FindOIDCIdentityProviderConfigByTwoPartKey = findOIDCIdentityProviderConfigByTwoPartKey
)
