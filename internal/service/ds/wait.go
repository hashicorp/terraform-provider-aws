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
	directoryCreatedTimeout       = 60 * time.Minute
	directoryDeletedTimeout       = 60 * time.Minute
	sharedDirectoryDeletedTimeout = 60 * time.Minute
)

func waitDirectoryCreated(conn *directoryservice.DirectoryService, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageRequested, directoryservice.DirectoryStageCreating, directoryservice.DirectoryStageCreated},
		Target:  []string{directoryservice.DirectoryStageActive},
		Refresh: statusDirectoryStage(conn, id),
		Timeout: directoryCreatedTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitDirectoryDeleted(conn *directoryservice.DirectoryService, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageActive, directoryservice.DirectoryStageDeleting},
		Target:  []string{},
		Refresh: statusDirectoryStage(conn, id),
		Timeout: directoryDeletedTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitSharedDirectoryDeleted(ctx context.Context, conn *directoryservice.DirectoryService, ownerId, sharedId string) (*directoryservice.SharedDirectory, error) {
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
		Refresh:                   statusSharedDirectory(ctx, conn, ownerId, sharedId),
		Timeout:                   sharedDirectoryDeletedTimeout,
		MinTimeout:                30 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.SharedDirectory); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ShareStatus)))

		return output, err
	}

	return nil, err
}

func waitDirectoryShared(ctx context.Context, conn *directoryservice.DirectoryService, dirId string) (*directoryservice.SharedDirectory, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			directoryservice.ShareStatusSharing,
			directoryservice.ShareStatusPendingAcceptance,
		},
		Target:                    []string{directoryservice.ShareStatusShared},
		Refresh:                   statusDirectoryShare(conn, dirId),
		Timeout:                   sharedDirectoryDeletedTimeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.SharedDirectory); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ShareStatus)))

		return output, err
	}

	return nil, err
}
