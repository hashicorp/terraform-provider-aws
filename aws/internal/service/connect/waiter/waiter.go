package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfconnect "github.com/hashicorp/terraform-provider-aws/aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// ConnectInstanceCreateTimeout Timeout for connect instance creation
	ConnectInstanceCreatedTimeout = 5 * time.Minute
	ConnectInstanceDeletedTimeout = 5 * time.Minute

	ConnectContactFlowCreateTimeout = 5 * time.Minute
	ConnectContactFlowUpdateTimeout = 5 * time.Minute
)

func InstanceCreated(ctx context.Context, conn *connect.Connect, instanceId string) (*connect.DescribeInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{connect.InstanceStatusCreationInProgress},
		Target:  []string{connect.InstanceStatusActive},
		Refresh: InstanceStatus(ctx, conn, instanceId),
		Timeout: ConnectInstanceCreatedTimeout,
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
func InstanceDeleted(ctx context.Context, conn *connect.Connect, instanceId string) (*connect.DescribeInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{connect.InstanceStatusActive},
		Target:  []string{tfconnect.InstanceStatusStatusNotFound},
		Refresh: InstanceStatus(ctx, conn, instanceId),
		Timeout: ConnectInstanceDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*connect.DescribeInstanceOutput); ok {
		return v, err
	}

	return nil, err
}
