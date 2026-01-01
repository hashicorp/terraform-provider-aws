// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package appstream

// Exports for use in tests only.
var (
	ResourceApplicationEntitlementAssociation = resourceApplicationEntitlementAssociation
	ResourceDirectoryConfig                   = resourceDirectoryConfig
	ResourceEntitlement                       = resourceEntitlement
	ResourceFleet                             = resourceFleet
	ResourceFleetStackAssociation             = resourceFleetStackAssociation
	ResourceImageBuilder                      = resourceImageBuilder
	ResourceStack                             = resourceStack
	ResourceUser                              = resourceUser
	ResourceUserStackAssociation              = resourceUserStackAssociation

	FindApplicationEntitlementAssociationByThreePartKey = findApplicationEntitlementAssociationByThreePartKey
	FindDirectoryConfigByID                             = findDirectoryConfigByID
	FindEntitlementByTwoPartKey                         = findEntitlementByTwoPartKey
	FindFleetByID                                       = findFleetByID
	FindFleetStackAssociationByTwoPartKey               = findFleetStackAssociationByTwoPartKey
	FindImageBuilderByID                                = findImageBuilderByID
	FindStackByID                                       = findStackByID
	FindUserByTwoPartKey                                = findUserByTwoPartKey
	FindUserStackAssociationByThreePartKey              = findUserStackAssociationByThreePartKey
)
