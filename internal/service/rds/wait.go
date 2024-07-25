// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func waitDBClusterRoleAssociationCreated(ctx context.Context, conn *rds.RDS, dbClusterID, roleARN string, timeout time.Duration) (*rds.DBClusterRole, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterRoleStatusPending},
		Target:     []string{clusterRoleStatusActive},
		Refresh:    statusDBClusterRole(ctx, conn, dbClusterID, roleARN),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBClusterRole); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterRoleAssociationDeleted(ctx context.Context, conn *rds.RDS, dbClusterID, roleARN string, timeout time.Duration) (*rds.DBClusterRole, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{clusterRoleStatusActive, clusterRoleStatusPending},
		Target:     []string{},
		Refresh:    statusDBClusterRole(ctx, conn, dbClusterID, roleARN),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBClusterRole); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterInstanceCreated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			InstanceStatusBackingUp,
			InstanceStatusConfiguringEnhancedMonitoring,
			InstanceStatusConfiguringIAMDatabaseAuth,
			InstanceStatusConfiguringLogExports,
			InstanceStatusCreating,
			InstanceStatusMaintenance,
			InstanceStatusModifying,
			InstanceStatusRebooting,
			InstanceStatusRenaming,
			InstanceStatusResettingMasterCredentials,
			InstanceStatusStarting,
			InstanceStatusStorageOptimization,
			InstanceStatusUpgrading,
		},
		Target:     []string{InstanceStatusAvailable},
		Refresh:    statusDBInstanceSDKv1(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterInstanceUpdated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			InstanceStatusBackingUp,
			InstanceStatusConfiguringEnhancedMonitoring,
			InstanceStatusConfiguringIAMDatabaseAuth,
			InstanceStatusConfiguringLogExports,
			InstanceStatusCreating,
			InstanceStatusMaintenance,
			InstanceStatusModifying,
			InstanceStatusRebooting,
			InstanceStatusRenaming,
			InstanceStatusResettingMasterCredentials,
			InstanceStatusStarting,
			InstanceStatusStorageOptimization,
			InstanceStatusUpgrading,
		},
		Target:     []string{InstanceStatusAvailable},
		Refresh:    statusDBInstanceSDKv1(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterInstanceDeleted(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) (*rds.DBInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			InstanceStatusConfiguringLogExports,
			InstanceStatusDeletePreCheck,
			InstanceStatusDeleting,
			InstanceStatusModifying,
		},
		Target:     []string{},
		Refresh:    statusDBInstanceSDKv1(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func waitReservedInstanceCreated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			ReservedInstanceStatePaymentPending,
		},
		Target:         []string{ReservedInstanceStateActive},
		Refresh:        statusReservedInstance(ctx, conn, id),
		NotFoundChecks: 5,
		Timeout:        timeout,
		MinTimeout:     10 * time.Second,
		Delay:          30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitDBSnapshotCreated(ctx context.Context, conn *rds.RDS, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{DBSnapshotCreating},
		Target:     []string{DBSnapshotAvailable},
		Refresh:    statusDBSnapshot(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
