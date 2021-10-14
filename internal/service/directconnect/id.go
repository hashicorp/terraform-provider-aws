package directconnect

import (
	"fmt"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func GatewayAssociationCreateResourceID(directConnectGatewayID, associatedGatewayID string) string {
	return fmt.Sprintf("ga-%s%s", directConnectGatewayID, associatedGatewayID)
}
