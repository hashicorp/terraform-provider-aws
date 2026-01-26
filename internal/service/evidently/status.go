// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package evidently

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/evidently"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func statusFeature(conn *evidently.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		featureName, projectNameOrARN, err := FeatureParseID(id)

		if err != nil {
			return nil, "", err
		}

		output, err := FindFeatureWithProjectNameorARN(ctx, conn, featureName, projectNameOrARN)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusLaunch(conn *evidently.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		launchName, projectNameOrARN, err := LaunchParseID(id)

		if err != nil {
			return nil, "", err
		}

		output, err := FindLaunchWithProjectNameorARN(ctx, conn, launchName, projectNameOrARN)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusProject(conn *evidently.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := FindProjectByNameOrARN(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}
