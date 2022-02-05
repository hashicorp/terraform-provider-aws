package gamelift

import (
	"time"

	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	buildReadyTimeout = 1 * time.Minute
)

func waitBuildReady(conn *gamelift.GameLift, id string) (*gamelift.Build, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{gamelift.BuildStatusInitialized},
		Target:  []string{gamelift.BuildStatusReady},
		Refresh: statusBuild(conn, id),
		Timeout: buildReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*gamelift.Build); ok {
		return output, err
	}

	return nil, err
}
