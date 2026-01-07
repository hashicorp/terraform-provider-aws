// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront

// Exports for use in tests only.
var (
	ResourceCachePolicy                 = resourceCachePolicy
	ResourceConnectionFunction          = newResourceConnectionFunction
	ResourceConnectionGroup             = newConnectionGroupResource
	ResourceContinuousDeploymentPolicy  = newContinuousDeploymentPolicyResource
	ResourceDistribution                = resourceDistribution
	ResourceDistributionTenant          = newDistributionTenantResource
	ResourceFieldLevelEncryptionConfig  = resourceFieldLevelEncryptionConfig
	ResourceFieldLevelEncryptionProfile = resourceFieldLevelEncryptionProfile
	ResourceFunction                    = resourceFunction
	ResourceKeyGroup                    = resourceKeyGroup
	ResourceMonitoringSubscription      = resourceMonitoringSubscription
	ResourceMultiTenantDistribution     = newMultiTenantDistributionResource
	ResourceOriginAccessControl         = resourceOriginAccessControl
	ResourceOriginAccessIdentity        = resourceOriginAccessIdentity
	ResourceOriginRequestPolicy         = resourceOriginRequestPolicy
	ResourcePublicKey                   = resourcePublicKey
	ResourceRealtimeLogConfig           = resourceRealtimeLogConfig
	ResourceResponseHeadersPolicy       = resourceResponseHeadersPolicy
	ResourceTrustStore                  = newTrustStoreResource
	ResourceVPCOrigin                   = newVPCOriginResource

	FindCachePolicyByID                        = findCachePolicyByID
	FindConnectionFunctionByTwoPartKey         = findConnectionFunctionByTwoPartKey
	FindConnectionGroupById                    = findConnectionGroupByID
	FindConnectionGroupByRoutingEndpoint       = findConnectionGroupByRoutingEndpoint
	FindContinuousDeploymentPolicyByID         = findContinuousDeploymentPolicyByID
	FindDistributionTenantByIdentifier         = findDistributionTenantByIdentifier
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
	FindTrustStoreByID                         = findTrustStoreByID
	FindVPCOriginByID                          = findVPCOriginByID

	WaitDistributionDeployed = waitDistributionDeployed
)
