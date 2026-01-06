// Copyright IBM Corp. 2014, 2025
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
)
