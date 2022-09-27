package controltower

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/controltower"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	resourceStatusFailed     = "FAILED"
	resourceStatusInProgress = "IN_PROGRESS"
	resourceStatusSucceeded  = "SUCCEEDED"
)

func statusControlCreated(ctx context.Context, conn *controltower.ControlTower, operation_identifier string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &controltower.GetControlOperationInput{
			OperationIdentifier: aws.String(operation_identifier),
		}

		output, err := conn.GetControlOperationWithContext(ctx, input)

		if err != nil {
			return output, resourceStatusFailed, err
		}

		return output.ControlOperation, *output.ControlOperation.Status, nil
	}
}
