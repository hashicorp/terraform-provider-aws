package sagemaker

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	notebookInstanceStatusNotFound  = "NotFound"
	imageStatusNotFound             = "NotFound"
	imageStatusFailed               = "Failed"
	imageVersionStatusNotFound      = "NotFound"
	imageVersionStatusFailed        = "Failed"
	domainStatusNotFound            = "NotFound"
	userProfileStatusNotFound       = "NotFound"
	modelPackageGroupStatusNotFound = "NotFound"
	appStatusNotFound               = "NotFound"
)

// StatusNotebookInstance fetches the NotebookInstance and its Status
func StatusNotebookInstance(conn *sagemaker.SageMaker, notebookName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeNotebookInstanceInput{
			NotebookInstanceName: aws.String(notebookName),
		}

		output, err := conn.DescribeNotebookInstance(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
			return nil, notebookInstanceStatusNotFound, nil
		}

		if err != nil {
			return nil, sagemaker.NotebookInstanceStatusFailed, err
		}

		if output == nil {
			return nil, notebookInstanceStatusNotFound, nil
		}

		return output, aws.StringValue(output.NotebookInstanceStatus), nil
	}
}

// StatusModelPackageGroup fetches the ModelPackageGroup and its Status
func StatusModelPackageGroup(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeModelPackageGroupInput{
			ModelPackageGroupName: aws.String(name),
		}

		output, err := conn.DescribeModelPackageGroup(input)

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
func StatusImage(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeImageInput{
			ImageName: aws.String(name),
		}

		output, err := conn.DescribeImage(input)

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
func StatusImageVersion(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeImageVersionInput{
			ImageName: aws.String(name),
		}

		output, err := conn.DescribeImageVersion(input)

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
func StatusDomain(conn *sagemaker.SageMaker, domainID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeDomainInput{
			DomainId: aws.String(domainID),
		}

		output, err := conn.DescribeDomain(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
			return nil, sagemaker.UserProfileStatusFailed, nil
		}

		if err != nil {
			return nil, sagemaker.DomainStatusFailed, err
		}

		if output == nil {
			return nil, domainStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func StatusFeatureGroup(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFeatureGroupByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.FeatureGroupStatus), nil
	}
}

func StatusFlowDefinition(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFlowDefinitionByName(conn, name)

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
func StatusUserProfile(conn *sagemaker.SageMaker, domainID, userProfileName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeUserProfileInput{
			DomainId:        aws.String(domainID),
			UserProfileName: aws.String(userProfileName),
		}

		output, err := conn.DescribeUserProfile(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
			return nil, userProfileStatusNotFound, nil
		}

		if err != nil {
			return nil, sagemaker.UserProfileStatusFailed, err
		}

		if output == nil {
			return nil, userProfileStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// StatusApp fetches the App and its Status
func StatusApp(conn *sagemaker.SageMaker, domainID, userProfileName, appType, appName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeAppInput{
			DomainId:        aws.String(domainID),
			UserProfileName: aws.String(userProfileName),
			AppType:         aws.String(appType),
			AppName:         aws.String(appName),
		}

		output, err := conn.DescribeApp(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
			return nil, appStatusNotFound, nil
		}

		if err != nil {
			return nil, sagemaker.AppStatusFailed, err
		}

		if output == nil {
			return nil, appStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func StatusProject(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindProjectByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ProjectStatus), nil
	}
}
