// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// FindCodeRepositoryByName returns the code repository corresponding to the specified name.
// Returns nil if no code repository is found.
func FindCodeRepositoryByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeCodeRepositoryOutput, error) {
	input := &sagemaker.DescribeCodeRepositoryInput{
		CodeRepositoryName: aws.String(name),
	}

	output, err := conn.DescribeCodeRepositoryWithContext(ctx, input)

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
func FindModelPackageGroupByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeModelPackageGroupOutput, error) {
	input := &sagemaker.DescribeModelPackageGroupInput{
		ModelPackageGroupName: aws.String(name),
	}

	output, err := conn.DescribeModelPackageGroupWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

func FindModelPackageGroupPolicyByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.GetModelPackageGroupPolicyOutput, error) {
	input := &sagemaker.GetModelPackageGroupPolicyInput{
		ModelPackageGroupName: aws.String(name),
	}

	output, err := conn.GetModelPackageGroupPolicyWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find Model Package Group") ||
		tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Cannot find resource policy") {
		return nil, &retry.NotFoundError{
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
func FindImageByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageOutput, error) {
	input := &sagemaker.DescribeImageInput{
		ImageName: aws.String(name),
	}

	output, err := conn.DescribeImageWithContext(ctx, input)

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
func FindImageVersionByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeImageVersionOutput, error) {
	input := &sagemaker.DescribeImageVersionInput{
		ImageName: aws.String(name),
	}

	output, err := conn.DescribeImageVersionWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

func FindDeviceByName(ctx context.Context, conn *sagemaker.SageMaker, deviceFleetName, deviceName string) (*sagemaker.DescribeDeviceOutput, error) {
	input := &sagemaker.DescribeDeviceInput{
		DeviceFleetName: aws.String(deviceFleetName),
		DeviceName:      aws.String(deviceName),
	}

	output, err := conn.DescribeDeviceWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No device with name") ||
		tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No device fleet with name") {
		return nil, &retry.NotFoundError{
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
func FindDeviceFleetByName(ctx context.Context, conn *sagemaker.SageMaker, id string) (*sagemaker.DescribeDeviceFleetOutput, error) {
	input := &sagemaker.DescribeDeviceFleetInput{
		DeviceFleetName: aws.String(id),
	}

	output, err := conn.DescribeDeviceFleetWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No devicefleet with name") {
		return nil, &retry.NotFoundError{
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
func FindDomainByName(ctx context.Context, conn *sagemaker.SageMaker, domainID string) (*sagemaker.DescribeDomainOutput, error) {
	input := &sagemaker.DescribeDomainInput{
		DomainId: aws.String(domainID),
	}

	output, err := conn.DescribeDomainWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return nil, &retry.NotFoundError{
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

func FindFeatureGroupByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFeatureGroupOutput, error) {
	input := &sagemaker.DescribeFeatureGroupInput{
		FeatureGroupName: aws.String(name),
	}

	output, err := conn.DescribeFeatureGroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return nil, &retry.NotFoundError{
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
func FindUserProfileByName(ctx context.Context, conn *sagemaker.SageMaker, domainID, userProfileName string) (*sagemaker.DescribeUserProfileOutput, error) {
	input := &sagemaker.DescribeUserProfileInput{
		DomainId:        aws.String(domainID),
		UserProfileName: aws.String(userProfileName),
	}

	output, err := conn.DescribeUserProfileWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return nil, &retry.NotFoundError{
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

// FindAppImageConfigByName returns the App Image Config corresponding to the specified App Image Config ID.
// Returns nil if no App Image Cofnig is found.
func FindAppImageConfigByName(ctx context.Context, conn *sagemaker.SageMaker, appImageConfigID string) (*sagemaker.DescribeAppImageConfigOutput, error) {
	input := &sagemaker.DescribeAppImageConfigInput{
		AppImageConfigName: aws.String(appImageConfigID),
	}

	output, err := conn.DescribeAppImageConfigWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output, nil
}

func listAppsByName(ctx context.Context, conn *sagemaker.SageMaker, domainID, userProfileOrSpaceName, appType, appName string) (*sagemaker.AppDetails, error) {
	input := &sagemaker.ListAppsInput{
		DomainIdEquals: aws.String(domainID),
	}
	var output []*sagemaker.AppDetails

	err := conn.ListAppsPagesWithContext(ctx, input, func(page *sagemaker.ListAppsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Apps {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		if aws.StringValue(v.AppName) == appName && aws.StringValue(v.AppType) == appType && (aws.StringValue(v.SpaceName) == userProfileOrSpaceName || aws.StringValue(v.UserProfileName) == userProfileOrSpaceName) {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func FindAppByName(ctx context.Context, conn *sagemaker.SageMaker, domainID, userProfileOrSpaceName, appType, appName string) (*sagemaker.DescribeAppOutput, error) {
	foundApp, err := listAppsByName(ctx, conn, domainID, userProfileOrSpaceName, appType, appName)

	if err != nil {
		return nil, err
	}

	input := &sagemaker.DescribeAppInput{
		AppName:  aws.String(appName),
		AppType:  aws.String(appType),
		DomainId: aws.String(domainID),
	}
	if foundApp.SpaceName != nil {
		input.SpaceName = foundApp.SpaceName
	}
	if foundApp.UserProfileName != nil {
		input.UserProfileName = foundApp.UserProfileName
	}

	output, err := conn.DescribeAppWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "RecordNotFound") {
		return nil, &retry.NotFoundError{
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

	if state := aws.StringValue(output.Status); state == sagemaker.AppStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return output, nil
}

func FindWorkforceByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.Workforce, error) {
	input := &sagemaker.DescribeWorkforceInput{
		WorkforceName: aws.String(name),
	}

	output, err := conn.DescribeWorkforceWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "No workforce") {
		return nil, &retry.NotFoundError{
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

func FindWorkteamByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.Workteam, error) {
	input := &sagemaker.DescribeWorkteamInput{
		WorkteamName: aws.String(name),
	}

	output, err := conn.DescribeWorkteamWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "The work team") {
		return nil, &retry.NotFoundError{
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

func FindHumanTaskUIByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeHumanTaskUiOutput, error) {
	input := &sagemaker.DescribeHumanTaskUiInput{
		HumanTaskUiName: aws.String(name),
	}

	output, err := conn.DescribeHumanTaskUiWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return nil, &retry.NotFoundError{
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

func FindEndpointByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeEndpointOutput, error) {
	input := &sagemaker.DescribeEndpointInput{
		EndpointName: aws.String(name),
	}

	output, err := conn.DescribeEndpointWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find endpoint") {
		return nil, &retry.NotFoundError{
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

func FindEndpointConfigByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeEndpointConfigOutput, error) {
	input := &sagemaker.DescribeEndpointConfigInput{
		EndpointConfigName: aws.String(name),
	}

	output, err := conn.DescribeEndpointConfigWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "Could not find endpoint configuration") {
		return nil, &retry.NotFoundError{
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

func FindFlowDefinitionByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeFlowDefinitionOutput, error) {
	input := &sagemaker.DescribeFlowDefinitionInput{
		FlowDefinitionName: aws.String(name),
	}

	output, err := conn.DescribeFlowDefinitionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return nil, &retry.NotFoundError{
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

func FindStudioLifecycleConfigByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeStudioLifecycleConfigOutput, error) {
	input := &sagemaker.DescribeStudioLifecycleConfigInput{
		StudioLifecycleConfigName: aws.String(name),
	}

	output, err := conn.DescribeStudioLifecycleConfigWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
		return nil, &retry.NotFoundError{
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

func FindProjectByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeProjectOutput, error) {
	input := &sagemaker.DescribeProjectInput{
		ProjectName: aws.String(name),
	}

	output, err := conn.DescribeProjectWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
		return nil, &retry.NotFoundError{
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
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNotebookInstanceByName(ctx context.Context, conn *sagemaker.SageMaker, name string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	input := &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(name),
	}

	output, err := conn.DescribeNotebookInstanceWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
		return nil, &retry.NotFoundError{
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

func FindSpaceByName(ctx context.Context, conn *sagemaker.SageMaker, domainId, name string) (*sagemaker.DescribeSpaceOutput, error) {
	input := &sagemaker.DescribeSpaceInput{
		SpaceName: aws.String(name),
		DomainId:  aws.String(domainId),
	}

	output, err := conn.DescribeSpaceWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
		return nil, &retry.NotFoundError{
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
