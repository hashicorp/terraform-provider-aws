// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ivschat

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ivschat"
	"github.com/aws/aws-sdk-go-v2/service/ivschat/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func waitLoggingConfigurationCreated(ctx context.Context, conn *ivschat.Client, id string, timeout time.Duration) (*ivschat.GetLoggingConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.LoggingConfigurationStateCreating),
		Target:                    enum.Slice(types.LoggingConfigurationStateActive),
		Refresh:                   statusLoggingConfiguration(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivschat.GetLoggingConfigurationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitLoggingConfigurationUpdated(ctx context.Context, conn *ivschat.Client, id string, timeout time.Duration) (*ivschat.GetLoggingConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.LoggingConfigurationStateUpdating),
		Target:                    enum.Slice(types.LoggingConfigurationStateActive),
		Refresh:                   statusLoggingConfiguration(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivschat.GetLoggingConfigurationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitLoggingConfigurationDeleted(ctx context.Context, conn *ivschat.Client, id string, timeout time.Duration) (*ivschat.GetLoggingConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.LoggingConfigurationStateDeleting, types.LoggingConfigurationStateActive),
		Target:  []string{},
		Refresh: statusLoggingConfiguration(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivschat.GetLoggingConfigurationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitRoomUpdated(ctx context.Context, conn *ivschat.Client, id string, timeout time.Duration, updateDetails *ivschat.UpdateRoomInput) (*ivschat.GetRoomOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusRoom(conn, id, updateDetails),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivschat.GetRoomOutput); ok {
		return out, err
	}

	return nil, err
}

func waitRoomDeleted(ctx context.Context, conn *ivschat.Client, id string, timeout time.Duration) (*ivschat.GetRoomOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusNormal},
		Target:  []string{},
		Refresh: statusRoom(conn, id, nil),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivschat.GetRoomOutput); ok {
		return out, err
	}

	return nil, err
}

func waitRoomCreated(ctx context.Context, conn *ivschat.Client, id string, timeout time.Duration) (*ivschat.GetRoomOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusRoom(conn, id, nil),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivschat.GetRoomOutput); ok {
		return out, err
	}

	return nil, err
}
