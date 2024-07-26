// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// statusACL fetches the MemoryDB ACL and its status.
func statusACL(ctx context.Context, conn *memorydb.MemoryDB, aclName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		acl, err := FindACLByName(ctx, conn, aclName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return acl, aws.StringValue(acl.Status), nil
	}
}

// statusCluster fetches the MemoryDB Cluster and its status.
func statusCluster(ctx context.Context, conn *memorydb.MemoryDB, clusterName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := FindClusterByName(ctx, conn, clusterName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return cluster, aws.StringValue(cluster.Status), nil
	}
}

// statusClusterParameterGroup fetches the MemoryDB Cluster and its parameter group status.
func statusClusterParameterGroup(ctx context.Context, conn *memorydb.MemoryDB, clusterName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := FindClusterByName(ctx, conn, clusterName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return cluster, aws.StringValue(cluster.ParameterGroupStatus), nil
	}
}

// statusClusterSecurityGroups fetches the MemoryDB Cluster and its security group status.
func statusClusterSecurityGroups(ctx context.Context, conn *memorydb.MemoryDB, clusterName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := FindClusterByName(ctx, conn, clusterName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		for _, sg := range cluster.SecurityGroups {
			// When at least one security group change is being applied (whether
			// that be adding or removing an SG), say that we're still in progress.

			if aws.StringValue(sg.Status) != ClusterSecurityGroupStatusActive {
				return cluster, ClusterSecurityGroupStatusModifying, nil
			}
		}

		return cluster, ClusterSecurityGroupStatusActive, nil
	}
}

// statusSnapshot fetches the MemoryDB Snapshot and its status.
func statusSnapshot(ctx context.Context, conn *memorydb.MemoryDB, snapshotName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		snapshot, err := FindSnapshotByName(ctx, conn, snapshotName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return snapshot, aws.StringValue(snapshot.Status), nil
	}
}

// statusUser fetches the MemoryDB user and its status.
func statusUser(ctx context.Context, conn *memorydb.MemoryDB, userName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		user, err := FindUserByName(ctx, conn, userName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return user, aws.StringValue(user.Status), nil
	}
}
