// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	botAliasDeletedTimeout = 5 * time.Minute
	intentDeletedTimeout   = 5 * time.Minute
)

func waitBotVersionCreated(ctx context.Context, conn *lexmodelbuildingservice.Client, name, version string, timeout time.Duration) (*lexmodelbuildingservice.GetBotOutput, error) { //nolint:unparam
	stateChangeConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusBuilding),
		Target: enum.Slice(
			awstypes.StatusNotBuilt,
			awstypes.StatusReady,
			awstypes.StatusReadyBasicTesting,
		),
		Refresh: statusBotVersion(ctx, conn, name, version),
		Timeout: timeout,
	}

	outputRaw, err := stateChangeConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelbuildingservice.GetBotOutput); ok {
		if output.Status == awstypes.StatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitBotDeleted(ctx context.Context, conn *lexmodelbuildingservice.Client, name string, timeout time.Duration) (*lexmodelbuildingservice.GetBotOutput, error) {
	stateChangeConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.StatusNotBuilt,
			awstypes.StatusReady,
			awstypes.StatusReadyBasicTesting,
		),
		Target:  []string{},
		Refresh: statusBotVersion(ctx, conn, name, BotVersionLatest),
		Timeout: timeout,
	}

	outputRaw, err := stateChangeConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*lexmodelbuildingservice.GetBotOutput); ok {
		if output.Status == awstypes.StatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitBotAliasDeleted(ctx context.Context, conn *lexmodelbuildingservice.Client, botAliasName, botName string) (*lexmodelbuildingservice.GetBotAliasOutput, error) {
	stateChangeConf := &retry.StateChangeConf{
		Pending: []string{serviceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: statusBotAlias(ctx, conn, botAliasName, botName),
		Timeout: botAliasDeletedTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetBotAliasOutput); ok {
		return v, err
	}

	return nil, err
}

func waitIntentDeleted(ctx context.Context, conn *lexmodelbuildingservice.Client, intentId string) (*lexmodelbuildingservice.GetIntentVersionsOutput, error) {
	stateChangeConf := &retry.StateChangeConf{
		Pending: []string{serviceStatusCreated},
		Target:  []string{}, // An empty slice indicates that the resource is gone
		Refresh: statusIntent(ctx, conn, intentId),
		Timeout: intentDeletedTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetIntentVersionsOutput); ok {
		return v, err
	}

	return nil, err
}

func waitSlotTypeDeleted(ctx context.Context, conn *lexmodelbuildingservice.Client, name string) (*lexmodelbuildingservice.GetSlotTypeOutput, error) {
	stateChangeConf := &retry.StateChangeConf{
		Pending: []string{serviceStatusCreated},
		Target:  []string{},
		Refresh: statusSlotType(ctx, conn, name, SlotTypeVersionLatest),
		Timeout: slotTypeDeleteTimeout,
	}
	outputRaw, err := stateChangeConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*lexmodelbuildingservice.GetSlotTypeOutput); ok {
		return v, err
	}

	return nil, err
}
