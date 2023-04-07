package pipes

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/pipes"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	pipeStatusRunning      = string(types.PipeStateRunning)
	pipeStatusStopped      = string(types.PipeStateStopped)
	pipeStatusCreating     = string(types.PipeStateCreating)
	pipeStatusUpdating     = string(types.PipeStateUpdating)
	pipeStatusDeleting     = string(types.PipeStateDeleting)
	pipeStatusStarting     = string(types.PipeStateStarting)
	pipeStatusStopping     = string(types.PipeStateStopping)
	pipeStatusCreateFailed = string(types.PipeStateCreateFailed)
	pipeStatusUpdateFailed = string(types.PipeStateUpdateFailed)
	pipeStatusStartFailed  = string(types.PipeStateStartFailed)
	pipeStatusStopFailed   = string(types.PipeStateStopFailed)
)

func statusPipe(ctx context.Context, conn *pipes.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindPipeByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.CurrentState), nil
	}
}
