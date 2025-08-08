// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

// Exports for use in tests only.
var (
	ResourcePrincipalAssociation    = resourcePrincipalAssociation
	ResourceResourceAssociation     = resourceResourceAssociation
	ResourceResourceShare           = resourceResourceShare
	ResourceResourceShareAccepter   = resourceResourceShareAccepter
	ResourceSharingWithOrganization = resourceSharingWithOrganization

	FindPrincipalAssociationByTwoPartKey     = findPrincipalAssociationByTwoPartKey
	FindResourceAssociationByTwoPartKey      = findResourceAssociationByTwoPartKey
	FindResourceShareOwnerOtherAccountsByARN = findResourceShareOwnerOtherAccountsByARN
	FindResourceShareOwnerSelfByARN          = findResourceShareOwnerSelfByARN
	FindSharingWithOrganization              = findSharingWithOrganization
)
