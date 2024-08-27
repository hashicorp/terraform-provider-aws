// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// statusFleetState fetches the fleet and its state
func statusFleetState(ctx context.Context, conn *appstream.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		fleet, err := FindFleetByName(ctx, conn, name)

		if err != nil {
			return nil, "Unknown", err
		}

		if fleet == nil {
			return fleet, "NotFound", nil
		}

		return fleet, string(fleet.State), nil
	}
}

func statusImageBuilderState(ctx context.Context, conn *appstream.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindImageBuilderByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

// statusUserAvailable fetches the user available
func statusUserAvailable(ctx context.Context, conn *appstream.Client, username, authType string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		user, err := FindUserByTwoPartKey(ctx, conn, username, authType)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return user, userAvailable, nil
	}
}
