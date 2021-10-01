package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfrds "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds"
)

const (
	RdsClusterInitiateUpgradeTimeout = 5 * time.Minute

	DBClusterRoleAssociationCreatedTimeout = 5 * time.Minute
	DBClusterRoleAssociationDeletedTimeout = 5 * time.Minute
)

func EventSubscriptionCreated(conn *rds.RDS, id string, timeout time.Duration) (*rds.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{tfrds.EventSubscriptionStatusCreating},
		Target:     []string{tfrds.EventSubscriptionStatusActive},
		Refresh:    EventSubscriptionStatus(conn, id),
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

func EventSubscriptionDeleted(conn *rds.RDS, id string, timeout time.Duration) (*rds.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{tfrds.EventSubscriptionStatusDeleting},
		Target:     []string{},
		Refresh:    EventSubscriptionStatus(conn, id),
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

func EventSubscriptionUpdated(conn *rds.RDS, id string, timeout time.Duration) (*rds.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{tfrds.EventSubscriptionStatusModifying},
		Target:     []string{tfrds.EventSubscriptionStatusActive},
		Refresh:    EventSubscriptionStatus(conn, id),
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

	if output, ok := outputRaw.(*rds.DBProxyEndpoint); ok {
		return output, err
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

	if output, ok := outputRaw.(*rds.DBProxyEndpoint); ok {
		return output, err
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
