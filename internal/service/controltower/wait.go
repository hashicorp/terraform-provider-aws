package controltower

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/controltower"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a control to be created, updated, or deleted
	controlTimeout = 5 * time.Minute
)

func waitControlCreated(ctx context.Context, conn *controltower.ControlTower, operation_identifier string) (*controltower.ControlOperation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{controltower.ControlOperationStatusInProgress},
		Target:  []string{controltower.ControlOperationStatusSucceeded},
		Refresh: statusControlCreated(ctx, conn, operation_identifier),
		Timeout: controlTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*controltower.ControlOperation); ok {
		return v, err
	}

	return nil, err
}
