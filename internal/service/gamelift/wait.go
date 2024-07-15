// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

const (
	buildReadyTimeout = 1 * time.Minute
)

func waitBuildReady(ctx context.Context, conn *gamelift.Client, id string) (*awstypes.Build, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.BuildStatusInitialized),
		Target:  enum.Slice(awstypes.BuildStatusReady),
		Refresh: statusBuild(ctx, conn, id),
		Timeout: buildReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Build); ok {
		return output, err
	}

	return nil, err
}

func waitFleetActive(ctx context.Context, conn *gamelift.Client, id string, timeout time.Duration) (*awstypes.FleetAttributes, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.FleetStatusActivating,
			awstypes.FleetStatusBuilding,
			awstypes.FleetStatusDownloading,
			awstypes.FleetStatusNew,
			awstypes.FleetStatusValidating,
		),
		Target:  enum.Slice(awstypes.FleetStatusActive),
		Refresh: statusFleet(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FleetAttributes); ok {
		return output, err
	}

	return nil, err
}

func waitFleetTerminated(ctx context.Context, conn *gamelift.Client, id string, timeout time.Duration) (*awstypes.FleetAttributes, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.FleetStatusActive,
			awstypes.FleetStatusDeleting,
			awstypes.FleetStatusError,
			awstypes.FleetStatusTerminated,
		),
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

	if output, ok := outputRaw.(*awstypes.FleetAttributes); ok {
		return output, err
	}

	return nil, err
}

func getFleetFailures(ctx context.Context, conn *gamelift.Client, id string) ([]awstypes.Event, error) {
	var events []awstypes.Event
	err := _getFleetFailures(ctx, conn, id, nil, &events)
	return events, err
}

func _getFleetFailures(ctx context.Context, conn *gamelift.Client, id string, nextToken *string, events *[]awstypes.Event) error {
	eOut, err := conn.DescribeFleetEvents(ctx, &gamelift.DescribeFleetEventsInput{
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

func isEventFailure(event awstypes.Event) bool {
	failureCodes := []awstypes.EventCode{
		awstypes.EventCodeFleetActivationFailed,
		awstypes.EventCodeFleetActivationFailedNoInstances,
		awstypes.EventCodeFleetBinaryDownloadFailed,
		awstypes.EventCodeFleetInitializationFailed,
		awstypes.EventCodeFleetStateError,
		awstypes.EventCodeFleetValidationExecutableRuntimeFailure,
		awstypes.EventCodeFleetValidationLaunchPathNotFound,
		awstypes.EventCodeFleetValidationTimedOut,
		awstypes.EventCodeFleetVpcPeeringFailed,
		awstypes.EventCodeGameSessionActivationTimeout,
		awstypes.EventCodeServerProcessCrashed,
		awstypes.EventCodeServerProcessForceTerminated,
		awstypes.EventCodeServerProcessInvalidPath,
		awstypes.EventCodeServerProcessProcessExitTimeout,
		awstypes.EventCodeServerProcessProcessReadyTimeout,
		awstypes.EventCodeServerProcessSdkInitializationTimeout,
		awstypes.EventCodeServerProcessTerminatedUnhealthy,
	}
	for _, fc := range failureCodes {
		if event.EventCode == fc {
			return true
		}
	}
	return false
}

func waitGameServerGroupActive(ctx context.Context, conn *gamelift.Client, name string, timeout time.Duration) (*awstypes.GameServerGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.GameServerGroupStatusNew,
			awstypes.GameServerGroupStatusActivating,
		),
		Target:  enum.Slice(awstypes.GameServerGroupStatusActive),
		Refresh: statusGameServerGroup(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GameServerGroup); ok {
		return output, err
	}

	return nil, err
}

func waitGameServerGroupTerminated(ctx context.Context, conn *gamelift.Client, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.GameServerGroupStatusDeleteScheduled,
			awstypes.GameServerGroupStatusDeleting,
		),
		Target:  []string{},
		Refresh: statusGameServerGroup(ctx, conn, name),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting GameLift Game Server Group (%s): %w", name, err)
	}

	return nil
}
