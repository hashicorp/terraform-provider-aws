package globalaccelerator

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func waitAcceleratorDeployed(ctx context.Context, conn *globalaccelerator.GlobalAccelerator, arn string, timeout time.Duration) (*globalaccelerator.Accelerator, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{globalaccelerator.AcceleratorStatusInProgress},
		Target:  []string{globalaccelerator.AcceleratorStatusDeployed},
		Refresh: statusAccelerator(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*globalaccelerator.Accelerator); ok {
		return output, err
	}

	return nil, err
}

// waitCustomRoutingAcceleratorDeployed waits for a Custom Routing Accelerator to return Deployed
func waitCustomRoutingAcceleratorDeployed(conn *globalaccelerator.GlobalAccelerator, arn string, timeout time.Duration) (*globalaccelerator.CustomRoutingAccelerator, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{globalaccelerator.AcceleratorStatusInProgress},
		Target:  []string{globalaccelerator.AcceleratorStatusDeployed},
		Refresh: statusCustomRoutingAccelerator(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*globalaccelerator.CustomRoutingAccelerator); ok {
		return v, err
	}

	return nil, err
}
