// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/synthetics"
	awstypes "github.com/aws/aws-sdk-go-v2/service/synthetics/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func retryCreateCanary(ctx context.Context, conn *synthetics.Client, d *schema.ResourceData, input *synthetics.CreateCanaryInput) (*awstypes.Canary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CanaryStateCreating, awstypes.CanaryStateUpdating),
		Target:  enum.Slice(awstypes.CanaryStateReady),
		Refresh: statusCanaryState(ctx, conn, d.Id()),
		Timeout: canaryCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*awstypes.Canary); ok {
		if status := output.Status; status.State == awstypes.CanaryStateError && status.StateReasonCode == awstypes.CanaryStateReasonCodeCreateFailed {
			// delete canary because it is the only way to reprovision if in an error state
			err = deleteCanary(ctx, conn, d.Id())
			if err != nil {
				return output, fmt.Errorf("deleting Synthetics Canary on retry (%s): %w", d.Id(), err)
			}

			_, err = conn.CreateCanary(ctx, input)
			if err != nil {
				return output, fmt.Errorf("creating Synthetics Canary on retry (%s): %w", d.Id(), err)
			}

			_, err = waitCanaryReady(ctx, conn, d.Id())
			if err != nil {
				return output, fmt.Errorf("waiting on Synthetics Canary on retry (%s): %w", d.Id(), err)
			}
		}
	}

	return nil, err
}

func deleteCanary(ctx context.Context, conn *synthetics.Client, name string) error {
	_, err := conn.DeleteCanary(ctx, &synthetics.DeleteCanaryInput{
		Name: aws.String(name),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Synthetics Canary (%s): %w", name, err)
	}

	_, err = waitCanaryDeleted(ctx, conn, name)

	if err != nil {
		return fmt.Errorf("waiting for Synthetics Canary (%s) delete: %w", name, err)
	}

	return nil
}
