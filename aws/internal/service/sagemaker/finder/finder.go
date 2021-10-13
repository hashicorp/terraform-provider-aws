package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfsagemaker "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// CodeRepositoryByName returns the code repository corresponding to the specified name.
// Returns nil if no code repository is found.
func CodeRepositoryByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeCodeRepositoryOutput, error) {
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

// ModelPackageGroupByName returns the Model Package Group corresponding to the specified name.
// Returns nil if no Model Package Group is found.
func ModelPackageGroupByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeModelPackageGroupOutput, error) {
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

func ModelPackageGroupPolicyByName(conn *sagemaker.SageMaker, name string) (*sagemaker.GetModelPackageGroupPolicyOutput, error) {
	input := &sagemaker.GetModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(name),
	}

	output, err := conn.GetModelPackageGroupPolicy(input)

	if tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "Cannot find Model Package Group") ||
		tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "Cannot find resource policy") {
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

// ImageByName returns the Image corresponding to the specified name.
// Returns nil if no Image is found.
func ImageByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageOutput, error) {
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

// ImageVersionByName returns the Image Version corresponding to the specified name.
// Returns nil if no Image Version is found.
func ImageVersionByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageVersionOutput, error) {
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

// DeviceFleetByName returns the Device Fleet corresponding to the specified Device Fleet name.
// Returns nil if no Device Fleet is found.
func DeviceFleetByName(conn *sagemaker.SageMaker, id string) (*sagemaker.DescribeDeviceFleetOutput, error) {
	input := &sagemaker.DescribeDeviceFleetInput{
		DeviceFleetName: aws.String(id),
	}

	output, err := conn.DescribeDeviceFleet(input)

	if tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "No devicefleet with name") {
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

// DomainByName returns the domain corresponding to the specified domain id.
// Returns nil if no domain is found.
func DomainByName(conn *sagemaker.SageMaker, domainID string) (*sagemaker.DescribeDomainOutput, error) {
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

func FeatureGroupByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFeatureGroupOutput, error) {
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

// UserProfileByName returns the domain corresponding to the specified domain id.
// Returns nil if no domain is found.
func UserProfileByName(conn *sagemaker.SageMaker, domainID, userProfileName string) (*sagemaker.DescribeUserProfileOutput, error) {
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

// AppImageConfigByName returns the App Image Config corresponding to the specified App Image Config ID.
// Returns nil if no App Image Cofnig is found.
func AppImageConfigByName(conn *sagemaker.SageMaker, appImageConfigID string) (*sagemaker.DescribeAppImageConfigOutput, error) {
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

// AppByName returns the domain corresponding to the specified domain id.
// Returns nil if no domain is found.
func AppByName(conn *sagemaker.SageMaker, domainID, userProfileName, appType, appName string) (*sagemaker.DescribeAppOutput, error) {
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

func WorkforceByName(conn *sagemaker.SageMaker, name string) (*sagemaker.Workforce, error) {
	input := &sagemaker.DescribeWorkforceInput{
		WorkforceName: aws.String(name),
	}

	output, err := conn.DescribeWorkforce(input)

	if tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "No workforce") {
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

func WorkteamByName(conn *sagemaker.SageMaker, name string) (*sagemaker.Workteam, error) {
	input := &sagemaker.DescribeWorkteamInput{
		WorkteamName: aws.String(name),
	}

	output, err := conn.DescribeWorkteam(input)

	if tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "The work team") {
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

func HumanTaskUiByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeHumanTaskUiOutput, error) {
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

func EndpointConfigByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeEndpointConfigOutput, error) {
	input := &sagemaker.DescribeEndpointConfigInput{
		EndpointConfigName: aws.String(name),
	}

	output, err := conn.DescribeEndpointConfig(input)

	if tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "Could not find endpoint configuration") {
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

func FlowDefinitionByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFlowDefinitionOutput, error) {
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

func StudioLifecycleConfigByName(conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeStudioLifecycleConfigOutput, error) {
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
