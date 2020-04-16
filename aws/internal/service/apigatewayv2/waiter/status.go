package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// DeploymentStatus fetches the Deployment and its Status
func DeploymentStatus(conn *apigatewayv2.ApiGatewayV2, apiId, deploymentId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apigatewayv2.GetDeploymentInput{
			ApiId:        aws.String(apiId),
			DeploymentId: aws.String(deploymentId),
		}

		output, err := conn.GetDeployment(input)

		if err != nil {
			return nil, apigatewayv2.DeploymentStatusFailed, err
		}

		// Error messages can also be contained in the response with FAILED status

		if aws.StringValue(output.DeploymentStatus) == apigatewayv2.DeploymentStatusFailed {
			return output, apigatewayv2.DeploymentStatusFailed, fmt.Errorf("%s", aws.StringValue(output.DeploymentStatusMessage))
		}

		return output, aws.StringValue(output.DeploymentStatus), nil
	}
}
