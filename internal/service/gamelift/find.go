// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindBuildByID(ctx context.Context, conn *gamelift.GameLift, id string) (*gamelift.Build, error) {
	input := &gamelift.DescribeBuildInput{
		BuildId: aws.String(id),
	}

	output, err := conn.DescribeBuildWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
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

func FindFleetByID(ctx context.Context, conn *gamelift.GameLift, id string) (*gamelift.FleetAttributes, error) {
	input := &gamelift.DescribeFleetAttributesInput{
		FleetIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeFleetAttributesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output.FleetAttributes) == 0 || output.FleetAttributes[0] == nil {
		return nil, tfresource.NewEmptyResultError(output)
	}

	if count := len(output.FleetAttributes); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, output)
	}

	fleet := output.FleetAttributes[0]

	if aws.StringValue(fleet.FleetId) != id {
		return nil, tfresource.NewEmptyResultError(id)
	}

	return fleet, nil
}

func FindGameServerGroupByName(ctx context.Context, conn *gamelift.GameLift, name string) (*gamelift.GameServerGroup, error) {
	input := &gamelift.DescribeGameServerGroupInput{
		GameServerGroupName: aws.String(name),
	}

	output, err := conn.DescribeGameServerGroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
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

func FindScriptByID(ctx context.Context, conn *gamelift.GameLift, id string) (*gamelift.Script, error) {
	input := &gamelift.DescribeScriptInput{
		ScriptId: aws.String(id),
	}

	output, err := conn.DescribeScriptWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
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
