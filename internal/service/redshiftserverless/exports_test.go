// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

// Exports for use in tests only.
var (
	ResourceCustomDomainAssociation = newResourceCustomDomainAssociation
	ResourceEndpointAccess          = resourceEndpointAccess
	ResourceNamespace               = resourceNamespace
	ResourceResourcePolicy          = resourceResourcePolicy
	ResourceSnapshot                = resourceSnapshot
	ResourceUsageLimit              = resourceUsageLimit
	ResourceWorkgroup               = resourceWorkgroup

	FindEndpointAccessByName = findEndpointAccessByName
	FindNamespaceByName      = findNamespaceByName
	FindResourcePolicyByARN  = findResourcePolicyByARN
	FindSnapshotByName       = findSnapshotByName
	FindUsageLimitByName     = findUsageLimitByName
)
