// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// waitImageStatusAvailable waits for an Image to return Available
func waitImageStatusAvailable(ctx context.Context, conn *imagebuilder.Client, imageBuildVersionArn string, timeout time.Duration) (*imagebuilder.Image, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			string(awstypes.ImageStatusBuilding),
			string(awstypes.ImageStatusCreating),
			string(awstypes.ImageStatusDistributing),
			string(awstypes.ImageStatusIntegrating),
			string(awstypes.ImageStatusPending),
			string(awstypes.ImageStatusTesting),
		},
		Target:  []string{string(awstypes.ImageStatusAvailable)},
		Refresh: statusImage(ctx, conn, imageBuildVersionArn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.Image); ok {
		return v, err
	}

	return nil, err
}
