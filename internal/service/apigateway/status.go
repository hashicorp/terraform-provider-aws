package apigateway

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func apiGatewayVpcLinkStatus(conn *apigateway.APIGateway, vpcLinkId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.GetVpcLink(&apigateway.GetVpcLinkInput{
			VpcLinkId: aws.String(vpcLinkId),
		})
		if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		// Error messages can also be contained in the response with FAILED status
		if aws.StringValue(output.Status) == apigateway.VpcLinkStatusFailed {
			return output, apigateway.VpcLinkStatusFailed, fmt.Errorf("%s: %s", apigateway.VpcLinkStatusFailed, aws.StringValue(output.StatusMessage))
		}

		return output, aws.StringValue(output.Status), nil
	}
}
