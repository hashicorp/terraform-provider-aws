package emrserverless

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrserverless"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	ApplicationCreatedTimeout    = 75 * time.Minute
	ApplicationCreatedMinTimeout = 10 * time.Second
	ApplicationCreatedDelay      = 30 * time.Second

	ApplicationDeletedTimeout    = 20 * time.Minute
	ApplicationDeletedMinTimeout = 10 * time.Second
	ApplicationDeletedDelay      = 30 * time.Second
)

func waitApplicationCreated(ctx context.Context, conn *emrserverless.EMRServerless, id string) (*emrserverless.Application, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{emrserverless.ApplicationStateCreating},
		Target:     []string{emrserverless.ApplicationStateCreated},
		Refresh:    statusApplication(ctx, conn, id),
		Timeout:    ApplicationCreatedTimeout,
		MinTimeout: ApplicationCreatedMinTimeout,
		Delay:      ApplicationCreatedDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*emrserverless.Application); ok {
		if stateChangeReason := output.StateDetails; stateChangeReason != nil {
			tfresource.SetLastError(err, fmt.Errorf(aws.StringValue(stateChangeReason)))
		}

		return output, err
	}

	return nil, err
}

func waitApplicationTerminated(ctx context.Context, conn *emrserverless.EMRServerless, id string) (*emrserverless.Application, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    emrserverless.ApplicationState_Values(),
		Target:     []string{},
		Refresh:    statusApplication(ctx, conn, id),
		Timeout:    ApplicationDeletedTimeout,
		MinTimeout: ApplicationDeletedMinTimeout,
		Delay:      ApplicationDeletedDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*emrserverless.Application); ok {
		if stateChangeReason := output.StateDetails; stateChangeReason != nil {
			tfresource.SetLastError(err, fmt.Errorf(aws.StringValue(stateChangeReason)))
		}

		return output, err
	}

	return nil, err
}
