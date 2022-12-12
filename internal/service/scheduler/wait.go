package scheduler

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func waitScheduleGroupActive(ctx context.Context, conn *scheduler.Client, name string, timeout time.Duration) (*scheduler.GetScheduleGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{scheduleGroupStatusActive},
		Refresh:                   statusScheduleGroup(ctx, conn, name),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*scheduler.GetScheduleGroupOutput); ok {
		return out, err
	}

	return nil, err
}

func waitScheduleGroupDeleted(ctx context.Context, conn *scheduler.Client, name string, timeout time.Duration) (*scheduler.GetScheduleGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{scheduleGroupStatusDeleting, scheduleGroupStatusActive},
		Target:  []string{},
		Refresh: statusScheduleGroup(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*scheduler.GetScheduleGroupOutput); ok {
		return out, err
	}

	return nil, err
}
