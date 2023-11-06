// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	notebookInstanceStatusNotFound  = "NotFound"
	imageStatusNotFound             = "NotFound"
	imageStatusFailed               = "Failed"
	imageVersionStatusNotFound      = "NotFound"
	imageVersionStatusFailed        = "Failed"
	modelPackageGroupStatusNotFound = "NotFound"
)

// StatusNotebookInstance fetches the NotebookInstance and its Status
func StatusNotebookInstance(ctx context.Context, conn *sagemaker.SageMaker, notebookName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNotebookInstanceByName(ctx, conn, notebookName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.NotebookInstanceStatus), nil
	}
}

// StatusModelPackageGroup fetches the ModelPackageGroup and its Status
func StatusModelPackageGroup(ctx context.Context, conn *sagemaker.SageMaker, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeModelPackageGroupInput{
			ModelPackageGroupName: aws.String(name),
		}

		output, err := conn.DescribeModelPackageGroupWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") {
			return nil, modelPackageGroupStatusNotFound, nil
		}

		if err != nil {
			return nil, sagemaker.ModelPackageGroupStatusFailed, err
		}

		if output == nil {
			return nil, modelPackageGroupStatusNotFound, nil
		}

		return output, aws.StringValue(output.ModelPackageGroupStatus), nil
	}
}

// StatusImage fetches the Image and its Status
func StatusImage(ctx context.Context, conn *sagemaker.SageMaker, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeImageInput{
			ImageName: aws.String(name),
		}

		output, err := conn.DescribeImageWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "No Image with the name") {
			return nil, imageStatusNotFound, nil
		}

		if err != nil {
			return nil, imageStatusFailed, err
		}

		if output == nil {
			return nil, imageStatusNotFound, nil
		}

		if aws.StringValue(output.ImageStatus) == sagemaker.ImageStatusCreateFailed {
			return output, sagemaker.ImageStatusCreateFailed, fmt.Errorf("%s", aws.StringValue(output.FailureReason))
		}

		return output, aws.StringValue(output.ImageStatus), nil
	}
}

// StatusImageVersion fetches the ImageVersion and its Status
func StatusImageVersion(ctx context.Context, conn *sagemaker.SageMaker, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeImageVersionInput{
			ImageName: aws.String(name),
		}

		output, err := conn.DescribeImageVersionWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, sagemaker.ErrCodeResourceNotFound, "No ImageVersion with the name") {
			return nil, imageVersionStatusNotFound, nil
		}

		if err != nil {
			return nil, imageVersionStatusFailed, err
		}

		if output == nil {
			return nil, imageVersionStatusNotFound, nil
		}

		if aws.StringValue(output.ImageVersionStatus) == sagemaker.ImageVersionStatusCreateFailed {
			return output, sagemaker.ImageVersionStatusCreateFailed, fmt.Errorf("%s", aws.StringValue(output.FailureReason))
		}

		return output, aws.StringValue(output.ImageVersionStatus), nil
	}
}

// StatusDomain fetches the Domain and its Status
func StatusDomain(ctx context.Context, conn *sagemaker.SageMaker, domainID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDomainByName(ctx, conn, domainID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func StatusFeatureGroup(ctx context.Context, conn *sagemaker.SageMaker, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFeatureGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.FeatureGroupStatus), nil
	}
}

func StatusFlowDefinition(ctx context.Context, conn *sagemaker.SageMaker, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFlowDefinitionByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.FlowDefinitionStatus), nil
	}
}

// StatusUserProfile fetches the UserProfile and its Status
func StatusUserProfile(ctx context.Context, conn *sagemaker.SageMaker, domainID, userProfileName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindUserProfileByName(ctx, conn, domainID, userProfileName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// StatusApp fetches the App and its Status
func StatusApp(ctx context.Context, conn *sagemaker.SageMaker, domainID, userProfileOrSpaceName, appType, appName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAppByName(ctx, conn, domainID, userProfileOrSpaceName, appType, appName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func StatusProject(ctx context.Context, conn *sagemaker.SageMaker, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindProjectByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ProjectStatus), nil
	}
}

func StatusWorkforce(ctx context.Context, conn *sagemaker.SageMaker, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindWorkforceByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func StatusSpace(ctx context.Context, conn *sagemaker.SageMaker, domainId, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSpaceByName(ctx, conn, domainId, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func StatusMonitoringSchedule(ctx context.Context, conn *sagemaker.SageMaker, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindMonitoringScheduleByName(ctx, conn, name)

		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.MonitoringScheduleStatus), nil
	}
}
