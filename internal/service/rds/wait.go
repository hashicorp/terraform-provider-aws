package rds

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	dbClusterRoleAssociationCreatedTimeout = 5 * time.Minute
	dbClusterRoleAssociationDeletedTimeout = 5 * time.Minute

	dbClusterActivityStreamStartedTimeout = 30 * time.Minute
	dbClusterActivityStreamStoppedTimeout = 30 * time.Minute
)

func waitEventSubscriptionCreated(conn *rds.RDS, id string, timeout time.Duration) (*rds.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{EventSubscriptionStatusCreating},
		Target:     []string{EventSubscriptionStatusActive},
		Refresh:    statusEventSubscription(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.EventSubscription); ok {
		return output, err
	}

	return nil, err
}

func waitEventSubscriptionDeleted(conn *rds.RDS, id string, timeout time.Duration) (*rds.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{EventSubscriptionStatusDeleting},
		Target:     []string{},
		Refresh:    statusEventSubscription(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.EventSubscription); ok {
		return output, err
	}

	return nil, err
}

func waitEventSubscriptionUpdated(conn *rds.RDS, id string, timeout time.Duration) (*rds.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{EventSubscriptionStatusModifying},
		Target:                    []string{EventSubscriptionStatusActive},
		Refresh:                   statusEventSubscription(conn, id),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		Delay:                     30 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.EventSubscription); ok {
		return output, err
	}

	return nil, err
}

// waitDBProxyEndpointAvailable waits for a DBProxyEndpoint to return Available
func waitDBProxyEndpointAvailable(conn *rds.RDS, id string, timeout time.Duration) (*rds.DBProxyEndpoint, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			rds.DBProxyEndpointStatusCreating,
			rds.DBProxyEndpointStatusModifying,
		},
		Target:  []string{rds.DBProxyEndpointStatusAvailable},
		Refresh: statusDBProxyEndpoint(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBProxyEndpoint); ok {
		return output, err
	}

	return nil, err
}

// waitDBProxyEndpointDeleted waits for a DBProxyEndpoint to return Deleted
func waitDBProxyEndpointDeleted(conn *rds.RDS, id string, timeout time.Duration) (*rds.DBProxyEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{rds.DBProxyEndpointStatusDeleting},
		Target:  []string{},
		Refresh: statusDBProxyEndpoint(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBProxyEndpoint); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterRoleAssociationCreated(conn *rds.RDS, dbClusterID, roleARN string) (*rds.DBClusterRole, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ClusterRoleStatusPending},
		Target:  []string{ClusterRoleStatusActive},
		Refresh: statusDBClusterRole(conn, dbClusterID, roleARN),
		Timeout: dbClusterRoleAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBClusterRole); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterRoleAssociationDeleted(conn *rds.RDS, dbClusterID, roleARN string) (*rds.DBClusterRole, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ClusterRoleStatusActive, ClusterRoleStatusPending},
		Target:  []string{},
		Refresh: statusDBClusterRole(conn, dbClusterID, roleARN),
		Timeout: dbClusterRoleAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBClusterRole); ok {
		return output, err
	}

	return nil, err
}

func waitDBInstanceDeleted(conn *rds.RDS, id string, timeout time.Duration) (*rds.DBInstance, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			InstanceStatusAvailable,
			InstanceStatusBackingUp,
			InstanceStatusConfiguringEnhancedMonitoring,
			InstanceStatusConfiguringLogExports,
			InstanceStatusCreating,
			InstanceStatusDeleting,
			InstanceStatusIncompatibleParameters,
			InstanceStatusIncompatibleRestore,
			InstanceStatusModifying,
			InstanceStatusStarting,
			InstanceStatusStopping,
			InstanceStatusStorageFull,
			InstanceStatusStorageOptimization,
		},
		Target:                    []string{},
		Refresh:                   statusDBInstance(conn, id),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		Delay:                     30 * time.Second,
		ContinuousTargetOccurence: 3,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func waitDBClusterInstanceDeleted(conn *rds.RDS, id string, timeout time.Duration) (*rds.DBInstance, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			InstanceStatusConfiguringLogExports,
			InstanceStatusDeleting,
			InstanceStatusModifying,
		},
		Target:     []string{},
		Refresh:    statusDBInstance(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}

// waitActivityStreamStarted waits for Aurora Cluster Activity Stream to be started
func waitActivityStreamStarted(ctx context.Context, conn *rds.RDS, dbClusterArn string) error {
	log.Printf("[DEBUG] Waiting for RDS Cluster Activity Stream %s to become started...", dbClusterArn)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{rds.ActivityStreamStatusStarting},
		Target:     []string{rds.ActivityStreamStatusStarted},
		Refresh:    statusDBClusterActivityStream(conn, dbClusterArn),
		Timeout:    dbClusterActivityStreamStartedTimeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for RDS Cluster Activity Stream (%s) to be started: %v", dbClusterArn, err)
	}
	return nil
}

// waitActivityStreamStarted waits for Aurora Cluster Activity Stream to be stopped
func waitActivityStreamStopped(ctx context.Context, conn *rds.RDS, dbClusterArn string) error {
	log.Printf("[DEBUG] Waiting for RDS Cluster Activity Stream %s to become stopped...", dbClusterArn)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{rds.ActivityStreamStatusStopping},
		Target:     []string{},
		Refresh:    statusDBClusterActivityStream(conn, dbClusterArn),
		Timeout:    dbClusterActivityStreamStoppedTimeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for RDS Cluster Activity Stream (%s) to be stopped: %v", dbClusterArn, err)
	}
	return nil
}

func waitDBInstanceAutomatedBackupCreated(conn *rds.RDS, arn string, timeout time.Duration) (*rds.DBInstanceAutomatedBackup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{InstanceAutomatedBackupStatusPending},
		Target:  []string{InstanceAutomatedBackupStatusReplicating},
		Refresh: statusDBInstanceAutomatedBackup(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBInstanceAutomatedBackup); ok {
		return output, err
	}

	return nil, err
}

// waitDBInstanceAutomatedBackupDeleted waits for a specified automated backup to be deleted from a database instance.
// The connection must be valid for the database instance's Region.
func waitDBInstanceAutomatedBackupDeleted(conn *rds.RDS, dbInstanceID, dbInstanceAutomatedBackupsARN string, timeout time.Duration) (*rds.DBInstance, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{strconv.FormatBool(true)},
		Target:  []string{strconv.FormatBool(false)},
		Refresh: statusDBInstanceHasAutomatedBackup(conn, dbInstanceID, dbInstanceAutomatedBackupsARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBInstance); ok {
		return output, err
	}

	return nil, err
}

func waitDBProxyCreated(conn *rds.RDS, name string, timeout time.Duration) (*rds.DBProxy, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{rds.DBProxyStatusCreating},
		Target:  []string{rds.DBProxyStatusAvailable},
		Refresh: statusDBProxy(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBProxy); ok {
		return output, err
	}

	return nil, err
}

func waitDBProxyDeleted(conn *rds.RDS, name string, timeout time.Duration) (*rds.DBProxy, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{rds.DBProxyStatusDeleting},
		Target:  []string{},
		Refresh: statusDBProxy(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBProxy); ok {
		return output, err
	}

	return nil, err
}

func waitDBProxyUpdated(conn *rds.RDS, name string, timeout time.Duration) (*rds.DBProxy, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{rds.DBProxyStatusModifying},
		Target:  []string{rds.DBProxyStatusAvailable},
		Refresh: statusDBProxy(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBProxy); ok {
		return output, err
	}

	return nil, err
}
