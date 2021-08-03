package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func DeploymentStatus(conn *appconfig.AppConfig, appID, envID string, deployNum int64) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &appconfig.GetDeploymentInput{
			ApplicationId:    aws.String(appID),
			DeploymentNumber: aws.Int64(deployNum),
			EnvironmentId:    aws.String(envID),
		}

		output, err := conn.GetDeployment(input)

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.StringValue(output.State), nil
	}
}
