package ec2

import (
	"fmt"
	"strings"
)

const clientVpnAuthorizationRuleIDSeparator = ","

func ClientVpnAuthorizationRuleCreateID(endpointID, targetNetworkCidr, accessGroupID string) string {
	parts := []string{endpointID, targetNetworkCidr}
	if accessGroupID != "" {
		parts = append(parts, accessGroupID)
	}
	id := strings.Join(parts, clientVpnAuthorizationRuleIDSeparator)
	return id
}

func ClientVpnAuthorizationRuleParseID(id string) (string, string, string, error) {
	parts := strings.Split(id, clientVpnAuthorizationRuleIDSeparator)
	if len(parts) == 2 {
		return parts[0], parts[1], "", nil
	}
	if len(parts) == 3 {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "",
		fmt.Errorf("unexpected format for ID (%q), expected endpoint-id"+clientVpnAuthorizationRuleIDSeparator+
			"target-network-cidr or endpoint-id"+clientVpnAuthorizationRuleIDSeparator+"target-network-cidr"+
			clientVpnAuthorizationRuleIDSeparator+"group-id", id)
}
