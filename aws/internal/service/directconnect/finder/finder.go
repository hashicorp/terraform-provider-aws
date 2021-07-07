package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func GatewayAssociationByID(conn *directconnect.DirectConnect, id string) (*directconnect.GatewayAssociation, error) {
	input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
		AssociationId: aws.String(id),
	}

	output, err := conn.DescribeDirectConnectGatewayAssociations(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DirectConnectGatewayAssociations) == 0 || output.DirectConnectGatewayAssociations[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	gatewayAssociation := output.DirectConnectGatewayAssociations[0]

	if state := aws.StringValue(gatewayAssociation.AssociationState); state == directconnect.GatewayAssociationStateDisassociated {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(gatewayAssociation.AssociationId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return gatewayAssociation, nil
}
