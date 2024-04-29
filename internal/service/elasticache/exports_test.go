// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

// Exports for use in tests only.
var (
	ResourceReplicationGroup     = resourceReplicationGroup
	ResourceServerlessCache      = newServerlessCacheResource
	ResourceSubnetGroup          = resourceSubnetGroup
	ResourceUser                 = resourceUser
	ResourceUserGroup            = resourceUserGroup
	ResourceUserGroupAssociation = resourceUserGroupAssociation

	FindCacheSubnetGroupByName           = findCacheSubnetGroupByName
	FindReplicationGroupByID             = findReplicationGroupByID
	FindServerlessCacheByID              = findServerlessCacheByID
	FindUserByID                         = findUserByID
	FindUserGroupByID                    = findUserGroupByID
	FindUserGroupAssociationByTwoPartKey = findUserGroupAssociationByTwoPartKey
	WaitReplicationGroupAvailable        = waitReplicationGroupAvailable
)
