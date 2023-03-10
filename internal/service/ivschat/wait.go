package ivschat

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ivschat"
	"github.com/aws/aws-sdk-go-v2/service/ivschat/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

func waitLoggingConfigurationCreated(ctx context.Context, conn *ivschat.Client, id string, timeout time.Duration) (*ivschat.GetLoggingConfigurationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   enum.Slice(types.LoggingConfigurationStateCreating),
		Target:                    enum.Slice(types.LoggingConfigurationStateActive),
		Refresh:                   statusLoggingConfiguration(ctx, conn, id),
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
	stateConf := &resource.StateChangeConf{
		Pending:                   enum.Slice(types.LoggingConfigurationStateUpdating),
		Target:                    enum.Slice(types.LoggingConfigurationStateActive),
		Refresh:                   statusLoggingConfiguration(ctx, conn, id),
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
	stateConf := &resource.StateChangeConf{
		Pending: enum.Slice(types.LoggingConfigurationStateDeleting, types.LoggingConfigurationStateActive),
		Target:  []string{},
		Refresh: statusLoggingConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivschat.GetLoggingConfigurationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitRoomUpdated(ctx context.Context, conn *ivschat.Client, id string, timeout time.Duration, updateDetails *ivschat.UpdateRoomInput) (*ivschat.GetRoomOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusRoom(ctx, conn, id, updateDetails),
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
	stateConf := &resource.StateChangeConf{
		Pending: []string{statusNormal},
		Target:  []string{},
		Refresh: statusRoom(ctx, conn, id, nil),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*ivschat.GetRoomOutput); ok {
		return out, err
	}

	return nil, err
}

func waitRoomCreated(ctx context.Context, conn *ivschat.Client, id string, timeout time.Duration) (*ivschat.GetRoomOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusRoom(ctx, conn, id, nil),
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
