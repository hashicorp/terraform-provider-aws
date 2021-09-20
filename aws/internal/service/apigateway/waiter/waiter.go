package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time for VpcLink to become available
	ApiGatewayVpcLinkAvailableTimeout = 20 * time.Minute

	// Maximum amount of time for VpcLink to delete
	ApiGatewayVpcLinkDeleteTimeout = 20 * time.Minute
)

func ApiGatewayVpcLinkAvailable(conn *apigateway.APIGateway, vpcLinkId string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{apigateway.VpcLinkStatusPending},
		Target:     []string{apigateway.VpcLinkStatusAvailable},
		Refresh:    apiGatewayVpcLinkStatus(conn, vpcLinkId),
		Timeout:    ApiGatewayVpcLinkAvailableTimeout,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func ApiGatewayVpcLinkDeleted(conn *apigateway.APIGateway, vpcLinkId string) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			apigateway.VpcLinkStatusPending,
			apigateway.VpcLinkStatusAvailable,
			apigateway.VpcLinkStatusDeleting,
		},
		Target:     []string{},
		Timeout:    ApiGatewayVpcLinkDeleteTimeout,
		MinTimeout: 1 * time.Second,
		Refresh:    apiGatewayVpcLinkStatus(conn, vpcLinkId),
	}

	_, err := stateConf.WaitForState()

	return err
}
