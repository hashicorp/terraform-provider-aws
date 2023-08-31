// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

// WARNING: As of 01/2022, the MemoryDB API does not provide a formal definition
// of its enumerations, and what documentation there is does not align with the
// API itself. The following values were determined experimentally, and is
// unlikely to be exhaustive.

const (
	ACLStatusActive    = "active"
	ACLStatusCreating  = "creating"
	ACLStatusDeleting  = "deleting"
	ACLStatusModifying = "modifying"
)

func ACLStatus_Values() []string {
	return []string{
		ACLStatusActive,
		ACLStatusCreating,
		ACLStatusDeleting,
		ACLStatusModifying,
	}
}

const (
	ClusterStatusAvailable    = "available"
	ClusterStatusCreating     = "creating"
	ClusterStatusDeleting     = "deleting"
	ClusterStatusSnapshotting = "snapshotting"
	ClusterStatusUpdating     = "updating"
)

func ClusterStatus_Values() []string {
	return []string{
		ClusterStatusAvailable,
		ClusterStatusCreating,
		ClusterStatusDeleting,
		ClusterStatusSnapshotting,
		ClusterStatusUpdating,
	}
}

const (
	ClusterParameterGroupStatusApplying = "applying"
	ClusterParameterGroupStatusInSync   = "in-sync"
)

func ClusterParameterGroupStatus_Values() []string {
	return []string{
		ClusterParameterGroupStatusApplying,
		ClusterParameterGroupStatusInSync,
	}
}

const (
	ClusterSecurityGroupStatusActive    = "active"
	ClusterSecurityGroupStatusModifying = "modifying"
)

func ClusterSecurityGroupStatus_Values() []string {
	return []string{
		ClusterSecurityGroupStatusActive,
		ClusterSecurityGroupStatusModifying,
	}
}

const (
	ClusterShardStatusAvailable = "available"
	ClusterShardStatusCreating  = "creating"
	ClusterShardStatusDeleting  = "deleting"
	ClusterShardStatusModifying = "modifying"
)

func ClusterShardStatus_Values() []string {
	return []string{
		ClusterShardStatusAvailable,
		ClusterShardStatusCreating,
		ClusterShardStatusDeleting,
		ClusterShardStatusModifying,
	}
}

const (
	ClusterSNSTopicStatusActive   = "ACTIVE"
	ClusterSNSTopicStatusInactive = "INACTIVE"
)

func ClusterSNSTopicStatus_Values() []string {
	return []string{
		ClusterSNSTopicStatusActive,
		ClusterSNSTopicStatusInactive,
	}
}

const (
	SnapshotStatusAvailable = "available"
	SnapshotStatusCopying   = "copying"
	SnapshotStatusCreating  = "creating"
	SnapshotStatusDeleting  = "deleting"
	SnapshotStatusRestoring = "restoring"
)

func SnapshotStatus_Values() []string {
	return []string{
		SnapshotStatusCreating,
		SnapshotStatusAvailable,
		SnapshotStatusRestoring,
		SnapshotStatusCopying,
		SnapshotStatusDeleting,
	}
}

const (
	UserStatusActive    = "active"
	UserStatusDeleting  = "deleting"
	UserStatusModifying = "modifying"
)

func UserStatus_Values() []string {
	return []string{
		UserStatusActive,
		UserStatusDeleting,
		UserStatusModifying,
	}
}
