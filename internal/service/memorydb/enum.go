// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

// WARNING: As of 01/2022, the MemoryDB API does not provide a formal definition
// of its enumerations, and what documentation there is does not align with the
// API itself. The following values were determined experimentally, and is
// unlikely to be exhaustive.

const (
	aclStatusActive    = "active"
	aclStatusCreating  = "creating"
	aclStatusDeleting  = "deleting"
	aclStatusModifying = "modifying"
)

const (
	clusterStatusAvailable    = "available"
	clusterStatusCreating     = "creating"
	clusterStatusDeleting     = "deleting"
	clusterStatusSnapshotting = "snapshotting"
	clusterStatusUpdating     = "updating"
)

const (
	clusterParameterGroupStatusApplying = "applying"
	clusterParameterGroupStatusInSync   = "in-sync"
)

const (
	clusterSecurityGroupStatusActive    = "active"
	clusterSecurityGroupStatusModifying = "modifying"
)

const (
	clusterShardStatusAvailable = "available"
	clusterShardStatusCreating  = "creating"
	clusterShardStatusDeleting  = "deleting"
	clusterShardStatusModifying = "modifying"
)

const (
	clusterSNSTopicStatusActive   = "ACTIVE"
	clusterSNSTopicStatusInactive = "INACTIVE"
)

const (
	snapshotStatusAvailable = "available"
	snapshotStatusCopying   = "copying"
	snapshotStatusCreating  = "creating"
	snapshotStatusDeleting  = "deleting"
	snapshotStatusRestoring = "restoring"
)

const (
	userStatusActive    = "active"
	userStatusDeleting  = "deleting"
	userStatusModifying = "modifying"
)

type clusterEngine string

const (
	clusterEngineRedis  clusterEngine = "redis"
	clusterEngineValkey clusterEngine = "valkey"
)

func (clusterEngine) Values() []clusterEngine {
	return []clusterEngine{
		clusterEngineRedis,
		clusterEngineValkey,
	}
}
