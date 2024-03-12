// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

// waitImageStatusAvailable waits for an Image to return Available
func waitImageStatusAvailable(ctx context.Context, conn *imagebuilder.Client, imageBuildVersionArn string, timeout time.Duration) (*awstypes.Image, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.ImageStatusBuilding,
			awstypes.ImageStatusCreating,
			awstypes.ImageStatusDistributing,
			awstypes.ImageStatusIntegrating,
			awstypes.ImageStatusPending,
			awstypes.ImageStatusTesting,
		),
		Target:  enum.Slice(awstypes.ImageStatusAvailable),
		Refresh: statusImage(ctx, conn, imageBuildVersionArn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.Image); ok {
		return v, err
	}

	return nil, err
}
