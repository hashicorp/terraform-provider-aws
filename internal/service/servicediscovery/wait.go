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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	OperationSuccessTimeout = 5 * time.Minute
)

// WaitOperationSuccess waits for an Operation to return Success
func WaitOperationSuccess(ctx context.Context, conn *servicediscovery.Client, operationID string) (*awstypes.Operation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.OperationStatusSubmitted, awstypes.OperationStatusPending),
		Target:  enum.Slice(awstypes.OperationStatusSuccess),
		Refresh: StatusOperation(ctx, conn, operationID),
		Timeout: OperationSuccessTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Operation); ok {
		// Error messages can also be contained in the response with FAIL status
		//   "ErrorCode":"CANNOT_CREATE_HOSTED_ZONE",
		//   "ErrorMessage":"The VPC that you chose, vpc-xxx in region xxx, is already associated with another private hosted zone that has an overlapping name space, xxx.. (Service: AmazonRoute53; Status Code: 400; Error Code: ConflictingDomainExists; Request ID: xxx)"
		//   "Status":"FAIL",
		if output.Status == awstypes.OperationStatusFail {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.ToString(output.ErrorCode), aws.ToString(output.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}
