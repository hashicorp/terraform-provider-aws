package cloud9

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	EnvironmentReadyTimeout   = 10 * time.Minute
	EnvironmentDeletedTimeout = 20 * time.Minute
)

func waitEnvironmentReady(ctx context.Context, conn *cloud9.Cloud9, id string) (*cloud9.Environment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloud9.EnvironmentLifecycleStatusCreating},
		Target:  []string{cloud9.EnvironmentLifecycleStatusCreated},
		Refresh: statusEnvironmentStatus(ctx, conn, id),
		Timeout: EnvironmentReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloud9.Environment); ok {
		if lifecycle := output.Lifecycle; aws.StringValue(lifecycle.Status) == cloud9.EnvironmentLifecycleStatusCreateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(lifecycle.Reason)))
		}

		return output, err
	}

	return nil, err
}

func waitEnvironmentDeleted(ctx context.Context, conn *cloud9.Cloud9, id string) (*cloud9.Environment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloud9.EnvironmentLifecycleStatusDeleting},
		Target:  []string{},
		Refresh: statusEnvironmentStatus(ctx, conn, id),
		Timeout: EnvironmentDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloud9.Environment); ok {
		if lifecycle := output.Lifecycle; aws.StringValue(lifecycle.Status) == cloud9.EnvironmentLifecycleStatusDeleteFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(lifecycle.Reason)))
		}

		return output, err
	}

	return nil, err
}
