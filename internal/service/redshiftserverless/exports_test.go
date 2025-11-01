// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

// Exports for use in tests only.
var (
	ResourceCustomDomainAssociation   = newCustomDomainAssociationResource
	ResourceEndpointAccess            = resourceEndpointAccess
	ResourceNamespace                 = resourceNamespace
	ResourceResourcePolicy            = resourceResourcePolicy
	ResourceSnapshot                  = resourceSnapshot
	ResourceSnapshotCopyConfiguration = newResourceSnapshotCopyConfiguration
	ResourceUsageLimit                = resourceUsageLimit
	ResourceWorkgroup                 = resourceWorkgroup

	FindCustomDomainAssociationByTwoPartKey = findCustomDomainAssociationByTwoPartKey
	FindEndpointAccessByName                = findEndpointAccessByName
	FindNamespaceByName                     = findNamespaceByName
	FindResourcePolicyByARN                 = findResourcePolicyByARN
	FindSnapshotByName                      = findSnapshotByName
	FindSnapshotCopyConfigurationByID       = findSnapshotCopyConfigurationByID
	FindUsageLimitByName                    = findUsageLimitByName
	FindWorkgroupByName                     = findWorkgroupByName
)
