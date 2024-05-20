// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitOperationSucceeded(ctx context.Context, conn *route53domains.Client, id string, timeout time.Duration) (*route53domains.GetOperationDetailOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.OperationStatusSubmitted, types.OperationStatusInProgress),
		Target:  enum.Slice(types.OperationStatusSuccessful),
		Timeout: timeout,
		Refresh: statusOperation(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53domains.GetOperationDetailOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.Message)))

		return output, err
	}

	return nil, err
}

func statusOperation(ctx context.Context, conn *route53domains.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findOperationDetailByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findOperationDetailByID(ctx context.Context, conn *route53domains.Client, id string) (*route53domains.GetOperationDetailOutput, error) {
	input := &route53domains.GetOperationDetailInput{
		OperationId: aws.String(id),
	}

	output, err := conn.GetOperationDetail(ctx, input)

	if errs.IsAErrorMessageContains[*types.InvalidInput](err, "No operation found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
