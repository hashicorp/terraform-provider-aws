// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicediscovery

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicediscovery"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicediscovery/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitOperationSucceeded(ctx context.Context, conn *servicediscovery.Client, operationID string) (*awstypes.Operation, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.OperationStatusSubmitted, awstypes.OperationStatusPending),
		Target:  enum.Slice(awstypes.OperationStatusSuccess),
		Refresh: statusOperation(ctx, conn, operationID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Operation); ok {
		// Error messages can also be contained in the response with FAIL status
		//   "ErrorCode":"CANNOT_CREATE_HOSTED_ZONE",
		//   "ErrorMessage":"The VPC that you chose, vpc-xxx in region xxx, is already associated with another private hosted zone that has an overlapping name space, xxx.. (Service: AmazonRoute53; Status Code: 400; Error Code: ConflictingDomainExists; Request ID: xxx)"
		//   "Status":"FAIL",
		if status := output.Status; status == awstypes.OperationStatusFail {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.ErrorCode), aws.ToString(output.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}

func statusOperation(ctx context.Context, conn *servicediscovery.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findOperationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findOperationByID(ctx context.Context, conn *servicediscovery.Client, id string) (*awstypes.Operation, error) {
	input := &servicediscovery.GetOperationInput{
		OperationId: aws.String(id),
	}

	output, err := conn.GetOperation(ctx, input)

	if errs.IsA[*awstypes.OperationNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Operation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Operation, nil
}
