package gamelift

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	buildReadyTimeout = 1 * time.Minute
)

func waitBuildReady(ctx context.Context, conn *gamelift.GameLift, id string) (*gamelift.Build, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{gamelift.BuildStatusInitialized},
		Target:  []string{gamelift.BuildStatusReady},
		Refresh: statusBuild(ctx, conn, id),
		Timeout: buildReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*gamelift.Build); ok {
		return output, err
	}

	return nil, err
}

func waitFleetActive(ctx context.Context, conn *gamelift.GameLift, id string, timeout time.Duration) (*gamelift.FleetAttributes, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			gamelift.FleetStatusActivating,
			gamelift.FleetStatusBuilding,
			gamelift.FleetStatusDownloading,
			gamelift.FleetStatusNew,
			gamelift.FleetStatusValidating,
		},
		Target:  []string{gamelift.FleetStatusActive},
		Refresh: statusFleet(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*gamelift.FleetAttributes); ok {
		return output, err
	}

	return nil, err
}

func waitFleetTerminated(ctx context.Context, conn *gamelift.GameLift, id string, timeout time.Duration) (*gamelift.FleetAttributes, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			gamelift.FleetStatusActive,
			gamelift.FleetStatusDeleting,
			gamelift.FleetStatusError,
			gamelift.FleetStatusTerminated,
		},
		Target:  []string{},
		Refresh: statusFleet(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if err != nil {
		events, fErr := getFleetFailures(ctx, conn, id)
		if fErr != nil {
			log.Printf("[WARN] Failed to poll fleet failures: %s", fErr)
		}
		if len(events) > 0 {
			return nil, fmt.Errorf("%s Recent failures:\n%+v", err, events)
		}
	}

	if output, ok := outputRaw.(*gamelift.FleetAttributes); ok {
		return output, err
	}

	return nil, err
}

func getFleetFailures(ctx context.Context, conn *gamelift.GameLift, id string) ([]*gamelift.Event, error) {
	var events []*gamelift.Event
	err := _getFleetFailures(ctx, conn, id, nil, &events)
	return events, err
}

func _getFleetFailures(ctx context.Context, conn *gamelift.GameLift, id string, nextToken *string, events *[]*gamelift.Event) error {
	eOut, err := conn.DescribeFleetEventsWithContext(ctx, &gamelift.DescribeFleetEventsInput{
		FleetId:   aws.String(id),
		NextToken: nextToken,
	})
	if err != nil {
		return err
	}

	for _, e := range eOut.Events {
		if isEventFailure(e) {
			*events = append(*events, e)
		}
	}

	if eOut.NextToken != nil {
		err := _getFleetFailures(ctx, conn, id, nextToken, events)
		if err != nil {
			return err
		}
	}

	return nil
}

func isEventFailure(event *gamelift.Event) bool {
	failureCodes := []string{
		gamelift.EventCodeFleetActivationFailed,
		gamelift.EventCodeFleetActivationFailedNoInstances,
		gamelift.EventCodeFleetBinaryDownloadFailed,
		gamelift.EventCodeFleetInitializationFailed,
		gamelift.EventCodeFleetStateError,
		gamelift.EventCodeFleetValidationExecutableRuntimeFailure,
		gamelift.EventCodeFleetValidationLaunchPathNotFound,
		gamelift.EventCodeFleetValidationTimedOut,
		gamelift.EventCodeFleetVpcPeeringFailed,
		gamelift.EventCodeGameSessionActivationTimeout,
		gamelift.EventCodeServerProcessCrashed,
		gamelift.EventCodeServerProcessForceTerminated,
		gamelift.EventCodeServerProcessInvalidPath,
		gamelift.EventCodeServerProcessProcessExitTimeout,
		gamelift.EventCodeServerProcessProcessReadyTimeout,
		gamelift.EventCodeServerProcessSdkInitializationTimeout,
		gamelift.EventCodeServerProcessTerminatedUnhealthy,
	}
	for _, fc := range failureCodes {
		if aws.StringValue(event.EventCode) == fc {
			return true
		}
	}
	return false
}

func waitGameServerGroupActive(ctx context.Context, conn *gamelift.GameLift, name string, timeout time.Duration) (*gamelift.GameServerGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			gamelift.GameServerGroupStatusNew,
			gamelift.GameServerGroupStatusActivating,
		},
		Target:  []string{gamelift.GameServerGroupStatusActive},
		Refresh: statusGameServerGroup(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*gamelift.GameServerGroup); ok {
		return output, err
	}

	return nil, err
}

func waitGameServerGroupTerminated(ctx context.Context, conn *gamelift.GameLift, name string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			gamelift.GameServerGroupStatusDeleteScheduled,
			gamelift.GameServerGroupStatusDeleting,
		},
		Target:  []string{},
		Refresh: statusGameServerGroup(ctx, conn, name),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting GameLift Game Server Group (%s): %w", name, err)
	}

	return nil
}
