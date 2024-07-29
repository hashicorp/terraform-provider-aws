// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	storediSCSIVolumeAvailableTimeout = 5 * time.Minute
	smbFileShareAvailableDelay        = 5 * time.Second
	smbFileShareDeletedDelay          = 5 * time.Second
)

// waitStorediSCSIVolumeAvailable waits for a StoredIscsiVolume to return Available
func waitStorediSCSIVolumeAvailable(ctx context.Context, conn *storagegateway.Client, volumeARN string) (*storagegateway.DescribeStorediSCSIVolumesOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"BOOTSTRAPPING", "CREATING", "RESTORING"},
		Target:  []string{"AVAILABLE"},
		Refresh: statusStorediSCSIVolume(ctx, conn, volumeARN),
		Timeout: storediSCSIVolumeAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*storagegateway.DescribeStorediSCSIVolumesOutput); ok {
		return output, err
	}

	return nil, err
}

func waitSMBFileShareCreated(ctx context.Context, conn *storagegateway.Client, arn string, timeout time.Duration) (*awstypes.SMBFileShareInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fileShareStatusCreating},
		Target:  []string{fileShareStatusAvailable},
		Refresh: statusSMBFileShare(ctx, conn, arn),
		Timeout: timeout,
		Delay:   smbFileShareAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SMBFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

func waitSMBFileShareDeleted(ctx context.Context, conn *storagegateway.Client, arn string, timeout time.Duration) (*awstypes.SMBFileShareInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{fileShareStatusAvailable, fileShareStatusDeleting, fileShareStatusForceDeleting},
		Target:         []string{},
		Refresh:        statusSMBFileShare(ctx, conn, arn),
		Timeout:        timeout,
		Delay:          smbFileShareDeletedDelay,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SMBFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

func waitSMBFileShareUpdated(ctx context.Context, conn *storagegateway.Client, arn string, timeout time.Duration) (*awstypes.SMBFileShareInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fileShareStatusUpdating},
		Target:  []string{fileShareStatusAvailable},
		Refresh: statusSMBFileShare(ctx, conn, arn),
		Timeout: timeout,
		Delay:   smbFileShareAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SMBFileShareInfo); ok {
		return output, err
	}

	return nil, err
}
