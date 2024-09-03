// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/memorydb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// statusACL fetches the MemoryDB ACL and its status.
func statusACL(ctx context.Context, conn *memorydb.Client, aclName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		acl, err := FindACLByName(ctx, conn, aclName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return acl, aws.ToString(acl.Status), nil
	}
}

// statusCluster fetches the MemoryDB Cluster and its status.
func statusCluster(ctx context.Context, conn *memorydb.Client, clusterName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := FindClusterByName(ctx, conn, clusterName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return cluster, aws.ToString(cluster.Status), nil
	}
}

// statusClusterParameterGroup fetches the MemoryDB Cluster and its parameter group status.
func statusClusterParameterGroup(ctx context.Context, conn *memorydb.Client, clusterName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := FindClusterByName(ctx, conn, clusterName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return cluster, aws.ToString(cluster.ParameterGroupStatus), nil
	}
}

// statusClusterSecurityGroups fetches the MemoryDB Cluster and its security group status.
func statusClusterSecurityGroups(ctx context.Context, conn *memorydb.Client, clusterName string) retry.StateRefreshFunc {
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

			if aws.ToString(sg.Status) != ClusterSecurityGroupStatusActive {
				return cluster, ClusterSecurityGroupStatusModifying, nil
			}
		}

		return cluster, ClusterSecurityGroupStatusActive, nil
	}
}

// statusSnapshot fetches the MemoryDB Snapshot and its status.
func statusSnapshot(ctx context.Context, conn *memorydb.Client, snapshotName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		snapshot, err := FindSnapshotByName(ctx, conn, snapshotName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return snapshot, aws.ToString(snapshot.Status), nil
	}
}

// statusUser fetches the MemoryDB user and its status.
func statusUser(ctx context.Context, conn *memorydb.Client, userName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		user, err := FindUserByName(ctx, conn, userName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return user, aws.ToString(user.Status), nil
	}
}
