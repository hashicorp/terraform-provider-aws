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

	FindResourceAssociationByTwoPartKey                = findResourceAssociationByTwoPartKey
	FindResourceShareAssociationByShareARNAndPrincipal = findResourceShareAssociationByShareARNAndPrincipal
	FindResourceShareOwnerOtherAccountsByARN           = findResourceShareOwnerOtherAccountsByARN
	FindResourceShareOwnerSelfByARN                    = findResourceShareOwnerSelfByARN
)
