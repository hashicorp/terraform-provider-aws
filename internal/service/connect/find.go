// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindBotAssociationV1ByNameAndRegionWithContext(ctx context.Context, conn *connect.Client, instanceID, name, region string) (*awstypes.LexBot, error) {
	var result *awstypes.LexBot

	input := &connect.ListBotsInput{
		InstanceId: aws.String(instanceID),
		LexVersion: awstypes.LexVersionV1,
		MaxResults: aws.Int32(ListBotsMaxResults),
	}

	pages := connect.NewListBotsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, cf := range page.LexBots {
			if name != "" && aws.ToString(cf.LexBot.Name) != name {
				continue
			}

			if region != "" && aws.ToString(cf.LexBot.LexRegion) != region {
				continue
			}

			result = cf.LexBot
		}
	}

	if result == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return result, nil
}

func FindLambdaFunctionAssociationByARNWithContext(ctx context.Context, conn *connect.Client, instanceID string, functionArn string) (string, error) {
	var result string

	input := &connect.ListLambdaFunctionsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(ListLambdaFunctionsMaxResults),
	}

	pages := connect.NewListLambdaFunctionsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return "", &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return "", err
		}

		for _, cf := range page.LambdaFunctions {
			if cf == functionArn {
				result = functionArn
			}
		}
	}

	if result == "" {
		return "", tfresource.NewEmptyResultError(input)
	}

	return result, nil
}
