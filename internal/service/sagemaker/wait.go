// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

const (
	notebookInstanceInServiceTimeout  = 60 * time.Minute
	notebookInstanceStoppedTimeout    = 10 * time.Minute
	notebookInstanceDeletedTimeout    = 10 * time.Minute
	modelPackageGroupCompletedTimeout = 10 * time.Minute
	modelPackageGroupDeletedTimeout   = 10 * time.Minute
	imageCreatedTimeout               = 10 * time.Minute
	imageDeletedTimeout               = 10 * time.Minute
	imageVersionCreatedTimeout        = 10 * time.Minute
	imageVersionDeletedTimeout        = 10 * time.Minute
	domainInServiceTimeout            = 20 * time.Minute
	domainDeletedTimeout              = 20 * time.Minute
	featureGroupCreatedTimeout        = 20 * time.Minute
	featureGroupDeletedTimeout        = 10 * time.Minute
	appInServiceTimeout               = 10 * time.Minute
	appDeletedTimeout                 = 10 * time.Minute
	flowDefinitionActiveTimeout       = 2 * time.Minute
	flowDefinitionDeletedTimeout      = 2 * time.Minute
	projectCreatedTimeout             = 15 * time.Minute
	projectDeletedTimeout             = 15 * time.Minute
	workforceActiveTimeout            = 10 * time.Minute
	workforceDeletedTimeout           = 10 * time.Minute
	spaceDeletedTimeout               = 10 * time.Minute
	spaceInServiceTimeout             = 10 * time.Minute
	mlflowTrackingServerTimeout       = 45 * time.Minute
	hubTimeout                        = 10 * time.Minute

	notebookInstanceStatusNotFound = "NotFound"
)

func waitNotebookInstanceInService(ctx context.Context, conn *sagemaker.Client, notebookName string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			notebookInstanceStatusNotFound,
			awstypes.NotebookInstanceStatusUpdating,
			awstypes.NotebookInstanceStatusPending,
			awstypes.NotebookInstanceStatusStopped,
		),
		Target:  enum.Slice(awstypes.NotebookInstanceStatusInService),
		Refresh: statusNotebookInstance(conn, notebookName),
		Timeout: notebookInstanceInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		if output.NotebookInstanceStatus == awstypes.NotebookInstanceStatusFailed {
			retry.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		}

		return err
	}

	return err
}

func waitNotebookInstanceStarted(ctx context.Context, conn *sagemaker.Client, notebookName string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NotebookInstanceStatusStopped),
		Target:  enum.Slice(awstypes.NotebookInstanceStatusInService, awstypes.NotebookInstanceStatusPending),
		Refresh: statusNotebookInstance(conn, notebookName),
		Timeout: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		if output.NotebookInstanceStatus == awstypes.NotebookInstanceStatusFailed {
			retry.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		}

		return err
	}

	return err
}

func waitNotebookInstanceStopped(ctx context.Context, conn *sagemaker.Client, notebookName string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NotebookInstanceStatusUpdating, awstypes.NotebookInstanceStatusStopping),
		Target:  enum.Slice(awstypes.NotebookInstanceStatusStopped),
		Refresh: statusNotebookInstance(conn, notebookName),
		Timeout: notebookInstanceStoppedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		if output.NotebookInstanceStatus == awstypes.NotebookInstanceStatusFailed {
			retry.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		}

		return err
	}

	return err
}

func waitNotebookInstanceDeleted(ctx context.Context, conn *sagemaker.Client, notebookName string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.NotebookInstanceStatusDeleting),
		Target:  []string{},
		Refresh: statusNotebookInstance(conn, notebookName),
		Timeout: notebookInstanceDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		if output.NotebookInstanceStatus == awstypes.NotebookInstanceStatusFailed {
			retry.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitModelPackageGroupCompleted(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeModelPackageGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ModelPackageGroupStatusPending, awstypes.ModelPackageGroupStatusInProgress),
		Target:  enum.Slice(awstypes.ModelPackageGroupStatusCompleted),
		Refresh: statusModelPackageGroup(conn, name),
		Timeout: modelPackageGroupCompletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeModelPackageGroupOutput); ok {
		return output, err
	}

	return nil, err
}

