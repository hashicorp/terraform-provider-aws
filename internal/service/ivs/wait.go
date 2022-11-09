package ivs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func waitPlaybackKeyPairCreated(ctx context.Context, conn *ivs.IVS, id string, timeout time.Duration) (*ivs.PlaybackKeyPair, error) {
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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
