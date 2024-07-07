// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func waitRegionCreated(ctx context.Context, conn *directoryservice.DirectoryService, directoryID, regionName string, timeout time.Duration) (*directoryservice.RegionDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageRequested, directoryservice.DirectoryStageCreating, directoryservice.DirectoryStageCreated},
		Target:  []string{directoryservice.DirectoryStageActive},
		Refresh: statusRegion(ctx, conn, directoryID, regionName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.RegionDescription); ok {
		return output, err
	}

	return nil, err
}

func waitRegionDeleted(ctx context.Context, conn *directoryservice.DirectoryService, directoryID, regionName string, timeout time.Duration) (*directoryservice.RegionDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageActive, directoryservice.DirectoryStageDeleting},
		Target:  []string{},
		Refresh: statusRegion(ctx, conn, directoryID, regionName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.RegionDescription); ok {
		return output, err
	}

	return nil, err
}
