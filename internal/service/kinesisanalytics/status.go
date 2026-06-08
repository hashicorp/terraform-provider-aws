// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesisanalytics

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kinesisanalytics"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func statusApplication(conn *kinesisanalytics.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		applicationDetail, err := findApplicationDetailByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return applicationDetail, string(applicationDetail.ApplicationStatus), nil
	}
}
