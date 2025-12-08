// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	serviceStatusCreated  = "CREATED"
	serviceStatusNotFound = "NOTFOUND"
	serviceStatusUnknown  = "UNKNOWN"
)

func statusBotVersion(ctx context.Context, conn *lexmodelbuildingservice.Client, name, version string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findBotVersionByName(ctx, conn, name, version)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusSlotType(ctx context.Context, conn *lexmodelbuildingservice.Client, name, version string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findSlotTypeVersionByName(ctx, conn, name, version)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, serviceStatusCreated, nil
	}
}

func statusIntent(ctx context.Context, conn *lexmodelbuildingservice.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := conn.GetIntentVersions(ctx, &lexmodelbuildingservice.GetIntentVersionsInput{
			Name: aws.String(id),
		})
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, serviceStatusNotFound, nil
		}
		if err != nil {
			return nil, serviceStatusUnknown, err
		}

		if output == nil || len(output.Intents) == 0 {
			return nil, serviceStatusNotFound, nil
		}

		return output, serviceStatusCreated, nil
	}
}

func statusBotAlias(ctx context.Context, conn *lexmodelbuildingservice.Client, botAliasName, botName string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := conn.GetBotAlias(ctx, &lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(botAliasName),
		})
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil, serviceStatusNotFound, nil
		}
		if err != nil {
			return nil, serviceStatusUnknown, err
		}
		if output == nil {
			return nil, serviceStatusNotFound, nil
		}

		return output, serviceStatusCreated, nil
	}
}
