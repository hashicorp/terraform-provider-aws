package ds

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	directoryCreatedTimeout = 60 * time.Minute
	directoryDeletedTimeout = 60 * time.Minute
)

func waitDirectoryCreated(conn *directoryservice.DirectoryService, id string, timeout time.Duration) (*directoryservice.DirectoryDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageRequested, directoryservice.DirectoryStageCreating, directoryservice.DirectoryStageCreated},
		Target:  []string{directoryservice.DirectoryStageActive},
		Refresh: statusDirectoryStage(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directoryservice.DirectoryDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StageReason)))

		return output, err
	}

	return nil, err
}

func waitDirectoryDeleted(conn *directoryservice.DirectoryService, id string, timeout time.Duration) (*directoryservice.DirectoryDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageActive, directoryservice.DirectoryStageDeleting},
		Target:  []string{},
		Refresh: statusDirectoryStage(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*directoryservice.DirectoryDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StageReason)))

		return output, err
	}

	return nil, err
}

func waitRegionCreated(ctx context.Context, conn *directoryservice.DirectoryService, directoryID, regionName string) (*directoryservice.RegionDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageRequested, directoryservice.DirectoryStageCreating, directoryservice.DirectoryStageCreated},
		Target:  []string{directoryservice.DirectoryStageActive},
		Refresh: statusRegion(ctx, conn, directoryID, regionName),
		Timeout: directoryCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.RegionDescription); ok {
		return output, err
	}

	return nil, err
}

func waitRegionDeleted(ctx context.Context, conn *directoryservice.DirectoryService, directoryID, regionName string) (*directoryservice.RegionDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageActive, directoryservice.DirectoryStageDeleting},
		Target:  []string{},
		Refresh: statusRegion(ctx, conn, directoryID, regionName),
		Timeout: directoryDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.RegionDescription); ok {
		return output, err
	}

	return nil, err
}

func waitSharedDirectoryDeleted(ctx context.Context, conn *directoryservice.DirectoryService, ownerDirectoryID, sharedDirectoryID string, timeout time.Duration) (*directoryservice.SharedDirectory, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			directoryservice.ShareStatusDeleting,
			directoryservice.ShareStatusShared,
			directoryservice.ShareStatusPendingAcceptance,
			directoryservice.ShareStatusRejectFailed,
			directoryservice.ShareStatusRejected,
			directoryservice.ShareStatusRejecting,
		},
		Target:                    []string{},
		Refresh:                   statusSharedDirectory(ctx, conn, ownerDirectoryID, sharedDirectoryID),
		Timeout:                   timeout,
		MinTimeout:                30 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.SharedDirectory); ok {
		return output, err
	}

	return nil, err
}

func waitDirectoryShared(ctx context.Context, conn *directoryservice.DirectoryService, id string, timeout time.Duration) (*directoryservice.SharedDirectory, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{directoryservice.ShareStatusPendingAcceptance, directoryservice.ShareStatusSharing},
		Target:                    []string{directoryservice.ShareStatusShared},
		Refresh:                   statusDirectoryShareStatus(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.SharedDirectory); ok {
		return output, err
	}

	return nil, err
}
