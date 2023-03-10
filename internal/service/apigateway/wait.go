package apigateway

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time for VpcLink to become available
	vpcLinkAvailableTimeout = 20 * time.Minute

	// Maximum amount of time for VpcLink to delete
	vpcLinkDeleteTimeout = 20 * time.Minute

	// Maximum amount of time for Stage Cache to be available
	stageCacheAvailableTimeout = 90 * time.Minute

	// Maximum amount of time for Stage Cache to update
	stageCacheUpdateTimeout = 30 * time.Minute
)

func waitVPCLinkAvailable(ctx context.Context, conn *apigateway.APIGateway, vpcLinkId string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{apigateway.VpcLinkStatusPending},
		Target:     []string{apigateway.VpcLinkStatusAvailable},
		Refresh:    vpcLinkStatus(ctx, conn, vpcLinkId),
		Timeout:    vpcLinkAvailableTimeout,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitVPCLinkDeleted(ctx context.Context, conn *apigateway.APIGateway, vpcLinkId string) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			apigateway.VpcLinkStatusPending,
			apigateway.VpcLinkStatusAvailable,
			apigateway.VpcLinkStatusDeleting,
		},
		Target:     []string{},
		Timeout:    vpcLinkDeleteTimeout,
		MinTimeout: 1 * time.Second,
		Refresh:    vpcLinkStatus(ctx, conn, vpcLinkId),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitStageCacheAvailable(ctx context.Context, conn *apigateway.APIGateway, restApiId, name string) (*apigateway.Stage, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			apigateway.CacheClusterStatusCreateInProgress,
			apigateway.CacheClusterStatusDeleteInProgress,
			apigateway.CacheClusterStatusFlushInProgress,
		},
		Target:  []string{apigateway.CacheClusterStatusAvailable},
		Refresh: stageCacheStatus(ctx, conn, restApiId, name),
		Timeout: stageCacheAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*apigateway.Stage); ok {
		return output, err
	}

	return nil, err
}

func waitStageCacheUpdated(ctx context.Context, conn *apigateway.APIGateway, restApiId, name string) (*apigateway.Stage, error) {
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
		Refresh: stageCacheStatus(ctx, conn, restApiId, name),
		Timeout: stageCacheUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*apigateway.Stage); ok {
		return output, err
	}

	return nil, err
}
