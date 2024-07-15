// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindBuildByID(ctx context.Context, conn *gamelift.Client, id string) (*awstypes.Build, error) {
	input := &gamelift.DescribeBuildInput{
		BuildId: aws.String(id),
	}

	output, err := conn.DescribeBuild(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Build == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Build, nil
}

func FindFleetByID(ctx context.Context, conn *gamelift.Client, id string) (*awstypes.FleetAttributes, error) {
	input := &gamelift.DescribeFleetAttributesInput{
		FleetIds: []string{id},
	}

	output, err := conn.DescribeFleetAttributes(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output.FleetAttributes)
}

func FindGameServerGroupByName(ctx context.Context, conn *gamelift.Client, name string) (*awstypes.GameServerGroup, error) {
	input := &gamelift.DescribeGameServerGroupInput{
		GameServerGroupName: aws.String(name),
	}

	output, err := conn.DescribeGameServerGroup(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.GameServerGroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.GameServerGroup, nil
}

func FindScriptByID(ctx context.Context, conn *gamelift.Client, id string) (*awstypes.Script, error) {
	input := &gamelift.DescribeScriptInput{
		ScriptId: aws.String(id),
	}

	output, err := conn.DescribeScript(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Script == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Script, nil
}
