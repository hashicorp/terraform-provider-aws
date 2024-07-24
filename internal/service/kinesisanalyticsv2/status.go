// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// statusSnapshotDetails fetches the SnapshotDetails and its Status
func statusSnapshotDetails(ctx context.Context, conn *kinesisanalyticsv2.Client, applicationName, snapshotName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		snapshotDetails, err := FindSnapshotDetailsByApplicationAndSnapshotNames(ctx, conn, applicationName, snapshotName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return snapshotDetails, string(snapshotDetails.SnapshotStatus), nil
	}
}
