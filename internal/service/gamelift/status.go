// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusBuild(ctx context.Context, conn *gamelift.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindBuildByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusFleet(ctx context.Context, conn *gamelift.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFleetByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusGameServerGroup(ctx context.Context, conn *gamelift.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindGameServerGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}
