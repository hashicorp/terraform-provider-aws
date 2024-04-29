// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

// Exports for use in tests only.
var (
	ResourceServerlessCache = newServerlessCacheResource
	ResourceSubnetGroup     = resourceSubnetGroup
	ResourceUser            = resourceUser
	ResourceUserGroup       = resourceUserGroup

	FindCacheSubnetGroupByName           = findCacheSubnetGroupByName
	FindServerlessCacheByID              = findServerlessCacheByID
	FindUserByID                         = findUserByID
	FindUserGroupByID                    = findUserGroupByID
	ReplicationGroupAvailableModifyDelay = replicationGroupAvailableModifyDelay
)
