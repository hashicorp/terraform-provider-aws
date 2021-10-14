package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for a Deployment to return Deployed
	DeploymentDeployedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a VPC Link to return Available
	VPCLinkAvailableTimeout = 10 * time.Minute

	// Maximum amount of time to wait for a VPC Link to return Deleted
	VPCLinkDeletedTimeout = 10 * time.Minute
)

// WaitDeploymentDeployed waits for a Deployment to return Deployed
func WaitDeploymentDeployed(conn *apigatewayv2.ApiGatewayV2, apiId, deploymentId string) (*apigatewayv2.GetDeploymentOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apigatewayv2.DeploymentStatusPending},
		Target:  []string{apigatewayv2.DeploymentStatusDeployed},
		Refresh: StatusDeployment(conn, apiId, deploymentId),
		Timeout: DeploymentDeployedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*apigatewayv2.GetDeploymentOutput); ok {
		return v, err
	}

	return nil, err
}

func WaitDomainNameAvailable(conn *apigatewayv2.ApiGatewayV2, name string, timeout time.Duration) (*apigatewayv2.GetDomainNameOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apigatewayv2.DomainNameStatusUpdating},
		Target:  []string{apigatewayv2.DomainNameStatusAvailable},
		Refresh: StatusDomainName(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*apigatewayv2.GetDomainNameOutput); ok {
		return v, err
	}

	return nil, err
}

// WaitVPCLinkAvailable waits for a VPC Link to return Available
func WaitVPCLinkAvailable(conn *apigatewayv2.ApiGatewayV2, vpcLinkId string) (*apigatewayv2.GetVpcLinkOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apigatewayv2.VpcLinkStatusPending},
		Target:  []string{apigatewayv2.VpcLinkStatusAvailable},
		Refresh: StatusVPCLink(conn, vpcLinkId),
		Timeout: VPCLinkAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*apigatewayv2.GetVpcLinkOutput); ok {
		return v, err
	}

	return nil, err
}

// WaitVPCLinkAvailable waits for a VPC Link to return Deleted
func WaitVPCLinkDeleted(conn *apigatewayv2.ApiGatewayV2, vpcLinkId string) (*apigatewayv2.GetVpcLinkOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apigatewayv2.VpcLinkStatusDeleting},
		Target:  []string{apigatewayv2.VpcLinkStatusFailed},
		Refresh: StatusVPCLink(conn, vpcLinkId),
		Timeout: VPCLinkDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*apigatewayv2.GetVpcLinkOutput); ok {
		return v, err
	}

	return nil, err
}
