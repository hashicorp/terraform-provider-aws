// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

// Exports for use in tests only.
var (
	ResourceAccessPolicy    = newAccessPolicyResource
	ResourceCollection      = newCollectionResource
	ResourceLifecyclePolicy = newLifecyclePolicyResource
	ResourceSecurityConfig  = newSecurityConfigResource
	ResourceSecurityPolicy  = newSecurityPolicyResource
	ResourceVPCEndpoint     = newVPCEndpointResource

	FindAccessPolicyByNameAndType    = findAccessPolicyByNameAndType
	FindCollectionByID               = findCollectionByID
	FindLifecyclePolicyByNameAndType = findLifecyclePolicyByNameAndType
	FindSecurityConfigByID           = findSecurityConfigByID
	FindSecurityPolicyByNameAndType  = findSecurityPolicyByNameAndType
	FindVPCEndpointByID              = findVPCEndpointByID
)
