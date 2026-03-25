// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ram

// Exports for use in tests only.
var (
	ResourcePermission                         = newPermissionResource
	ResourcePrincipalAssociation               = resourcePrincipalAssociation
	ResourceResourceAssociation                = resourceResourceAssociation
	ResourceResourceShare                      = resourceResourceShare
	ResourceResourceShareAccepter              = resourceResourceShareAccepter
	ResourceResourceShareAssociationsExclusive = newResourceShareAssociationsExclusiveResource
	ResourceSharingWithOrganization            = resourceSharingWithOrganization

	FindAssociationsForResourceShare         = findAssociationsForResourceShare
	FindPermissionByARN                      = findPermissionByARN
	FindPrincipalAssociationByTwoPartKey     = findPrincipalAssociationByTwoPartKey
	FindResourceAssociationByTwoPartKey      = findResourceAssociationByTwoPartKey
	FindResourceShareOwnerOtherAccountsByARN = findResourceShareOwnerOtherAccountsByARN
	FindResourceShareOwnerSelfByARN          = findResourceShareOwnerSelfByARN

	WaitResourceAssociationCreated = waitResourceAssociationCreated
)
