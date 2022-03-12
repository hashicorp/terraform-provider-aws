package apigateway

import (
	"time"

	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time for VpcLink to become available
	apiGatewayVPCLinkAvailableTimeout = 20 * time.Minute

	// Maximum amount of time for VpcLink to delete
	apiGatewayVPCLinkDeleteTimeout = 20 * time.Minute

	// Maximum amount of time for Stage Cache to be available
	apiGatewayStageCacheAvailableTimeout = 90 * time.Minute

	// Maximum amount of time for Stage Cache to update
	apiGatewayStageCacheUpdateTimeout = 30 * time.Minute
)

func waitAPIGatewayVPCLinkAvailable(conn *apigateway.APIGateway, vpcLinkId string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{apigateway.VpcLinkStatusPending},
		Target:     []string{apigateway.VpcLinkStatusAvailable},
		Refresh:    apiGatewayVpcLinkStatus(conn, vpcLinkId),
		Timeout:    apiGatewayVPCLinkAvailableTimeout,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitAPIGatewayVPCLinkDeleted(conn *apigateway.APIGateway, vpcLinkId string) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			apigateway.VpcLinkStatusPending,
			apigateway.VpcLinkStatusAvailable,
			apigateway.VpcLinkStatusDeleting,
		},
		Target:     []string{},
		Timeout:    apiGatewayVPCLinkDeleteTimeout,
		MinTimeout: 1 * time.Second,
		Refresh:    apiGatewayVpcLinkStatus(conn, vpcLinkId),
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitStageCacheAvailable(conn *apigateway.APIGateway, restApiId, name string) (*apigateway.Stage, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			apigateway.CacheClusterStatusCreateInProgress,
			apigateway.CacheClusterStatusDeleteInProgress,
			apigateway.CacheClusterStatusFlushInProgress,
		},
		Target:  []string{apigateway.CacheClusterStatusAvailable},
		Refresh: stageCacheStatus(conn, restApiId, name),
		Timeout: apiGatewayStageCacheAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*apigateway.Stage); ok {
		return output, err
	}

	return nil, err
}

func waitStageCacheUpdated(conn *apigateway.APIGateway, restApiId, name string) (*apigateway.Stage, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			apigateway.CacheClusterStatusCreateInProgress,
			apigateway.CacheClusterStatusFlushInProgress,
		},
		Target: []string{
			apigateway.CacheClusterStatusAvailable,
			// There's an AWS API bug (raised & confirmed in Sep 2016 by support)
			// which causes the stage to remain in deletion state forever
			apigateway.CacheClusterStatusDeleteInProgress,
		},
		Refresh: stageCacheStatus(conn, restApiId, name),
		Timeout: apiGatewayStageCacheUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*apigateway.Stage); ok {
		return output, err
	}

	return nil, err
}
