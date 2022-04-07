package directconnect

import (
	"fmt"
)

func GatewayAssociationCreateResourceID(directConnectGatewayID, associatedGatewayID string) string {
	return fmt.Sprintf("ga-%s%s", directConnectGatewayID, associatedGatewayID)
}