func waitModelPackageGroupDeleted(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeModelPackageGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ModelPackageGroupStatusDeleting),
		Target:  []string{},
		Refresh: statusModelPackageGroup(conn, name),
		Timeout: modelPackageGroupDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeModelPackageGroupOutput); ok {
		return output, err
	}

	return nil, err
}

func waitImageCreated(ctx context.Context, conn *sagemaker.Client, name string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ImageStatusCreating, awstypes.ImageStatusUpdating),
		Target:  enum.Slice(awstypes.ImageStatusCreated),
		Refresh: statusImage(conn, name),
		Timeout: imageCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeImageOutput); ok {
		if status, reason := output.ImageStatus, aws.ToString(output.FailureReason); (status == awstypes.ImageStatusCreateFailed || status == awstypes.ImageStatusUpdateFailed) && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return err
	}

	return err
}

func waitImageDeleted(ctx context.Context, conn *sagemaker.Client, name string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ImageStatusDeleting),
		Target:  []string{},
		Refresh: statusImage(conn, name),
		Timeout: imageDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeImageOutput); ok {
		if status, reason := output.ImageStatus, aws.ToString(output.FailureReason); status == awstypes.ImageStatusDeleteFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return err
	}

	return err
}

func waitImageVersionCreated(ctx context.Context, conn *sagemaker.Client, name string, version int32) (*sagemaker.DescribeImageVersionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ImageVersionStatusCreating),
		Target:  enum.Slice(awstypes.ImageVersionStatusCreated),
		Refresh: statusImageVersionByTwoPartKey(conn, name, version),
		Timeout: imageVersionCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeImageVersionOutput); ok {
		if status, reason := output.ImageVersionStatus, aws.ToString(output.FailureReason); status == awstypes.ImageVersionStatusCreateFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitImageVersionDeleted(ctx context.Context, conn *sagemaker.Client, name string, version int32) (*sagemaker.DescribeImageVersionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ImageVersionStatusDeleting),
		Target:  []string{},
		Refresh: statusImageVersionByID(conn, name, version),
		Timeout: imageVersionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeImageVersionOutput); ok {
		if status, reason := output.ImageVersionStatus, aws.ToString(output.FailureReason); status == awstypes.ImageVersionStatusDeleteFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitDomainInService(ctx context.Context, conn *sagemaker.Client, domainID string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainStatusPending, awstypes.DomainStatusUpdating),
		Target:  enum.Slice(awstypes.DomainStatusInService),
		Refresh: statusDomain(conn, domainID),
		Timeout: domainInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeDomainOutput); ok {
		if status, reason := output.Status, aws.ToString(output.FailureReason); status == awstypes.DomainStatusFailed || status == awstypes.DomainStatusUpdateFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return err
	}

	return err
}

func waitDomainDeleted(ctx context.Context, conn *sagemaker.Client, domainID string) (*sagemaker.DescribeDomainOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainStatusDeleting),
		Target:  []string{},
		Refresh: statusDomain(conn, domainID),
		Timeout: domainDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeDomainOutput); ok {
		if status, reason := output.Status, aws.ToString(output.FailureReason); status == awstypes.DomainStatusDeleteFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitFeatureGroupCreated(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeFeatureGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FeatureGroupStatusCreating),
		Target:  enum.Slice(awstypes.FeatureGroupStatusCreated),
		Refresh: statusFeatureGroup(conn, name),
		Timeout: featureGroupCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeFeatureGroupOutput); ok {
		if status, reason := output.FeatureGroupStatus, aws.ToString(output.FailureReason); status == awstypes.FeatureGroupStatusCreateFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitFeatureGroupDeleted(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeFeatureGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FeatureGroupStatusDeleting),
		Target:  []string{},
		Refresh: statusFeatureGroup(conn, name),
		Timeout: featureGroupDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeFeatureGroupOutput); ok {
		if status, reason := output.FeatureGroupStatus, aws.ToString(output.FailureReason); status == awstypes.FeatureGroupStatusDeleteFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitFeatureGroupUpdated(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeFeatureGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LastUpdateStatusValueInProgress),
		Target:  enum.Slice(awstypes.LastUpdateStatusValueSuccessful),
		Refresh: statusFeatureGroupUpdate(conn, name),
		Timeout: featureGroupDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeFeatureGroupOutput); ok {
		if v := output.LastUpdateStatus; v != nil && v.Status == awstypes.LastUpdateStatusValueFailed {
			retry.SetLastError(err, errors.New(*v.FailureReason))
		}

		return output, err
	}

	return nil, err
}

