// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/synthetics"
	awstypes "github.com/aws/aws-sdk-go-v2/service/synthetics/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	canaryCreatedTimeout = 5 * time.Minute
	canaryRunningTimeout = 5 * time.Minute
	canaryStoppedTimeout = 5 * time.Minute
	canaryDeletedTimeout = 5 * time.Minute
)

func waitCanaryReady(ctx context.Context, conn *synthetics.Client, name string) (*awstypes.Canary, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CanaryStateUpdating, awstypes.CanaryStateCreating),
		Target:  enum.Slice(awstypes.CanaryStateReady),
		Refresh: statusCanaryState(ctx, conn, name),
		Timeout: canaryCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Canary); ok {
		if status := output.Status; status.State == awstypes.CanaryStateError {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", status.StateReasonCode, aws.ToString(status.StateReason)))
		}

		return output, err
	}

	return nil, err
}

func waitCanaryStopped(ctx context.Context, conn *synthetics.Client, name string) (*awstypes.Canary, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.CanaryStateStopping,
			awstypes.CanaryStateUpdating,
			awstypes.CanaryStateRunning,
			awstypes.CanaryStateReady,
			awstypes.CanaryStateStarting),
		Target:  enum.Slice(awstypes.CanaryStateStopped),
		Refresh: statusCanaryState(ctx, conn, name),
		Timeout: canaryStoppedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Canary); ok {
		if status := output.Status; status.State == awstypes.CanaryStateError {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", status.StateReasonCode, aws.ToString(status.StateReason)))
		}

		return output, err
	}

	return nil, err
}

func waitCanaryRunning(ctx context.Context, conn *synthetics.Client, name string) (*awstypes.Canary, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.CanaryStateStarting,
			awstypes.CanaryStateUpdating,
			awstypes.CanaryStateReady),
		Target:  enum.Slice(awstypes.CanaryStateRunning),
		Refresh: statusCanaryState(ctx, conn, name),
		Timeout: canaryRunningTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Canary); ok {
		if status := output.Status; status.State == awstypes.CanaryStateError {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", status.StateReasonCode, aws.ToString(status.StateReason)))
		}

		return output, err
	}

	return nil, err
}

func waitCanaryDeleted(ctx context.Context, conn *synthetics.Client, name string) (*awstypes.Canary, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CanaryStateDeleting, awstypes.CanaryStateStopped),
		Target:  []string{},
		Refresh: statusCanaryState(ctx, conn, name),
		Timeout: canaryDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Canary); ok {
		if status := output.Status; status.State == awstypes.CanaryStateError {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", status.StateReasonCode, aws.ToString(status.StateReason)))
		}

		return output, err
	}

	return nil, err
}
