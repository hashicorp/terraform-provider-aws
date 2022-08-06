package ec2

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/create"
)

const managedPrefixListEntryIDSeparator = ","

func ManagedPrefixListEntryCreateID(prefixListID, cidrBlock string) string {
	parts := []string{prefixListID, cidrBlock}
	id := strings.Join(parts, managedPrefixListEntryIDSeparator)
	return id
}

func ManagedPrefixListEntryParseID(id string) (string, string, error) {
	parts := strings.Split(id, managedPrefixListEntryIDSeparator)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "",
		fmt.Errorf("unexpected format for ID (%q), expected prefix-list-id"+managedPrefixListEntryIDSeparator+"cidr-block", id)
}

// RouteCreateID returns a route resource ID.
func RouteCreateID(routeTableID, destination string) string {
	return fmt.Sprintf("r-%s%d", routeTableID, create.StringHashcode(destination))
}

func VPCEndpointRouteTableAssociationCreateID(vpcEndpointID, routeTableID string) string {
	return fmt.Sprintf("a-%s%d", vpcEndpointID, create.StringHashcode(routeTableID))
}

func VPCEndpointSecurityGroupAssociationCreateID(vpcEndpointID, securityGroupID string) string {
	return fmt.Sprintf("a-%s%d", vpcEndpointID, create.StringHashcode(securityGroupID))
}

func VPCEndpointSubnetAssociationCreateID(vpcEndpointID, subnetID string) string {
	return fmt.Sprintf("a-%s%d", vpcEndpointID, create.StringHashcode(subnetID))
}

func VPNGatewayVPCAttachmentCreateID(vpnGatewayID, vpcID string) string {
	return fmt.Sprintf("vpn-attachment-%x", create.StringHashcode(fmt.Sprintf("%s-%s", vpcID, vpnGatewayID)))
}

const vpnGatewayRoutePropagationIDSeparator = "_"

func VPNGatewayRoutePropagationCreateID(routeTableID, gatewayID string) string {
	parts := []string{gatewayID, routeTableID}
	id := strings.Join(parts, vpnGatewayRoutePropagationIDSeparator)
	return id
}

func VPNGatewayRoutePropagationParseID(id string) (string, string, error) {
	parts := strings.Split(id, vpnGatewayRoutePropagationIDSeparator)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[1], parts[0], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected vpn-gateway-id%[2]sroute-table-id", id, vpnGatewayRoutePropagationIDSeparator)
}
