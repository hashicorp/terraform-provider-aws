// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evidently

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/evidently"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusFeature(ctx context.Context, conn *evidently.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		featureName, projectNameOrARN, err := FeatureParseID(id)

		if err != nil {
			return nil, "", err
		}

		output, err := FindFeatureWithProjectNameorARN(ctx, conn, featureName, projectNameOrARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusLaunch(ctx context.Context, conn *evidently.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		launchName, projectNameOrARN, err := LaunchParseID(id)

		if err != nil {
			return nil, "", err
		}

		output, err := FindLaunchWithProjectNameorARN(ctx, conn, launchName, projectNameOrARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusProject(ctx context.Context, conn *evidently.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindProjectByNameOrARN(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}
