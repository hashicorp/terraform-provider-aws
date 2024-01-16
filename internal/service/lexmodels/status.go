// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	serviceStatusCreated  = "CREATED"
	serviceStatusNotFound = "NOTFOUND"
	serviceStatusUnknown  = "UNKNOWN"
)

func statusBotVersion(ctx context.Context, conn *lexmodelbuildingservice.LexModelBuildingService, name, version string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindBotVersionByName(ctx, conn, name, version)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusSlotType(ctx context.Context, conn *lexmodelbuildingservice.LexModelBuildingService, name, version string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSlotTypeVersionByName(ctx, conn, name, version)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, serviceStatusCreated, nil
	}
}

func statusIntent(ctx context.Context, conn *lexmodelbuildingservice.LexModelBuildingService, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetIntentVersionsWithContext(ctx, &lexmodelbuildingservice.GetIntentVersionsInput{
			Name: aws.String(id),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
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

func statusBotAlias(ctx context.Context, conn *lexmodelbuildingservice.LexModelBuildingService, botAliasName, botName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetBotAliasWithContext(ctx, &lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(botAliasName),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
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
