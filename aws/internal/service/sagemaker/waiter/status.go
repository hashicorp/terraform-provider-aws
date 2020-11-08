package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	SagemakerNotebookInstanceStatusNotFound = "NotFound"
	SagemakerImageStatusNotFound            = "NotFound"
	SagemakerImageStatusFailed              = "Failed"
)

// NotebookInstanceStatus fetches the NotebookInstance and its Status
func NotebookInstanceStatus(conn *sagemaker.SageMaker, notebookName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeNotebookInstanceInput{
			NotebookInstanceName: aws.String(notebookName),
		}

		output, err := conn.DescribeNotebookInstance(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
			return nil, SagemakerNotebookInstanceStatusNotFound, nil
		}

		if err != nil {
			return nil, sagemaker.NotebookInstanceStatusFailed, err
		}

		if output == nil {
			return nil, SagemakerNotebookInstanceStatusNotFound, nil
		}

		return output, aws.StringValue(output.NotebookInstanceStatus), nil
	}
}

// ImageStatus fetches the Image and its Status
func ImageStatus(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeImageInput{
			ImageName: aws.String(name),
		}

		output, err := conn.DescribeImage(input)

		if tfawserr.ErrMessageContains(err, "ValidationException", "RecordNotFound") {
			return nil, SagemakerImageStatusNotFound, nil
		}

		if err != nil {
			return nil, SagemakerImageStatusFailed, err
		}

		if output == nil {
			return nil, SagemakerImageStatusNotFound, nil
		}

		return output, aws.StringValue(output.ImageStatus), nil
	}
}
