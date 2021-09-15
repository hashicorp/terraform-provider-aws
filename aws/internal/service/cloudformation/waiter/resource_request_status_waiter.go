package waiter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func ResourceRequestStatusProgressEventOperationStatusSuccess(ctx context.Context, conn *cloudformation.CloudFormation, requestToken string, timeout time.Duration) (*cloudformation.ProgressEvent, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			cloudformation.OperationStatusInProgress,
			cloudformation.OperationStatusPending,
		},
		Target:  []string{cloudformation.OperationStatusSuccess},
		Refresh: ResourceRequestStatusProgressEventOperationStatus(ctx, conn, requestToken),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudformation.ProgressEvent); ok {
		if err != nil && output != nil {
			newErr := fmt.Errorf("%s", output)

			var te *resource.TimeoutError
			var use *resource.UnexpectedStateError
			if ok := errors.As(err, &te); ok && te.LastError == nil {
				te.LastError = newErr
			} else if ok := errors.As(err, &use); ok && use.LastError == nil {
				use.LastError = newErr
			}
		}

		return output, err
	}

	return nil, err
}
