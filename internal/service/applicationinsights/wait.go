package applicationinsights

import (
	"time"

	"github.com/aws/aws-sdk-go/service/applicationinsights"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ApplicationCreatedTimeout = 2 * time.Minute

	ApplicationDeletedTimeout = 2 * time.Minute
)

func waitApplicationCreated(conn *applicationinsights.ApplicationInsights, name string) (*applicationinsights.ApplicationInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"CREATING"},
		Target:  []string{"NOT_CONFIGURED"},
		Refresh: statusApplication(conn, name),
		Timeout: ApplicationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()
	if output, ok := outputRaw.(*applicationinsights.ApplicationInfo); ok {
		return output, err
	}

	return nil, err
}

func waitApplicationTerminated(conn *applicationinsights.ApplicationInsights, name string) (*applicationinsights.ApplicationInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"NOT_CONFIGURED", "DELETING"},
		Target:  []string{},
		Refresh: statusApplication(conn, name),
		Timeout: ApplicationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()
	if output, ok := outputRaw.(*applicationinsights.ApplicationInfo); ok {
		return output, err
	}

	return nil, err
}
