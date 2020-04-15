package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	// Maximum amount of time to wait for a Deployment to return Deployed
	DeploymentDeployedTimeout = 5 * time.Minute
)

// DeploymentDeployed waits for a Deployment to return Deployed
func DeploymentDeployed(conn *apigatewayv2.ApiGatewayV2, apiId, deploymentId string) (*apigatewayv2.GetDeploymentOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apigatewayv2.DeploymentStatusPending},
		Target:  []string{apigatewayv2.DeploymentStatusDeployed},
		Refresh: DeploymentStatus(conn, apiId, deploymentId),
		Timeout: DeploymentDeployedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*apigatewayv2.GetDeploymentOutput); ok {
		return v, err
	}

	return nil, err
}
