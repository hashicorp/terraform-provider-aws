package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/terraform"
)

func resourceAwsDxGatewayAssociationMigrateState(v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found Direct Connect gateway association state v0; migrating to v1")
		return migrateDxGatewayAssociationStateV0toV1(is, meta)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateDxGatewayAssociationStateV0toV1(is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	conn := meta.(*AWSClient).dxconn

	// dx_gateway_association_id was introduced in v2.8.0. Handle the case where it's not yet present.
	if _, ok := is.Attributes["dx_gateway_association_id"]; !ok {
		resp, err := conn.DescribeDirectConnectGatewayAssociations(&directconnect.DescribeDirectConnectGatewayAssociationsInput{
			DirectConnectGatewayId: aws.String(is.Attributes["dx_gateway_id"]),
			VirtualGatewayId:       aws.String(is.Attributes["vpn_gateway_id"]),
		})
		if err != nil {
			return nil, err
		}

		if len(resp.DirectConnectGatewayAssociations) == 0 {
			return nil, fmt.Errorf("Direct Connect gateway association not found, remove from state using 'terraform state rm'")
		}

		is.Attributes["dx_gateway_association_id"] = aws.StringValue(resp.DirectConnectGatewayAssociations[0].AssociationId)
	}

	return is, nil
}
