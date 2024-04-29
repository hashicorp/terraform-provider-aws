// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

// Exports for use in tests only.
var (
	ResourceGlobalReplicationGroup = resourceGlobalReplicationGroup
	ResourceParameterGroup         = resourceParameterGroup
	ResourceReplicationGroup       = resourceReplicationGroup
	ResourceServerlessCache        = newServerlessCacheResource
	ResourceSubnetGroup            = resourceSubnetGroup
	ResourceUser                   = resourceUser
	ResourceUserGroup              = resourceUserGroup
	ResourceUserGroupAssociation   = resourceUserGroupAssociation

	FindCacheSubnetGroupByName           = findCacheSubnetGroupByName
	FindCacheParameterGroupByName        = findCacheParameterGroupByName
	FindGlobalReplicationGroupByID       = findGlobalReplicationGroupByID
	FindReplicationGroupByID             = findReplicationGroupByID
	FindServerlessCacheByID              = findServerlessCacheByID
	FindUserByID                         = findUserByID
	FindUserGroupByID                    = findUserGroupByID
	FindUserGroupAssociationByTwoPartKey = findUserGroupAssociationByTwoPartKey
	ParameterChanges                     = parameterChanges
	ParameterHash                        = parameterHash
	WaitReplicationGroupAvailable        = waitReplicationGroupAvailable
)
