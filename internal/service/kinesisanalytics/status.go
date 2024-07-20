// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalytics

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kinesisanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// statusApplication fetches the ApplicationDetail and its Status
func statusApplication(ctx context.Context, conn *kinesisanalytics.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		applicationDetail, err := FindApplicationDetailByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return applicationDetail, string(applicationDetail.ApplicationStatus), nil
	}
}