func waitAppInService(ctx context.Context, conn *sagemaker.Client, domainID, userProfileOrSpaceName, appType, appName string) (*sagemaker.DescribeAppOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AppStatusPending),
		Target:  enum.Slice(awstypes.AppStatusInService),
		Refresh: statusApp(conn, domainID, userProfileOrSpaceName, appType, appName),
		Timeout: appInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeAppOutput); ok {
		if status, reason := output.Status, aws.ToString(output.FailureReason); status == awstypes.AppStatusFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitAppDeleted(ctx context.Context, conn *sagemaker.Client, domainID, userProfileOrSpaceName, appType, appName string) (*sagemaker.DescribeAppOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AppStatusDeleting),
		Target:  []string{},
		Refresh: statusApp(conn, domainID, userProfileOrSpaceName, appType, appName),
		Timeout: appDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeAppOutput); ok {
		if status, reason := output.Status, aws.ToString(output.FailureReason); status == awstypes.AppStatusFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitFlowDefinitionActive(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeFlowDefinitionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FlowDefinitionStatusInitializing),
		Target:  enum.Slice(awstypes.FlowDefinitionStatusActive),
		Refresh: statusFlowDefinition(conn, name),
		Timeout: flowDefinitionActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeFlowDefinitionOutput); ok {
		if status, reason := output.FlowDefinitionStatus, aws.ToString(output.FailureReason); status == awstypes.FlowDefinitionStatusFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitFlowDefinitionDeleted(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeFlowDefinitionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FlowDefinitionStatusDeleting),
		Target:  []string{},
		Refresh: statusFlowDefinition(conn, name),
		Timeout: flowDefinitionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeFlowDefinitionOutput); ok {
		if status, reason := output.FlowDefinitionStatus, aws.ToString(output.FailureReason); status == awstypes.FlowDefinitionStatusFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitProjectDeleted(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeProjectOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ProjectStatusDeleteInProgress, awstypes.ProjectStatusPending),
		Target:  []string{},
		Refresh: statusProject(conn, name),
		Timeout: projectDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeProjectOutput); ok {
		if status, reason := output.ProjectStatus, aws.ToString(output.ServiceCatalogProvisionedProductDetails.ProvisionedProductStatusMessage); status == awstypes.ProjectStatusDeleteFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitProjectCreated(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeProjectOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ProjectStatusPending, awstypes.ProjectStatusCreateInProgress),
		Target:  enum.Slice(awstypes.ProjectStatusCreateCompleted),
		Refresh: statusProject(conn, name),
		Timeout: projectCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeProjectOutput); ok {
		if status, reason := output.ProjectStatus, aws.ToString(output.ServiceCatalogProvisionedProductDetails.ProvisionedProductStatusMessage); status == awstypes.ProjectStatusCreateFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitProjectUpdated(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeProjectOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ProjectStatusPending, awstypes.ProjectStatusUpdateInProgress),
		Target:  enum.Slice(awstypes.ProjectStatusUpdateCompleted),
		Refresh: statusProject(conn, name),
		Timeout: projectCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeProjectOutput); ok {
		if status, reason := output.ProjectStatus, aws.ToString(output.ServiceCatalogProvisionedProductDetails.ProvisionedProductStatusMessage); status == awstypes.ProjectStatusUpdateFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitWorkforceActive(ctx context.Context, conn *sagemaker.Client, name string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkforceStatusInitializing, awstypes.WorkforceStatusUpdating),
		Target:  enum.Slice(awstypes.WorkforceStatusActive),
		Refresh: statusWorkforce(conn, name),
		Timeout: workforceActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Workforce); ok {
		if status, reason := output.Status, aws.ToString(output.FailureReason); status == awstypes.WorkforceStatusFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return err
	}

	return err
}

