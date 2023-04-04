package pipes

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func waitPipeCreated(ctx context.Context, conn *pipes.Client, id string, timeout time.Duration) (*pipes.DescribePipeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{pipeStatusCreating},
		Target:                    []string{pipeStatusRunning, pipeStatusStopped},
		Refresh:                   statusPipe(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*pipes.DescribePipeOutput); ok {
		if reason := aws.ToString(out.StateReason); reason != "" && err != nil {
			err = fmt.Errorf("%s: %s", err, reason)
		}

		return out, err
	}

	return nil, err
}

func waitPipeUpdated(ctx context.Context, conn *pipes.Client, id string, timeout time.Duration) (*pipes.DescribePipeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{pipeStatusUpdating},
		Target:                    []string{pipeStatusRunning, pipeStatusStopped},
		Refresh:                   statusPipe(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*pipes.DescribePipeOutput); ok {
		if reason := aws.ToString(out.StateReason); reason != "" && err != nil {
			err = fmt.Errorf("%s: %s", err, reason)
		}

		return out, err
	}

	return nil, err
}

func waitPipeDeleted(ctx context.Context, conn *pipes.Client, id string, timeout time.Duration) (*pipes.DescribePipeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{pipeStatusDeleting},
		Target:  []string{},
		Refresh: statusPipe(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*pipes.DescribePipeOutput); ok {
		if reason := aws.ToString(out.StateReason); reason != "" && err != nil {
			err = fmt.Errorf("%s: %s", err, reason)
		}

		return out, err
	}

	return nil, err
}
