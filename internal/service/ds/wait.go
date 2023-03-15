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

func waitDirectoryCreated(ctx context.Context, conn *directoryservice.DirectoryService, id string, timeout time.Duration) (*directoryservice.DirectoryDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageRequested, directoryservice.DirectoryStageCreating, directoryservice.DirectoryStageCreated},
		Target:  []string{directoryservice.DirectoryStageActive},
		Refresh: statusDirectoryStage(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.DirectoryDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StageReason)))

		return output, err
	}

	return nil, err
}

func waitDirectoryDeleted(ctx context.Context, conn *directoryservice.DirectoryService, id string, timeout time.Duration) (*directoryservice.DirectoryDescription, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.DirectoryStageActive, directoryservice.DirectoryStageDeleting},
		Target:  []string{},
		Refresh: statusDirectoryStage(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.DirectoryDescription); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StageReason)))

		return output, err
	}

	return nil, err
}

func waitDomainControllerCreated(ctx context.Context, conn *directoryservice.DirectoryService, directoryID, domainControllerID string, timeout time.Duration) (*directoryservice.DomainController, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.DomainControllerStatusCreating},
		Target:  []string{directoryservice.DomainControllerStatusActive},
		Refresh: statusDomainController(ctx, conn, directoryID, domainControllerID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.DomainController); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitDomainControllerDeleted(ctx context.Context, conn *directoryservice.DirectoryService, directoryID, domainControllerID string, timeout time.Duration) (*directoryservice.DomainController, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.DomainControllerStatusDeleting},
		Target:  []string{},
		Refresh: statusDomainController(ctx, conn, directoryID, domainControllerID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.DomainController); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitRadiusCompleted(ctx context.Context, conn *directoryservice.DirectoryService, directoryID string, timeout time.Duration) (*directoryservice.DirectoryDescription, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{directoryservice.RadiusStatusCreating},
		Target:  []string{directoryservice.RadiusStatusCompleted},
		Refresh: statusRadius(ctx, conn, directoryID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.DirectoryDescription); ok {
		return output, err
	}

	return nil, err
}

func waitRegionCreated(ctx context.Context, conn *directoryservice.DirectoryService, directoryID, regionName string, timeout time.Duration) (*directoryservice.RegionDescription, error) {
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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
		Refresh:                   statusDirectoryShareStatus(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*directoryservice.SharedDirectory); ok {
		return output, err
	}

	return nil, err
}
