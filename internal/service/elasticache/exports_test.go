// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

// Exports for use in tests only.
var (
	ResourceCluster                = resourceCluster
	ResourceGlobalReplicationGroup = resourceGlobalReplicationGroup
	ResourceParameterGroup         = resourceParameterGroup
	ResourceReplicationGroup       = resourceReplicationGroup
	ResourceServerlessCache        = newServerlessCacheResource
	ResourceSubnetGroup            = resourceSubnetGroup
	ResourceUser                   = resourceUser
	ResourceUserGroup              = resourceUserGroup
	ResourceUserGroupAssociation   = resourceUserGroupAssociation

	FindCacheClusterByID                 = findCacheClusterByID
	FindCacheParameterGroup              = findCacheParameterGroup
	FindCacheParameterGroupByName        = findCacheParameterGroupByName
	FindCacheSubnetGroupByName           = findCacheSubnetGroupByName
	FindGlobalReplicationGroupByID       = findGlobalReplicationGroupByID
	FindReplicationGroupByID             = findReplicationGroupByID
	FindServerlessCacheByID              = findServerlessCacheByID
	FindUserByID                         = findUserByID
	FindUserGroupByID                    = findUserGroupByID
	FindUserGroupAssociationByTwoPartKey = findUserGroupAssociationByTwoPartKey
	ParameterChanges                     = parameterChanges
	ParameterHash                        = parameterHash
	WaitCacheClusterDeleted              = waitCacheClusterDeleted
	WaitReplicationGroupAvailable        = waitReplicationGroupAvailable

	DeleteCacheCluster                        = deleteCacheCluster
	DiffVersion                               = diffVersion
	EmptyDescription                          = emptyDescription
	EngineMemcached                           = engineMemcached
	EngineRedis                               = engineRedis
	EngineVersionForceNewOnDowngrade          = engineVersionForceNewOnDowngrade
	EngineVersionIsDowngrade                  = engineVersionIsDowngrade
	GlobalReplicationGroupRegionPrefixFormat  = globalReplicationGroupRegionPrefixFormat
	NormalizeEngineVersion                    = normalizeEngineVersion
	ParamGroupNameRequiresMajorVersionUpgrade = paramGroupNameRequiresMajorVersionUpgrade
	ValidateClusterEngineVersion              = validateClusterEngineVersion
	ValidMemcachedVersionString               = validMemcachedVersionString
	ValidRedisVersionString                   = validRedisVersionString
)

type (
	VersionDiff = versionDiff
)
