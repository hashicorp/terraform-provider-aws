package sagemaker

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	NotebookInstanceInServiceTimeout  = 60 * time.Minute
	NotebookInstanceStoppedTimeout    = 10 * time.Minute
	NotebookInstanceDeletedTimeout    = 10 * time.Minute
	ModelPackageGroupCompletedTimeout = 10 * time.Minute
	ModelPackageGroupDeletedTimeout   = 10 * time.Minute
	ImageCreatedTimeout               = 10 * time.Minute
	ImageDeletedTimeout               = 10 * time.Minute
	ImageVersionCreatedTimeout        = 10 * time.Minute
	ImageVersionDeletedTimeout        = 10 * time.Minute
	DomainInServiceTimeout            = 10 * time.Minute
	DomainDeletedTimeout              = 10 * time.Minute
	FeatureGroupCreatedTimeout        = 10 * time.Minute
	FeatureGroupDeletedTimeout        = 10 * time.Minute
	UserProfileInServiceTimeout       = 10 * time.Minute
	UserProfileDeletedTimeout         = 10 * time.Minute
	AppInServiceTimeout               = 10 * time.Minute
	AppDeletedTimeout                 = 10 * time.Minute
	FlowDefinitionActiveTimeout       = 2 * time.Minute
	FlowDefinitionDeletedTimeout      = 2 * time.Minute
	ProjectCreatedTimeout             = 2 * time.Minute
	ProjectDeletedTimeout             = 5 * time.Minute
)

// WaitNotebookInstanceInService waits for a NotebookInstance to return InService
func WaitNotebookInstanceInService(conn *sagemaker.SageMaker, notebookName string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			notebookInstanceStatusNotFound,
			sagemaker.NotebookInstanceStatusUpdating,
			sagemaker.NotebookInstanceStatusPending,
			sagemaker.NotebookInstanceStatusStopped,
		},
		Target:  []string{sagemaker.NotebookInstanceStatusInService},
		Refresh: StatusNotebookInstance(conn, notebookName),
		Timeout: NotebookInstanceInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitNotebookInstanceStopped waits for a NotebookInstance to return Stopped
func WaitNotebookInstanceStopped(conn *sagemaker.SageMaker, notebookName string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.NotebookInstanceStatusUpdating,
			sagemaker.NotebookInstanceStatusStopping,
		},
		Target:  []string{sagemaker.NotebookInstanceStatusStopped},
		Refresh: StatusNotebookInstance(conn, notebookName),
		Timeout: NotebookInstanceStoppedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitNotebookInstanceDeleted waits for a NotebookInstance to return Deleted
func WaitNotebookInstanceDeleted(conn *sagemaker.SageMaker, notebookName string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.NotebookInstanceStatusDeleting,
		},
		Target:  []string{},
		Refresh: StatusNotebookInstance(conn, notebookName),
		Timeout: NotebookInstanceDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeNotebookInstanceOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitModelPackageGroupCompleted waits for a ModelPackageGroup to return Created
func WaitModelPackageGroupCompleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeModelPackageGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.ModelPackageGroupStatusPending,
			sagemaker.ModelPackageGroupStatusInProgress,
		},
		Target:  []string{sagemaker.ModelPackageGroupStatusCompleted},
		Refresh: StatusModelPackageGroup(conn, name),
		Timeout: ModelPackageGroupCompletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeModelPackageGroupOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitModelPackageGroupDeleted waits for a ModelPackageGroup to return Created
func WaitModelPackageGroupDeleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeModelPackageGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.ModelPackageGroupStatusDeleting,
		},
		Target:  []string{},
		Refresh: StatusModelPackageGroup(conn, name),
		Timeout: ModelPackageGroupDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeModelPackageGroupOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitImageCreated waits for a Image to return Created
func WaitImageCreated(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.ImageStatusCreating,
			sagemaker.ImageStatusUpdating,
		},
		Target:  []string{sagemaker.ImageStatusCreated},
		Refresh: StatusImage(conn, name),
		Timeout: ImageCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeImageOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitImageDeleted waits for a Image to return Deleted
func WaitImageDeleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.ImageStatusDeleting},
		Target:  []string{},
		Refresh: StatusImage(conn, name),
		Timeout: ImageDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeImageOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitImageVersionCreated waits for a ImageVersion to return Created
func WaitImageVersionCreated(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageVersionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.ImageVersionStatusCreating,
		},
		Target:  []string{sagemaker.ImageVersionStatusCreated},
		Refresh: StatusImageVersion(conn, name),
		Timeout: ImageVersionCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeImageVersionOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitImageVersionDeleted waits for a ImageVersion to return Deleted
func WaitImageVersionDeleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageVersionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.ImageVersionStatusDeleting},
		Target:  []string{},
		Refresh: StatusImageVersion(conn, name),
		Timeout: ImageVersionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeImageVersionOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitDomainInService waits for a Domain to return InService
func WaitDomainInService(conn *sagemaker.SageMaker, domainID string) (*sagemaker.DescribeDomainOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			domainStatusNotFound,
			sagemaker.DomainStatusPending,
			sagemaker.DomainStatusUpdating,
		},
		Target:  []string{sagemaker.DomainStatusInService},
		Refresh: StatusDomain(conn, domainID),
		Timeout: DomainInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeDomainOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitDomainDeleted waits for a Domain to return Deleted
func WaitDomainDeleted(conn *sagemaker.SageMaker, domainID string) (*sagemaker.DescribeDomainOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.DomainStatusDeleting,
		},
		Target:  []string{},
		Refresh: StatusDomain(conn, domainID),
		Timeout: DomainDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeDomainOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitFeatureGroupCreated waits for a Feature Group to return Created
func WaitFeatureGroupCreated(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFeatureGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.FeatureGroupStatusCreating},
		Target:  []string{sagemaker.FeatureGroupStatusCreated},
		Refresh: StatusFeatureGroup(conn, name),
		Timeout: FeatureGroupCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeFeatureGroupOutput); ok {
		if status, reason := aws.StringValue(output.FeatureGroupStatus), aws.StringValue(output.FailureReason); status == sagemaker.FeatureGroupStatusCreateFailed && reason != "" {
			tfresource.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

// WaitFeatureGroupDeleted waits for a Feature Group to return Deleted
func WaitFeatureGroupDeleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFeatureGroupOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.FeatureGroupStatusDeleting},
		Target:  []string{},
		Refresh: StatusFeatureGroup(conn, name),
		Timeout: FeatureGroupDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeFeatureGroupOutput); ok {
		if status, reason := aws.StringValue(output.FeatureGroupStatus), aws.StringValue(output.FailureReason); status == sagemaker.FeatureGroupStatusDeleteFailed && reason != "" {
			tfresource.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

// WaitUserProfileInService waits for a UserProfile to return InService
func WaitUserProfileInService(conn *sagemaker.SageMaker, domainID, userProfileName string) (*sagemaker.DescribeUserProfileOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			userProfileStatusNotFound,
			sagemaker.UserProfileStatusPending,
			sagemaker.UserProfileStatusUpdating,
		},
		Target:  []string{sagemaker.UserProfileStatusInService},
		Refresh: StatusUserProfile(conn, domainID, userProfileName),
		Timeout: UserProfileInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeUserProfileOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitUserProfileDeleted waits for a UserProfile to return Deleted
func WaitUserProfileDeleted(conn *sagemaker.SageMaker, domainID, userProfileName string) (*sagemaker.DescribeUserProfileOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.UserProfileStatusDeleting,
		},
		Target:  []string{},
		Refresh: StatusUserProfile(conn, domainID, userProfileName),
		Timeout: UserProfileDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeUserProfileOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitAppInService waits for a App to return InService
func WaitAppInService(conn *sagemaker.SageMaker, domainID, userProfileName, appType, appName string) (*sagemaker.DescribeAppOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			appStatusNotFound,
			sagemaker.AppStatusPending,
		},
		Target:  []string{sagemaker.AppStatusInService},
		Refresh: StatusApp(conn, domainID, userProfileName, appType, appName),
		Timeout: AppInServiceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeAppOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitAppDeleted waits for a App to return Deleted
func WaitAppDeleted(conn *sagemaker.SageMaker, domainID, userProfileName, appType, appName string) (*sagemaker.DescribeAppOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.AppStatusDeleting,
		},
		Target: []string{
			sagemaker.AppStatusDeleted,
		},
		Refresh: StatusApp(conn, domainID, userProfileName, appType, appName),
		Timeout: AppDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeAppOutput); ok {
		return output, err
	}

	return nil, err
}

