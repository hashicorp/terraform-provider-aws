package gamelift

import (
	"time"

	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	buildReadyTimeout = 1 * time.Minute
)

func waitBuildReady(conn *gamelift.GameLift, id string) (*gamelift.Build, error) {
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

func waitFleetActive(conn *gamelift.GameLift, id string, timeout time.Duration) (*gamelift.FleetAttributes, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			gamelift.FleetStatusActivating,
			gamelift.FleetStatusBuilding,
			gamelift.FleetStatusDownloading,
			gamelift.FleetStatusNew,
			gamelift.FleetStatusValidating,
		},
		Target:  []string{gamelift.FleetStatusActive},
		Refresh: statusFleet(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*gamelift.FleetAttributes); ok {
		return output, err
	}

	return nil, err
}

// func waitFleetTerminated(conn *gamelift.GameLift, id string, timeout time.Duration) (*gamelift.FleetAttributes, error) { //nolint:unparam
// 	stateConf := &resource.StateChangeConf{
// 		Pending: []string{
// 			gamelift.FleetStatusActive,
// 			gamelift.FleetStatusDeleting,
// 			gamelift.FleetStatusError,
// 			gamelift.FleetStatusTerminated,
// 		},
// 		Target:  []string{},
// 		Refresh: statusFleet(conn, id),
// 		Timeout: timeout,
// 	}

// 	outputRaw, err := stateConf.WaitForState()

// 	if output, ok := outputRaw.(*gamelift.FleetAttributes); ok {
// 		return output, err
// 	}

// 	return nil, err
// }
