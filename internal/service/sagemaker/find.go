package sagemaker

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// FindCodeRepositoryByName returns the code repository corresponding to the specified name.
// Returns nil if no code repository is found.
func FindCodeRepositoryByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeCodeRepositoryOutput, error) {
	input := &sagemaker.DescribeCodeRepositoryInput{
		CodeRepositoryName: aws.String(name),
	}

	output, err := conn.DescribeCodeRepository(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

// FindModelPackageGroupByName returns the Model Package Group corresponding to the specified name.
// Returns nil if no Model Package Group is found.
func FindModelPackageGroupByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeModelPackageGroupOutput, error) {
	input := &sagemaker.DescribeModelPackageGroupInput{
		ModelPackageGroupName: aws.String(name),
	}

	output, err := conn.DescribeModelPackageGroup(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

func FindModelPackageGroupPolicyByName(conn *sagemaker.SageMaker, name string) (*sagemaker.GetModelPackageGroupPolicyOutput, error) {
	input := &sagemaker.GetModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(name),
	}

	output, err := conn.GetModelPackageGroupPolicy(input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find Model Package Group") ||
		tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find resource policy") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// FindImageByName returns the Image corresponding to the specified name.
// Returns nil if no Image is found.
func FindImageByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageOutput, error) {
	input := &sagemaker.DescribeImageInput{
		ImageName: aws.String(name),
	}

	output, err := conn.DescribeImage(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

// FindImageVersionByName returns the Image Version corresponding to the specified name.
// Returns nil if no Image Version is found.
func FindImageVersionByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageVersionOutput, error) {
	input := &sagemaker.DescribeImageVersionInput{
		ImageName: aws.String(name),
	}

	output, err := conn.DescribeImageVersion(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

func FindDeviceByName(conn *sagemaker.SageMaker, deviceFleetName, deviceName string) (*sagemaker.DescribeDeviceOutput, error) {
	input := &sagemaker.DescribeDeviceInput{
		DeviceFleetName: aws.String(deviceFleetName),
		DeviceName:      aws.String(deviceName),
	}

	output, err := conn.DescribeDevice(input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No device with name") ||
		tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No device fleet with name") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// FindDeviceFleetByName returns the Device Fleet corresponding to the specified Device Fleet name.
// Returns nil if no Device Fleet is found.
func FindDeviceFleetByName(conn *sagemaker.SageMaker, id string) (*sagemaker.DescribeDeviceFleetOutput, error) {
	input := &sagemaker.DescribeDeviceFleetInput{
		DeviceFleetName: aws.String(id),
	}

	output, err := conn.DescribeDeviceFleet(input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No devicefleet with name") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// FindDomainByName returns the domain corresponding to the specified domain id.
// Returns nil if no domain is found.
func FindDomainByName(conn *sagemaker.SageMaker, domainID string) (*sagemaker.DescribeDomainOutput, error) {
	input := &sagemaker.DescribeDomainInput{
		DomainId: aws.String(domainID),
	}

	output, err := conn.DescribeDomain(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

func FindFeatureGroupByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFeatureGroupOutput, error) {
	input := &sagemaker.DescribeFeatureGroupInput{
		FeatureGroupName: aws.String(name),
	}

	output, err := conn.DescribeFeatureGroup(input)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// FindUserProfileByName returns the domain corresponding to the specified domain id.
// Returns nil if no domain is found.
func FindUserProfileByName(conn *sagemaker.SageMaker, domainID, userProfileName string) (*sagemaker.DescribeUserProfileOutput, error) {
	input := &sagemaker.DescribeUserProfileInput{
		DomainId:        aws.String(domainID),
		UserProfileName: aws.String(userProfileName),
	}

	output, err := conn.DescribeUserProfile(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

// FindAppImageConfigByName returns the App Image Config corresponding to the specified App Image Config ID.
// Returns nil if no App Image Cofnig is found.
func FindAppImageConfigByName(conn *sagemaker.SageMaker, appImageConfigID string) (*sagemaker.DescribeAppImageConfigOutput, error) {
	input := &sagemaker.DescribeAppImageConfigInput{
		AppImageConfigName: aws.String(appImageConfigID),
	}

	output, err := conn.DescribeAppImageConfig(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

// FindAppByName returns the domain corresponding to the specified domain id.
// Returns nil if no domain is found.
func FindAppByName(conn *sagemaker.SageMaker, domainID, userProfileName, appType, appName string) (*sagemaker.DescribeAppOutput, error) {
	input := &sagemaker.DescribeAppInput{
		DomainId:        aws.String(domainID),
		UserProfileName: aws.String(userProfileName),
		AppType:         aws.String(appType),
		AppName:         aws.String(appName),
	}

	output, err := conn.DescribeApp(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

func FindWorkforceByName(conn *sagemaker.SageMaker, name string) (*sagemaker.Workforce, error) {
	input := &sagemaker.DescribeWorkforceInput{
		WorkforceName: aws.String(name),
	}

	output, err := conn.DescribeWorkforce(input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No workforce") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workforce == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workforce, nil
}

func FindWorkteamByName(conn *sagemaker.SageMaker, name string) (*sagemaker.Workteam, error) {
	input := &sagemaker.DescribeWorkteamInput{
		WorkteamName: aws.String(name),
	}

	output, err := conn.DescribeWorkteam(input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "The work team") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workteam == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workteam, nil
}

func FindHumanTaskUIByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeHumanTaskUiOutput, error) {
	input := &sagemaker.DescribeHumanTaskUiInput{
		HumanTaskUiName: aws.String(name),
	}

	output, err := conn.DescribeHumanTaskUi(input)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindEndpointByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeEndpointOutput, error) {
	input := &sagemaker.DescribeEndpointInput{
		EndpointName: aws.String(name),
	}

	output, err := conn.DescribeEndpoint(input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find endpoint") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if aws.StringValue(output.EndpointStatus) == sagemaker.EndpointStatusDeleting {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindEndpointConfigByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeEndpointConfigOutput, error) {
	input := &sagemaker.DescribeEndpointConfigInput{
		EndpointConfigName: aws.String(name),
	}

	output, err := conn.DescribeEndpointConfig(input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find endpoint configuration") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindFlowDefinitionByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFlowDefinitionOutput, error) {
	input := &sagemaker.DescribeFlowDefinitionInput{
		FlowDefinitionName: aws.String(name),
	}

	output, err := conn.DescribeFlowDefinition(input)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindStudioLifecycleConfigByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeStudioLifecycleConfigOutput, error) {
	input := &sagemaker.DescribeStudioLifecycleConfigInput{
		StudioLifecycleConfigName: aws.String(name),
	}

	output, err := conn.DescribeStudioLifecycleConfig(input)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindProjectByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeProjectOutput, error) {
	input := &sagemaker.DescribeProjectInput{
		ProjectName: aws.String(name),
	}

	output, err := conn.DescribeProject(input)

	if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	status := aws.StringValue(output.ProjectStatus)
	if status == sagemaker.ProjectStatusDeleteCompleted {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNotebookInstanceByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	input := &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(name),
	}

	output, err := conn.DescribeNotebookInstance(input)

	if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
