package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// waitAcceleratorDeployed waits for an Accelerator to return Deployed
func waitAcceleratorDeployed(conn *globalaccelerator.GlobalAccelerator, arn string, timeout time.Duration) (*globalaccelerator.Accelerator, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{globalaccelerator.AcceleratorStatusInProgress},
		Target:  []string{globalaccelerator.AcceleratorStatusDeployed},
		Refresh: statusAccelerator(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*globalaccelerator.Accelerator); ok {
		return v, err
	}

	return nil, err
}
