package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfrds "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds"
)

const (
	// Maximum amount of time to wait for an EventSubscription to return Deleted
	EventSubscriptionDeletedTimeout  = 10 * time.Minute
	RdsClusterInitiateUpgradeTimeout = 5 * time.Minute

	DBClusterRoleAssociationCreatedTimeout = 5 * time.Minute
	DBClusterRoleAssociationDeletedTimeout = 5 * time.Minute

	// DB Instance Automated Backups Replication timeouts
	DBInstanceAutomatedBackupsReplicationStartedTimeout = 30 * time.Minute
	DBInstanceAutomatedBackupsReplicationDeletedTimeout = 5 * time.Minute

	// DB Instance Automated Backups Replication states
	DBInstanceAutomatedBackupsPending     = "pending"
	DBInstanceAutomatedBackupsReplicating = "replicating"
	DBInstanceAutomatedBackupsDeleting    = "deleting"
)

// EventSubscriptionDeleted waits for a EventSubscription to return Deleted
func EventSubscriptionDeleted(conn *rds.RDS, subscriptionName string) (*rds.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{EventSubscriptionStatusNotFound},
		Refresh: EventSubscriptionStatus(conn, subscriptionName),
		Timeout: EventSubscriptionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*rds.EventSubscription); ok {
		return v, err
	}

	return nil, err
}

// DBProxyEndpointAvailable waits for a DBProxyEndpoint to return Available
func DBProxyEndpointAvailable(conn *rds.RDS, id string, timeout time.Duration) (*rds.DBProxyEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			rds.DBProxyEndpointStatusCreating,
			rds.DBProxyEndpointStatusModifying,
		},
		Target:  []string{rds.DBProxyEndpointStatusAvailable},
		Refresh: DBProxyEndpointStatus(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*rds.DBProxyEndpoint); ok {
		return v, err
	}

	return nil, err
}

// DBProxyEndpointDeleted waits for a DBProxyEndpoint to return Deleted
func DBProxyEndpointDeleted(conn *rds.RDS, id string, timeout time.Duration) (*rds.DBProxyEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{rds.DBProxyEndpointStatusDeleting},
		Target:  []string{},
		Refresh: DBProxyEndpointStatus(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*rds.DBProxyEndpoint); ok {
		return v, err
	}

	return nil, err
}

func DBClusterRoleAssociationCreated(conn *rds.RDS, dbClusterID, roleARN string) (*rds.DBClusterRole, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{tfrds.DBClusterRoleStatusPending},
		Target:  []string{tfrds.DBClusterRoleStatusActive},
		Refresh: DBClusterRoleStatus(conn, dbClusterID, roleARN),
		Timeout: DBClusterRoleAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBClusterRole); ok {
		return output, err
	}

	return nil, err
}

func DBClusterRoleAssociationDeleted(conn *rds.RDS, dbClusterID, roleARN string) (*rds.DBClusterRole, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{tfrds.DBClusterRoleStatusActive, tfrds.DBClusterRoleStatusPending},
		Target:  []string{},
		Refresh: DBClusterRoleStatus(conn, dbClusterID, roleARN),
		Timeout: DBClusterRoleAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBClusterRole); ok {
		return output, err
	}

	return nil, err
}

// DBInstanceAutomatedBackupsReplicationStarted waits for a DBInstanceAutomatedBackup to return replicating
func DBInstanceAutomatedBackupsReplicationStarted(ctx context.Context, conn *rds.RDS, dbInstanceAutomatedBackupsArn string) (*rds.DBInstanceAutomatedBackup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{DBInstanceAutomatedBackupsPending},
		Target:  []string{DBInstanceAutomatedBackupsReplicating},
		Refresh: DBInstanceAutomatedBackupsReplicationStatus(ctx, conn, dbInstanceAutomatedBackupsArn),
		Timeout: DBInstanceAutomatedBackupsReplicationStartedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBInstanceAutomatedBackup); ok {
		return output, err
	}

	return nil, err
}

// DBInstanceAutomatedBackupsReplicationDeleted waits for a DBInstanceAutomatedBackup to return deleting
func DBInstanceAutomatedBackupsReplicationDeleted(ctx context.Context, conn *rds.RDS, dbInstanceAutomatedBackupsArn string) (*rds.DBInstanceAutomatedBackup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{DBInstanceAutomatedBackupsReplicating},
		Target:  []string{DBInstanceAutomatedBackupsDeleting},
		Refresh: DBInstanceAutomatedBackupsReplicationStatus(ctx, conn, dbInstanceAutomatedBackupsArn),
		Timeout: DBInstanceAutomatedBackupsReplicationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*rds.DBInstanceAutomatedBackup); ok {
		return output, err
	}

	return nil, err
}