// WaitFlowDefinitionActive waits for a FlowDefinition to return Active
func WaitFlowDefinitionActive(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFlowDefinitionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.FlowDefinitionStatusInitializing},
		Target:  []string{sagemaker.FlowDefinitionStatusActive},
		Refresh: StatusFlowDefinition(conn, name),
		Timeout: FlowDefinitionActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeFlowDefinitionOutput); ok {
		if status, reason := aws.StringValue(output.FlowDefinitionStatus), aws.StringValue(output.FailureReason); status == sagemaker.FlowDefinitionStatusFailed && reason != "" {
			tfresource.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

// WaitFlowDefinitionDeleted waits for a FlowDefinition to return Deleted
func WaitFlowDefinitionDeleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFlowDefinitionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.FlowDefinitionStatusDeleting},
		Target:  []string{},
		Refresh: StatusFlowDefinition(conn, name),
		Timeout: FlowDefinitionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeFlowDefinitionOutput); ok {
		if status, reason := aws.StringValue(output.FlowDefinitionStatus), aws.StringValue(output.FailureReason); status == sagemaker.FlowDefinitionStatusFailed && reason != "" {
			tfresource.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

// WaitProjectDeleted waits for a FlowDefinition to return Deleted
func WaitProjectDeleted(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeProjectOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.ProjectStatusDeleteInProgress, sagemaker.ProjectStatusPending},
		Target:  []string{},
		Refresh: StatusProject(conn, name),
		Timeout: ProjectDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeProjectOutput); ok {
		if status, reason := aws.StringValue(output.ProjectStatus), aws.StringValue(output.ServiceCatalogProvisionedProductDetails.ProvisionedProductStatusMessage); status == sagemaker.ProjectStatusDeleteFailed && reason != "" {
			tfresource.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

// WaitProjectCreated waits for a Project to return Created
func WaitProjectCreated(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeProjectOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.ProjectStatusPending, sagemaker.ProjectStatusCreateInProgress},
		Target:  []string{sagemaker.ProjectStatusCreateCompleted},
		Refresh: StatusProject(conn, name),
		Timeout: ProjectCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeProjectOutput); ok {
		if status, reason := aws.StringValue(output.ProjectStatus), aws.StringValue(output.ServiceCatalogProvisionedProductDetails.ProvisionedProductStatusMessage); status == sagemaker.ProjectStatusCreateFailed && reason != "" {
			tfresource.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}

// WaitProjectUpdated waits for a Project to return Updated
func WaitProjectUpdated(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeProjectOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{sagemaker.ProjectStatusPending, sagemaker.ProjectStatusUpdateInProgress},
		Target:  []string{sagemaker.ProjectStatusUpdateCompleted},
		Refresh: StatusProject(conn, name),
		Timeout: ProjectCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*sagemaker.DescribeProjectOutput); ok {
		if status, reason := aws.StringValue(output.ProjectStatus), aws.StringValue(output.ServiceCatalogProvisionedProductDetails.ProvisionedProductStatusMessage); status == sagemaker.ProjectStatusUpdateFailed && reason != "" {
			tfresource.SetLastError(err, errors.New(reason))
		}

		return output, err
	}

	return nil, err
}
