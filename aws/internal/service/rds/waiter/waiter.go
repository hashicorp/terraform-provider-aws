package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfrds "github.com/hashicorp/terraform-provider-aws/aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
)

const (
	rdsClusterInitiateUpgradeTimeout = 5 * time.Minute

	dbClusterRoleAssociationCreatedTimeout = 5 * time.Minute
	dbClusterRoleAssociationDeletedTimeout = 5 * time.Minute
)

func waitEventSubscriptionCreated(conn *rds.RDS, id string, timeout time.Duration) (*rds.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{tfrds.EventSubscriptionStatusCreating},
		Target:     []string{tfrds.EventSubscriptionStatusActive},
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
		Pending:    []string{tfrds.EventSubscriptionStatusDeleting},
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
		Pending:    []string{tfrds.EventSubscriptionStatusModifying},
		Target:     []string{tfrds.EventSubscriptionStatusActive},
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

// waitDBProxyEndpointAvailable waits for a DBProxyEndpoint to return Available
func waitDBProxyEndpointAvailable(conn *rds.RDS, id string, timeout time.Duration) (*rds.DBProxyEndpoint, error) {
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
		Pending: []string{tfrds.ClusterRoleStatusPending},
		Target:  []string{tfrds.ClusterRoleStatusActive},
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
		Pending: []string{tfrds.ClusterRoleStatusActive, tfrds.ClusterRoleStatusPending},
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
			tfrds.InstanceStatusAvailable,
			tfrds.InstanceStatusBackingUp,
			tfrds.InstanceStatusConfiguringEnhancedMonitoring,
			tfrds.InstanceStatusConfiguringLogExports,
			tfrds.InstanceStatusCreating,
			tfrds.InstanceStatusDeleting,
			tfrds.InstanceStatusIncompatibleParameters,
			tfrds.InstanceStatusModifying,
			tfrds.InstanceStatusStarting,
			tfrds.InstanceStatusStopping,
			tfrds.InstanceStatusStorageFull,
			tfrds.InstanceStatusStorageOptimization,
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

func waitDBClusterInstanceDeleted(conn *rds.RDS, id string, timeout time.Duration) (*rds.DBInstance, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			tfrds.InstanceStatusConfiguringLogExports,
			tfrds.InstanceStatusDeleting,
			tfrds.InstanceStatusModifying,
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
