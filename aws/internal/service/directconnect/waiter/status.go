package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/directconnect/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func GatewayState(conn *directconnect.DirectConnect, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.GatewayByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.DirectConnectGatewayState), nil
	}
}

func GatewayAssociationState(conn *directconnect.DirectConnect, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.GatewayAssociationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.AssociationState), nil
	}
}
