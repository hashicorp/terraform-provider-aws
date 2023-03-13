package apigatewayv2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
func WaitDeploymentDeployed(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, apiId, deploymentId string) (*apigatewayv2.GetDeploymentOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apigatewayv2.DeploymentStatusPending},
		Target:  []string{apigatewayv2.DeploymentStatusDeployed},
		Refresh: StatusDeployment(ctx, conn, apiId, deploymentId),
		Timeout: DeploymentDeployedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*apigatewayv2.GetDeploymentOutput); ok {
		return v, err
	}

	return nil, err
}

func WaitDomainNameAvailable(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, name string, timeout time.Duration) (*apigatewayv2.GetDomainNameOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apigatewayv2.DomainNameStatusUpdating},
		Target:  []string{apigatewayv2.DomainNameStatusAvailable},
		Refresh: StatusDomainName(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*apigatewayv2.GetDomainNameOutput); ok {
		return v, err
	}

	return nil, err
}

// WaitVPCLinkAvailable waits for a VPC Link to return Available
func WaitVPCLinkAvailable(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, vpcLinkId string) (*apigatewayv2.GetVpcLinkOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apigatewayv2.VpcLinkStatusPending},
		Target:  []string{apigatewayv2.VpcLinkStatusAvailable},
		Refresh: StatusVPCLink(ctx, conn, vpcLinkId),
		Timeout: VPCLinkAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*apigatewayv2.GetVpcLinkOutput); ok {
		return v, err
	}

	return nil, err
}

// WaitVPCLinkAvailable waits for a VPC Link to return Deleted
func WaitVPCLinkDeleted(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, vpcLinkId string) (*apigatewayv2.GetVpcLinkOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{apigatewayv2.VpcLinkStatusDeleting},
		Target:  []string{apigatewayv2.VpcLinkStatusFailed},
		Refresh: StatusVPCLink(ctx, conn, vpcLinkId),
		Timeout: VPCLinkDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*apigatewayv2.GetVpcLinkOutput); ok {
		return v, err
	}

	return nil, err
}
