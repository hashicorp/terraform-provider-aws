package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// NotebookInstanceStatus fetches the NotebookInstance and its Status
func NotebookInstanceStatus(conn *sagemaker.SageMaker, notebookName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &sagemaker.DescribeNotebookInstanceInput{
			NotebookInstanceName: aws.String(notebookName),
		}

		output, err := conn.DescribeNotebookInstance(input)
		if err != nil {
			return nil, sagemaker.NotebookInstanceStatusFailed, err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.NotebookInstanceStatus), nil
	}
}
