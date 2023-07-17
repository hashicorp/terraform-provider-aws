// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusGlobalClusterRefreshFunc(ctx context.Context, conn *docdb.DocDB, globalClusterID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		globalCluster, err := FindGlobalClusterById(ctx, conn, globalClusterID)

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeGlobalClusterNotFoundFault) || globalCluster == nil {
			return nil, GlobalClusterStatusDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("reading DocumentDB Global Cluster (%s): %w", globalClusterID, err)
		}

		return globalCluster, aws.StringValue(globalCluster.Status), nil
	}
}

func statusDBClusterRefreshFunc(ctx context.Context, conn *docdb.DocDB, dBClusterID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		dBCluster, err := FindDBClusterById(ctx, conn, dBClusterID)

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBClusterNotFoundFault) || dBCluster == nil {
			return nil, DBClusterStatusDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("reading DocumentDB Cluster (%s): %w", dBClusterID, err)
		}

		return dBCluster, aws.StringValue(dBCluster.Status), nil
	}
}

func statusDBClusterSnapshotRefreshFunc(ctx context.Context, conn *docdb.DocDB, dBClusterSnapshotID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		dBClusterSnapshot, err := FindDBClusterSnapshotById(ctx, conn, dBClusterSnapshotID)

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBClusterSnapshotNotFoundFault) || dBClusterSnapshot == nil {
			return nil, DBClusterSnapshotStatusDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("reading DocumentDB Cluster Snapshot (%s): %w", dBClusterSnapshotID, err)
		}

		return dBClusterSnapshot, aws.StringValue(dBClusterSnapshot.Status), nil
	}
}

func statusDBInstanceRefreshFunc(ctx context.Context, conn *docdb.DocDB, dBInstanceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		dBInstance, err := FindDBInstanceById(ctx, conn, dBInstanceID)

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBInstanceNotFoundFault) || dBInstance == nil {
			return nil, DBInstanceStatusDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("reading DocumentDB Instance (%s): %w", dBInstanceID, err)
		}

		return dBInstance, aws.StringValue(dBInstance.DBInstanceStatus), nil
	}
}

func statusDBSubnetGroupRefreshFunc(ctx context.Context, conn *docdb.DocDB, dBSubnetGroupName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		dBSubnetGroup, err := FindDBSubnetGroupByName(ctx, conn, dBSubnetGroupName)

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBSubnetGroupNotFoundFault) || dBSubnetGroup == nil {
			return nil, DBSubnetGroupStatusDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("reading DocumentDB Subnet Group (%s): %w", dBSubnetGroupName, err)
		}

		return dBSubnetGroup, aws.StringValue(dBSubnetGroup.SubnetGroupStatus), nil
	}
}

func statusEventSubscription(ctx context.Context, conn *docdb.DocDB, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEventSubscriptionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
