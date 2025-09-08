// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

// Exports for use in tests only.
var (
	ResourceCachePolicy                 = resourceCachePolicy
	ResourceContinuousDeploymentPolicy  = newContinuousDeploymentPolicyResource
	ResourceDistribution                = resourceDistribution
	ResourceDistributionTenant          = resourceDistributionTenant
	ResourceFieldLevelEncryptionConfig  = resourceFieldLevelEncryptionConfig
	ResourceFieldLevelEncryptionProfile = resourceFieldLevelEncryptionProfile
	ResourceFunction                    = resourceFunction
	ResourceKeyGroup                    = resourceKeyGroup
	ResourceMonitoringSubscription      = resourceMonitoringSubscription
	ResourceOriginAccessControl         = resourceOriginAccessControl
	ResourceOriginAccessIdentity        = resourceOriginAccessIdentity
	ResourceOriginRequestPolicy         = resourceOriginRequestPolicy
	ResourcePublicKey                   = resourcePublicKey
	ResourceRealtimeLogConfig           = resourceRealtimeLogConfig
	ResourceResponseHeadersPolicy       = resourceResponseHeadersPolicy
	ResourceVPCOrigin                   = newVPCOriginResource

	FindCachePolicyByID                        = findCachePolicyByID
	FindContinuousDeploymentPolicyByID         = findContinuousDeploymentPolicyByID
	FindDistributionTenantById                 = findDistributionTenantByID
	FindFieldLevelEncryptionConfigByID         = findFieldLevelEncryptionConfigByID
	FindFieldLevelEncryptionProfileByID        = findFieldLevelEncryptionProfileByID
	FindFunctionByTwoPartKey                   = findFunctionByTwoPartKey
	FindKeyGroupByID                           = findKeyGroupByID
	FindKeyValueStoreByName                    = findKeyValueStoreByName
	FindMonitoringSubscriptionByDistributionID = findMonitoringSubscriptionByDistributionID
	FindOriginAccessControlByID                = findOriginAccessControlByID
	FindOriginAccessIdentityByID               = findOriginAccessIdentityByID
	FindOriginRequestPolicyByID                = findOriginRequestPolicyByID
	FindPublicKeyByID                          = findPublicKeyByID
	FindRealtimeLogConfigByARN                 = findRealtimeLogConfigByARN
	FindResponseHeadersPolicyByID              = findResponseHeadersPolicyByID
	FindVPCOriginByID                          = findVPCOriginByID
	WaitDistributionDeployed                   = waitDistributionDeployed
)
