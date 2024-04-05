// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func waitPlaybackKeyPairCreated(ctx context.Context, conn *ivs.IVS, id string, timeout time.Duration) (*ivs.PlaybackKeyPair, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusPlaybackKeyPair(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivs.PlaybackKeyPair); ok {
		return out, err
	}

	return nil, err
}

func waitPlaybackKeyPairDeleted(ctx context.Context, conn *ivs.IVS, id string, timeout time.Duration) (*ivs.PlaybackKeyPair, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusNormal},
		Target:  []string{},
		Refresh: statusPlaybackKeyPair(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivs.PlaybackKeyPair); ok {
		return out, err
	}

	return nil, err
}

func waitRecordingConfigurationCreated(ctx context.Context, conn *ivs.IVS, id string, timeout time.Duration) (*ivs.RecordingConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{ivs.RecordingConfigurationStateCreating},
		Target:                    []string{ivs.RecordingConfigurationStateActive},
		Refresh:                   statusRecordingConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivs.RecordingConfiguration); ok {
		return out, err
	}

	return nil, err
}

func waitRecordingConfigurationDeleted(ctx context.Context, conn *ivs.IVS, id string, timeout time.Duration) (*ivs.RecordingConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ivs.RecordingConfigurationStateActive},
		Target:  []string{},
		Refresh: statusRecordingConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivs.RecordingConfiguration); ok {
		return out, err
	}

	return nil, err
}

func waitChannelCreated(ctx context.Context, conn *ivs.IVS, id string, timeout time.Duration) (*ivs.Channel, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusChannel(ctx, conn, id, nil),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivs.Channel); ok {
		return out, err
	}

	return nil, err
}

func waitChannelUpdated(ctx context.Context, conn *ivs.IVS, id string, timeout time.Duration, updateDetails *ivs.UpdateChannelInput) (*ivs.Channel, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusChannel(ctx, conn, id, updateDetails),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivs.Channel); ok {
		return out, err
	}

	return nil, err
}

func waitChannelDeleted(ctx context.Context, conn *ivs.IVS, id string, timeout time.Duration) (*ivs.Channel, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusNormal},
		Target:  []string{},
		Refresh: statusChannel(ctx, conn, id, nil),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivs.Channel); ok {
		return out, err
	}

	return nil, err
}
