// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

// Exports for use in tests only.
var (
	ResourceAccessPolicy   = newResourceAccessPolicy
	ResourceCollection     = newResourceCollection
	ResourceSecurityConfig = newResourceSecurityConfig
	ResourceSecurityPolicy = newResourceSecurityPolicy
	ResourceVPCEndpoint    = newResourceVPCEndpoint

	FindAccessPolicyByNameAndType   = findAccessPolicyByNameAndType
	FindCollectionByID              = findCollectionByID
	FindSecurityConfigByID          = findSecurityConfigByID
	FindSecurityPolicyByNameAndType = findSecurityPolicyByNameAndType
	FindVPCEndpointByID             = findVPCEndpointByID
)
