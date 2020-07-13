package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
)

// LocalGatewayRouteTableVpcAssociationState fetches the LocalGatewayRouteTableVpcAssociation and its State
func LocalGatewayRouteTableVpcAssociationState(conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{
			LocalGatewayRouteTableVpcAssociationIds: aws.StringSlice([]string{localGatewayRouteTableVpcAssociationID}),
		}

		output, err := conn.DescribeLocalGatewayRouteTableVpcAssociations(input)

		if err != nil {
			return nil, "", err
		}

		var association *ec2.LocalGatewayRouteTableVpcAssociation

		for _, outputAssociation := range output.LocalGatewayRouteTableVpcAssociations {
			if outputAssociation == nil {
				continue
			}

			if aws.StringValue(outputAssociation.LocalGatewayRouteTableVpcAssociationId) == localGatewayRouteTableVpcAssociationID {
				association = outputAssociation
				break
			}
		}

		if association == nil {
			return association, ec2.RouteTableAssociationStateCodeDisassociated, nil
		}

		return association, aws.StringValue(association.State), nil
	}
}

const (
	ClientVpnEndpointStatusNotFound = "NotFound"

	ClientVpnEndpointStatusUnknown = "Unknown"
)

// ClientVpnEndpointStatus fetches the Client VPN endpoint and its Status
func ClientVpnEndpointStatus(conn *ec2.EC2, endpointID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeClientVpnEndpoints(&ec2.DescribeClientVpnEndpointsInput{
			ClientVpnEndpointIds: aws.StringSlice([]string{endpointID}),
		})
		if tfec2.ErrCodeEquals(err, tfec2.ErrCodeClientVpnEndpointIdNotFound) {
			return nil, ClientVpnEndpointStatusNotFound, nil
		}
		if err != nil {
			return nil, ClientVpnEndpointStatusUnknown, err
		}

		if result == nil || len(result.ClientVpnEndpoints) == 0 || result.ClientVpnEndpoints[0] == nil {
			return nil, ClientVpnEndpointStatusNotFound, nil
		}

		endpoint := result.ClientVpnEndpoints[0]
		if endpoint.Status == nil || endpoint.Status.Code == nil {
			return endpoint, ClientVpnEndpointStatusUnknown, nil
		}

		return endpoint, aws.StringValue(endpoint.Status.Code), nil
	}
}

const (
	ClientVpnAuthorizationRuleStatusNotFound = "NotFound"

	ClientVpnAuthorizationRuleStatusUnknown = "Unknown"
)

// ClientVpnAuthorizationRuleStatus fetches the Client VPN authorization rule and its Status
func ClientVpnAuthorizationRuleStatus(conn *ec2.EC2, authorizationRuleID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		endpointID, targetNetworkCidr, accessGroupID, err := tfec2.ClientVpnAuthorizationRuleParseID(authorizationRuleID)
		if err != nil {
			return nil, ClientVpnAuthorizationRuleStatusUnknown, err
		}

		result, err := finder.ClientVpnAuthorizationRule(conn, endpointID, targetNetworkCidr, accessGroupID)
		if tfec2.ErrCodeEquals(err, tfec2.ErrCodeClientVpnAuthorizationRuleNotFound) {
			return nil, ClientVpnAuthorizationRuleStatusNotFound, nil
		}
		if err != nil {
			return nil, ClientVpnAuthorizationRuleStatusUnknown, err
		}

		if result == nil || len(result.AuthorizationRules) == 0 || result.AuthorizationRules[0] == nil {
			return nil, ClientVpnAuthorizationRuleStatusNotFound, nil
		}

		if len(result.AuthorizationRules) > 1 {
			return nil, ClientVpnAuthorizationRuleStatusUnknown, fmt.Errorf("internal error: found %d results for Client VPN authorization rule (%s) status, need 1", len(result.AuthorizationRules), authorizationRuleID)
		}

		rule := result.AuthorizationRules[0]
		if rule.Status == nil || rule.Status.Code == nil {
			return rule, ClientVpnAuthorizationRuleStatusUnknown, nil
		}

		return rule, aws.StringValue(rule.Status.Code), nil
	}
}

const (
	ClientVpnRouteStatusNotFound = "NotFound"

	ClientVpnRouteStatusUnknown = "Unknown"
)

// ClientVpnRouteStatus fetches the Client VPN route and its Status
func ClientVpnRouteStatus(conn *ec2.EC2, routeID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		endpointID, targetSubnetID, destinationCidr, err := tfec2.ClientVpnRouteParseID(routeID)
		if err != nil {
			return nil, ClientVpnRouteStatusNotFound, err
		}

		result, err := finder.ClientVpnRoute(conn, endpointID, targetSubnetID, destinationCidr)
		if tfec2.ErrCodeEquals(err, tfec2.ErrCodeClientVpnRouteNotFound) {
			return nil, ClientVpnRouteStatusNotFound, nil
		}
		if err != nil {
			return nil, ClientVpnRouteStatusUnknown, err
		}

		if result == nil || len(result.Routes) == 0 || result.Routes[0] == nil {
			return nil, ClientVpnRouteStatusNotFound, nil
		}

		if len(result.Routes) > 1 {
			return nil, ClientVpnRouteStatusUnknown, fmt.Errorf("internal error: found %d results for Client VPN route (%s) status, need 1", len(result.Routes), routeID)
		}

		rule := result.Routes[0]
		if rule.Status == nil || rule.Status.Code == nil {
			return rule, ClientVpnRouteStatusUnknown, nil
		}

		return rule, aws.StringValue(rule.Status.Code), nil
	}
}
