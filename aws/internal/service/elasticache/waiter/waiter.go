package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	replicationGroupAvailableMinTimeout = 10 * time.Second
	replicationGroupAvailableDelay      = 30 * time.Second

	replicationGroupDeletedMinTimeout = 10 * time.Second
	replicationGroupDeletedDelay      = 30 * time.Second
)

// ReplicationGroupAvailable waits for a ReplicationGroup to return Available
func ReplicationGroupAvailable(conn *elasticache.ElastiCache, replicationGroupID string, timeout time.Duration) (*elasticache.ReplicationGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ReplicationGroupStatusCreating, ReplicationGroupStatusModifying, ReplicationGroupStatusSnapshotting},
		Target:     []string{ReplicationGroupStatusAvailable},
		Refresh:    ReplicationGroupStatus(conn, replicationGroupID),
		Timeout:    timeout,
		MinTimeout: replicationGroupAvailableMinTimeout,
		Delay:      replicationGroupAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*elasticache.ReplicationGroup); ok {
		return v, err
	}
	return nil, err
}

// ReplicationGroupDeleted waits for a ReplicationGroup to be deleted
func ReplicationGroupDeleted(conn *elasticache.ElastiCache, replicationGroupID string, timeout time.Duration) (*elasticache.ReplicationGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ReplicationGroupStatusCreating, ReplicationGroupStatusAvailable, ReplicationGroupStatusDeleting},
		Target:     []string{},
		Refresh:    ReplicationGroupStatus(conn, replicationGroupID),
		Timeout:    timeout,
		MinTimeout: replicationGroupDeletedMinTimeout,
		Delay:      replicationGroupDeletedDelay,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*elasticache.ReplicationGroup); ok {
		return v, err
	}
	return nil, err
}
