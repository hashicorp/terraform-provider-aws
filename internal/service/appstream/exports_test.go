// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appstream

// Exports for use in tests only.
var (
	ResourceDirectoryConfig       = resourceDirectoryConfig
	ResourceFleet                 = resourceFleet
	ResourceFleetStackAssociation = resourceFleetStackAssociation
	ResourceImageBuilder          = resourceImageBuilder
	ResourceStack                 = resourceStack
	ResourceUser                  = resourceUser
	ResourceUserStackAssociation  = resourceUserStackAssociation

	FindDirectoryConfigByID                = findDirectoryConfigByID
	FindFleetByID                          = findFleetByID
	FindFleetStackAssociationByTwoPartKey  = findFleetStackAssociationByTwoPartKey
	FindImageBuilderByID                   = findImageBuilderByID
	FindStackByID                          = findStackByID
	FindUserByTwoPartKey                   = findUserByTwoPartKey
	FindUserStackAssociationByThreePartKey = findUserStackAssociationByThreePartKey
)
