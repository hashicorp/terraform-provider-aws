// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ivs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ivs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

func waitPlaybackKeyPairCreated(ctx context.Context, conn *ivs.Client, id string, timeout time.Duration) (*awstypes.PlaybackKeyPair, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusPlaybackKeyPair(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.PlaybackKeyPair); ok {
		return out, err
	}

	return nil, err
}

func waitPlaybackKeyPairDeleted(ctx context.Context, conn *ivs.Client, id string, timeout time.Duration) (*awstypes.PlaybackKeyPair, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusNormal},
		Target:  []string{},
		Refresh: statusPlaybackKeyPair(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.PlaybackKeyPair); ok {
		return out, err
	}

	return nil, err
}

func waitRecordingConfigurationCreated(ctx context.Context, conn *ivs.Client, id string, timeout time.Duration) (*awstypes.RecordingConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RecordingConfigurationStateCreating),
		Target:                    enum.Slice(awstypes.RecordingConfigurationStateActive),
		Refresh:                   statusRecordingConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.RecordingConfiguration); ok {
		return out, err
	}

	return nil, err
}

func waitRecordingConfigurationDeleted(ctx context.Context, conn *ivs.Client, id string, timeout time.Duration) (*awstypes.RecordingConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RecordingConfigurationStateActive),
		Target:  []string{},
		Refresh: statusRecordingConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.RecordingConfiguration); ok {
		return out, err
	}

	return nil, err
}

func waitChannelCreated(ctx context.Context, conn *ivs.Client, id string, timeout time.Duration) (*awstypes.Channel, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusChannel(ctx, conn, id, nil),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Channel); ok {
		return out, err
	}

	return nil, err
}

func waitChannelUpdated(ctx context.Context, conn *ivs.Client, id string, timeout time.Duration, updateDetails *ivs.UpdateChannelInput) (*awstypes.Channel, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusChannel(ctx, conn, id, updateDetails),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Channel); ok {
		return out, err
	}

	return nil, err
}

func waitChannelDeleted(ctx context.Context, conn *ivs.Client, id string, timeout time.Duration) (*awstypes.Channel, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusNormal},
		Target:  []string{},
		Refresh: statusChannel(ctx, conn, id, nil),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.Channel); ok {
		return out, err
	}

	return nil, err
}
