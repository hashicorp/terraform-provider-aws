package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a Canary to return Ready
	CanaryCreatedTimeout = 5 * time.Minute
	CanaryDeletedTimeout = 5 * time.Minute
)

// CanaryReady waits for a Canary to return Ready
func CanaryReady(conn *synthetics.Synthetics, name string) (*synthetics.GetCanaryOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{synthetics.CanaryStateCreating, synthetics.CanaryStateUpdating},
		Target:  []string{synthetics.CanaryStateReady},
		Refresh: CanaryStatus(conn, name),
		Timeout: CanaryCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*synthetics.GetCanaryOutput); ok {
		return v, err
	}

	return nil, err
}

// CanaryReady waits for a Canary to return Ready
func CanaryDeleted(conn *synthetics.Synthetics, name string) (*synthetics.GetCanaryOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{synthetics.CanaryStateDeleting},
		Target:  []string{},
		Refresh: CanaryStatus(conn, name),
		Timeout: CanaryDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*synthetics.GetCanaryOutput); ok {
		return v, err
	}

	return nil, err
}
