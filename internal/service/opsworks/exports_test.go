// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

// Exports for use in tests only.
var (
	ResourceRailsAppLayer = resourceRailsAppLayer
	ResourceRDSDBInstance = resourceRDSDBInstance
	ResourceStack         = resourceStack
	ResourceUserProfile   = resourceUserProfile

	FindAppByID                   = findAppByID
	FindInstanceByID              = findInstanceByID
	FindLayerByID                 = findLayerByID
	FindPermissionByTwoPartKey    = findPermissionByTwoPartKey
	FindRDSDBInstanceByTwoPartKey = findRDSDBInstanceByTwoPartKey
	FindStackByID                 = findStackByID
	FindUserProfileByARN          = findUserProfileByARN
)
