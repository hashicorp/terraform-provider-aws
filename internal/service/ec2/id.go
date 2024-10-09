// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/create"
)

// routeCreateID returns a route resource ID.
func routeCreateID(routeTableID, destination string) string {
	return fmt.Sprintf("r-%s%d", routeTableID, create.StringHashcode(destination))
}

func vpcEndpointRouteTableAssociationCreateID(vpcEndpointID, routeTableID string) string {
	return fmt.Sprintf("a-%s%d", vpcEndpointID, create.StringHashcode(routeTableID))
}

func vpcEndpointSecurityGroupAssociationCreateID(vpcEndpointID, securityGroupID string) string {
	return fmt.Sprintf("a-%s%d", vpcEndpointID, create.StringHashcode(securityGroupID))
}

func vpcEndpointSubnetAssociationCreateID(vpcEndpointID, subnetID string) string {
	return fmt.Sprintf("a-%s%d", vpcEndpointID, create.StringHashcode(subnetID))
}

func vpnGatewayVPCAttachmentCreateID(vpnGatewayID, vpcID string) string {
	return fmt.Sprintf("vpn-attachment-%x", create.StringHashcode(fmt.Sprintf("%s-%s", vpcID, vpnGatewayID)))
}

const vpnGatewayRoutePropagationIDSeparator = "_"

func vpnGatewayRoutePropagationCreateID(routeTableID, gatewayID string) string {
	parts := []string{gatewayID, routeTableID}
	id := strings.Join(parts, vpnGatewayRoutePropagationIDSeparator)
	return id
}

func vpnGatewayRoutePropagationParseID(id string) (string, string, error) {
	parts := strings.Split(id, vpnGatewayRoutePropagationIDSeparator)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[1], parts[0], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected vpn-gateway-id%[2]sroute-table-id", id, vpnGatewayRoutePropagationIDSeparator)
}
