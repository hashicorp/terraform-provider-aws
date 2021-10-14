package connect

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// ConnectInstanceCreateTimeout Timeout for connect instance creation
	connectInstanceCreatedTimeout = 5 * time.Minute
	connectInstanceDeletedTimeout = 5 * time.Minute

	connectContactFlowCreateTimeout = 5 * time.Minute
	connectContactFlowUpdateTimeout = 5 * time.Minute
)

func waitInstanceCreated(ctx context.Context, conn *connect.Connect, instanceId string) (*connect.DescribeInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{connect.InstanceStatusCreationInProgress},
		Target:  []string{connect.InstanceStatusActive},
		Refresh: statusInstance(ctx, conn, instanceId),
		Timeout: connectInstanceCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*connect.DescribeInstanceOutput); ok {
		return v, err
	}

	return nil, err
}

// We don't have a PENDING_DELETION or DELETED for the Connect instance.
// If the Connect Instance has an associated EXISTING DIRECTORY, removing the connect instance
// will cause an error because it is still has authorized applications.
func waitInstanceDeleted(ctx context.Context, conn *connect.Connect, instanceId string) (*connect.DescribeInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{connect.InstanceStatusActive},
		Target:  []string{InstanceStatusStatusNotFound},
		Refresh: statusInstance(ctx, conn, instanceId),
		Timeout: connectInstanceDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*connect.DescribeInstanceOutput); ok {
		return v, err
	}

	return nil, err
}