func waitWorkforceDeleted(ctx context.Context, conn *sagemaker.Client, name string) (*awstypes.Workforce, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkforceStatusDeleting),
		Target:  []string{},
		Refresh: statusWorkforce(conn, name),
		Timeout: workforceDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Workforce); ok {
		if status, reason := output.Status, aws.ToString(output.FailureReason); status == awstypes.WorkforceStatusFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitSpaceInService(ctx context.Context, conn *sagemaker.Client, domainId, name string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SpaceStatusPending, awstypes.SpaceStatusUpdating),
		Target:  enum.Slice(awstypes.SpaceStatusInService),
		Refresh: statusSpace(conn, domainId, name),
		Timeout: spaceInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeSpaceOutput); ok {
		if status, reason := output.Status, aws.ToString(output.FailureReason); status == awstypes.SpaceStatusUpdateFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return err
	}

	return err
}

func waitSpaceDeleted(ctx context.Context, conn *sagemaker.Client, domainId, name string) (*sagemaker.DescribeSpaceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SpaceStatusDeleting),
		Target:  []string{},
		Refresh: statusSpace(conn, domainId, name),
		Timeout: spaceDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeSpaceOutput); ok {
		if status, reason := output.Status, aws.ToString(output.FailureReason); status == awstypes.SpaceStatusDeleteFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitMlflowTrackingServerCreated(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeMlflowTrackingServerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TrackingServerStatusCreating),
		Target:  enum.Slice(awstypes.TrackingServerStatusCreated),
		Refresh: statusMlflowTrackingServer(conn, name),
		Timeout: mlflowTrackingServerTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeMlflowTrackingServerOutput); ok {
		return output, err
	}

	return nil, err
}

func waitMlflowTrackingServerUpdated(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeMlflowTrackingServerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TrackingServerStatusUpdating),
		Target:  enum.Slice(awstypes.TrackingServerStatusUpdated),
		Refresh: statusMlflowTrackingServer(conn, name),
		Timeout: mlflowTrackingServerTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeMlflowTrackingServerOutput); ok {
		return output, err
	}

	return nil, err
}

func waitMlflowTrackingServerDeleted(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeMlflowTrackingServerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TrackingServerStatusDeleting),
		Target:  []string{},
		Refresh: statusMlflowTrackingServer(conn, name),
		Timeout: mlflowTrackingServerTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeMlflowTrackingServerOutput); ok {
		return output, err
	}

	return nil, err
}

func waitHubInService(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeHubOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.HubStatusCreating),
		Target:  enum.Slice(awstypes.HubStatusInService),
		Refresh: statusHub(conn, name),
		Timeout: hubTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeHubOutput); ok {
		if status, reason := output.HubStatus, aws.ToString(output.FailureReason); status == awstypes.HubStatusCreateFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitHubDeleted(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeHubOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.HubStatusDeleting),
		Target:  []string{},
		Refresh: statusHub(conn, name),
		Timeout: hubTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeHubOutput); ok {
		if status, reason := output.HubStatus, aws.ToString(output.FailureReason); status == awstypes.HubStatusDeleteFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

func waitHubUpdated(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeHubOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.HubStatusUpdating),
		Target:  enum.Slice(awstypes.HubStatusInService),
		Refresh: statusHub(conn, name),
		Timeout: hubTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeHubOutput); ok {
		if status, reason := output.HubStatus, aws.ToString(output.FailureReason); status == awstypes.HubStatusUpdateFailed && reason != "" {
			retry.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}
