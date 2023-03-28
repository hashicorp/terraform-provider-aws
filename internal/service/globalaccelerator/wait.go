package globalaccelerator

import (
	"time"

	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

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
