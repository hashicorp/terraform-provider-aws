// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findChangeSetByTwoPartKey(ctx context.Context, conn *cloudformation.Client, stackID, changeSetName string) (*cloudformation.DescribeChangeSetOutput, error) {
	input := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: aws.String(changeSetName),
		StackName:     aws.String(stackID),
	}

	output, err := conn.DescribeChangeSet(ctx, input)

	if errs.IsA[*awstypes.ChangeSetNotFoundException](err) {
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

func statusChangeSet(ctx context.Context, conn *cloudformation.Client, stackID, changeSetName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findChangeSetByTwoPartKey(ctx, conn, stackID, changeSetName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitChangeSetCreated(ctx context.Context, conn *cloudformation.Client, stackID, changeSetName string) (*cloudformation.DescribeChangeSetOutput, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ChangeSetStatusCreateInProgress, awstypes.ChangeSetStatusCreatePending),
		Target:  enum.Slice(awstypes.ChangeSetStatusCreateComplete),
		Timeout: timeout,
		Refresh: statusChangeSet(ctx, conn, stackID, changeSetName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudformation.DescribeChangeSetOutput); ok {
		if output.Status == awstypes.ChangeSetStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}
