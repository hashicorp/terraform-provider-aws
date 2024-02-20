// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// statusApplication fetches the ApplicationDetail and its Status
func statusApplication(ctx context.Context, conn *kinesisanalyticsv2.KinesisAnalyticsV2, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		applicationDetail, err := FindApplicationDetailByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return applicationDetail, aws.StringValue(applicationDetail.ApplicationStatus), nil
	}
}

// statusSnapshotDetails fetches the SnapshotDetails and its Status
func statusSnapshotDetails(ctx context.Context, conn *kinesisanalyticsv2.KinesisAnalyticsV2, applicationName, snapshotName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		snapshotDetails, err := FindSnapshotDetailsByApplicationAndSnapshotNames(ctx, conn, applicationName, snapshotName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return snapshotDetails, aws.StringValue(snapshotDetails.SnapshotStatus), nil
	}
}
