// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitDomainControllerCreated(ctx context.Context, conn *directoryservice.Client, directoryID, domainControllerID string, timeout time.Duration) (*awstypes.DomainController, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainControllerStatusCreating),
		Target:  enum.Slice(awstypes.DomainControllerStatusActive),
		Refresh: statusDomainController(ctx, conn, directoryID, domainControllerID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DomainController); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitDomainControllerDeleted(ctx context.Context, conn *directoryservice.Client, directoryID, domainControllerID string, timeout time.Duration) (*awstypes.DomainController, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainControllerStatusDeleting),
		Target:  []string{},
		Refresh: statusDomainController(ctx, conn, directoryID, domainControllerID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DomainController); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitRadiusCompleted(ctx context.Context, conn *directoryservice.Client, directoryID string, timeout time.Duration) (*awstypes.DirectoryDescription, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RadiusStatusCreating),
		Target:  enum.Slice(awstypes.RadiusStatusCompleted),
		Refresh: statusRadius(ctx, conn, directoryID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DirectoryDescription); ok {
		return output, err
	}

	return nil, err
}

func waitRegionCreated(ctx context.Context, conn *directoryservice.Client, directoryID, regionName string, timeout time.Duration) (*awstypes.RegionDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectoryStageRequested, awstypes.DirectoryStageCreating, awstypes.DirectoryStageCreated),
		Target:  enum.Slice(awstypes.DirectoryStageActive),
		Refresh: statusRegion(ctx, conn, directoryID, regionName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.RegionDescription); ok {
		return output, err
	}

	return nil, err
}

func waitRegionDeleted(ctx context.Context, conn *directoryservice.Client, directoryID, regionName string, timeout time.Duration) (*awstypes.RegionDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DirectoryStageActive, awstypes.DirectoryStageDeleting),
		Target:  []string{},
		Refresh: statusRegion(ctx, conn, directoryID, regionName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.RegionDescription); ok {
		return output, err
	}

	return nil, err
}

func waitSharedDirectoryDeleted(ctx context.Context, conn *directoryservice.Client, ownerDirectoryID, sharedDirectoryID string, timeout time.Duration) (*awstypes.SharedDirectory, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.ShareStatusDeleting,
			awstypes.ShareStatusShared,
			awstypes.ShareStatusPendingAcceptance,
			awstypes.ShareStatusRejectFailed,
			awstypes.ShareStatusRejected,
			awstypes.ShareStatusRejecting,
		),
		Target:                    []string{},
		Refresh:                   statusSharedDirectory(ctx, conn, ownerDirectoryID, sharedDirectoryID),
		Timeout:                   timeout,
		MinTimeout:                30 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SharedDirectory); ok {
		return output, err
	}

	return nil, err
}

func waitDirectoryShared(ctx context.Context, conn *directoryservice.Client, id string, timeout time.Duration) (*awstypes.SharedDirectory, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ShareStatusPendingAcceptance, awstypes.ShareStatusSharing),
		Target:                    enum.Slice(awstypes.ShareStatusShared),
		Refresh:                   statusDirectoryShareStatus(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SharedDirectory); ok {
		return output, err
	}

	return nil, err
}
