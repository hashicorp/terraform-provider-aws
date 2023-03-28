package synthetics

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	canaryCreateFail = "CREATE_FAILED"
)

func retryCreateCanary(ctx context.Context, conn *synthetics.Synthetics, d *schema.ResourceData, input *synthetics.CreateCanaryInput) (*synthetics.Canary, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{synthetics.CanaryStateCreating, synthetics.CanaryStateUpdating},
		Target:  []string{synthetics.CanaryStateReady},
		Refresh: statusCanaryState(ctx, conn, d.Id()),
		Timeout: canaryCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*synthetics.Canary); ok {
		if status := output.Status; aws.StringValue(status.State) == synthetics.CanaryStateError && aws.StringValue(status.StateReasonCode) == canaryCreateFail {
			// delete canary because it is the only way to reprovision if in an error state
			err = deleteCanary(ctx, conn, d.Id())
			if err != nil {
				return output, fmt.Errorf("error deleting Synthetics Canary on retry (%s): %w", d.Id(), err)
			}

			_, err = conn.CreateCanaryWithContext(ctx, input)
			if err != nil {
				return output, fmt.Errorf("error creating Synthetics Canary on retry (%s): %w", d.Id(), err)
			}

			_, err = waitCanaryReady(ctx, conn, d.Id())
			if err != nil {
				return output, fmt.Errorf("error waiting on Synthetics Canary on retry (%s): %w", d.Id(), err)
			}
		}
	}

	return nil, err
}

func deleteCanary(ctx context.Context, conn *synthetics.Synthetics, name string) error {
	_, err := conn.DeleteCanaryWithContext(ctx, &synthetics.DeleteCanaryInput{
		Name: aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, synthetics.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Synthetics Canary (%s): %w", name, err)
	}

	_, err = waitCanaryDeleted(ctx, conn, name)

	if err != nil {
		return fmt.Errorf("error waiting for Synthetics Canary (%s) delete: %w", name, err)
	}

	return nil
}
